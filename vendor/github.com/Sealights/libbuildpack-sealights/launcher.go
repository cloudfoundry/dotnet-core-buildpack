package sealights

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

const AgentName = "SL.DotNet.dll"
const SealightsCli = "sealights"
const DefaultAgentMode = "help"
const ProfilerId = "01CA2C22-DC03-4FF5-8350-59E32A3536BA"

type Launcher struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	AgentDirAbsolute   string
	AgentDirForRuntime string
	DotNetDir          string
}

func NewLauncher(log *libbuildpack.Logger, options *SealightsOptions, agentInstallationDir string, dotnetInstallationDir string, buildDir string) *Launcher {
	agentDirForRuntime := filepath.Join("${HOME}", agentInstallationDir)
	agentDirAbsolute := filepath.Join(buildDir, agentInstallationDir)
	return &Launcher{Log: log, Options: options, AgentDirForRuntime: agentDirForRuntime, AgentDirAbsolute: agentDirAbsolute, DotNetDir: dotnetInstallationDir}
}

func (la *Launcher) ModifyStartParameters(stager *libbuildpack.Stager) error {
	la.updateAgentPath(stager)

	releaseInfo := NewReleaseInfo(stager.BuildDir())

	startCommand := releaseInfo.GetStartCommand()
	newStartCommand := la.updateStartCommand(startCommand)

	if la.Options.Verb != "" {
		// update application launch command only if Verb is provided
		err := releaseInfo.SetStartCommand(newStartCommand)
		if err != nil {
			return err
		}
	}

	la.Log.Info(fmt.Sprintf("Sealights: Start command updated. From '%s' to '%s'", startCommand, newStartCommand))

	return nil
}

func (la *Launcher) updateAgentPath(stager *libbuildpack.Stager) {
	if strings.HasPrefix(la.AgentDirForRuntime, stager.BuildDir()) {
		clearPath := strings.TrimPrefix(la.AgentDirForRuntime, stager.BuildDir())
		la.AgentDirForRuntime = filepath.Join(".", clearPath)
	}
}

func (la *Launcher) updateStartCommand(originalCommand string) string {
	// expected command format:
	// cd ${DEPS_DIR}/0/dotnet_publish && exec ./app --server.urls http://0.0.0.0:${PORT}
	// cd ${DEPS_DIR}/0/dotnet_publish && exec dotnet ./app.dll --server.urls http://0.0.0.0:${PORT}

	parts := strings.SplitAfterN(originalCommand, "exec ", 2)

	newCmd := parts[0] + la.buildCommandLine(parts[1])

	return newCmd
}

//dotnet SL.DotNet.dll testListener --logAppendFile true --logFilename /tmp/collector.log --tokenFile /tmp/sltoken.txt --buildSessionIdFile /tmp/buildsessionid.txt --target dotnet --workingDir /tmp/app --profilerLogDir /tmp/ --profilerLogLevel 7 --targetArgs \"test app.dll\"
func (la *Launcher) buildCommandLine(command string) string {
	if la.Options.CustomCommand != "" {
		return la.Options.CustomCommand
	}

	var sb strings.Builder
	options := la.Options

	agent := filepath.Join(la.AgentDirForRuntime, AgentName)
	dotnetCli := "dotnet"
	if la.DotNetDir != "" {
		dotnetCli = filepath.Join(la.DotNetDir, "dotnet")
	}

	agentMode := DefaultAgentMode
	if options.Verb != "" {
		agentMode = options.Verb
	}

	sb.WriteString(fmt.Sprintf("%s %s %s", dotnetCli, agent, agentMode))

	for key, value := range la.Options.SlArguments {
		la.Log.Info(fmt.Sprintf("Added: --%s %s", key, value))

		sb.WriteString(fmt.Sprintf(" --%s %s", key, value))
	}

	if la.Options.ParseArgsFromCmd == "true" {
		_, exists := la.Options.SlArguments["workingDir"]
		if !exists {
			sb.WriteString(" --workingDir ${PWD}")
		}

		parsedTarget, parsedArgs := la.getTargetArgs(command)
		_, exists = la.Options.SlArguments["target"]
		if !exists {
			sb.WriteString(fmt.Sprintf(" --target %s", parsedTarget))
		}

		_, exists = la.Options.SlArguments["targetArgs"]
		if !exists {
			sb.WriteString(fmt.Sprintf(" --targetArgs \"%s\"", parsedArgs))
		}
	}

	testListenerSessionKey, sessionKeyExists := la.Options.SlArguments["testListenerSessionKey"]
	if sessionKeyExists {
		exportEnvCmd, err := la.addProfilerConfiguration(la.AgentDirForRuntime, testListenerSessionKey)
		if err != nil {
			la.Log.Error("Sealights. Failed to parse arguments")
			return command
		}

		sb.WriteString(fmt.Sprintf(" && %s && %s", exportEnvCmd, command))

		la.addSealightsEntryPoint(dotnetCli, agent)
	}

	return sb.String()
}

func (la *Launcher) getTargetArgs(command string) (target string, args string) {
	if strings.HasPrefix(command, "dotnet") || la.DotNetDir == "" {
		// use dotnet as target and remove it from command
		target = "dotnet"
		command = strings.TrimPrefix(command, "dotnet")
		command = strings.TrimPrefix(command, " ")
	} else {
		// use dotnet from sealights folder
		target = filepath.Join(la.DotNetDir, "dotnet")
	}

	parts := strings.SplitN(command, " ", 2)
	withoutArguments := parts[0]
	args = fmt.Sprintf("test %s", withoutArguments)

	if strings.HasPrefix(args, "--") {
		args = fmt.Sprintf(" %s", args)
	}

	return
}

func (la *Launcher) addProfilerConfiguration(agentPath string, collectorId string) (string, error) {
	agentEnvFileName := "sealights.envrc"
	exportCommand := "export"
	executeCommand := "source"

	if runtime.GOOS == "windows" {
		agentEnvFileName = "sealights.bat"
		exportCommand = "set"
		executeCommand = "call"
	}

	la.Log.Debug(fmt.Sprintf("Create file %s", agentEnvFileName))

	agentEnvFile := filepath.Join(la.AgentDirAbsolute, agentEnvFileName)
	homeBasedEnvFile := filepath.Join(la.AgentDirForRuntime, agentEnvFileName)
	file, err := os.OpenFile(agentEnvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		la.Log.Error(fmt.Sprint(err))
		return "", err
	}
	defer file.Close()

	agentProfilerLibx86 := filepath.Join(la.AgentDirForRuntime, "SL.DotNet.ProfilerLib_x86.dll")
	agentProfilerLibx64 := filepath.Join(la.AgentDirForRuntime, "SL.DotNet.ProfilerLib_x64.dll")

	fileContent := ""

	fileContent += fmt.Sprintf("%s Cor_Profiler={%s}\n", exportCommand, ProfilerId)
	fileContent += fmt.Sprintf("%s Cor_Enable_Profiling=1\n", exportCommand)
	fileContent += fmt.Sprintf("%s Cor_Profiler_Path=%s\n", exportCommand, agentProfilerLibx64)
	fileContent += fmt.Sprintf("%s COR_PROFILER_PATH_32=%s\n", exportCommand, agentProfilerLibx86)
	fileContent += fmt.Sprintf("%s COR_PROFILER_PATH_64=%s\n", exportCommand, agentProfilerLibx64)
	fileContent += fmt.Sprintf("%s SeaLights_CollectorId=%s\n", exportCommand, collectorId)

	if _, err = file.WriteString(fileContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s && env", executeCommand, homeBasedEnvFile), nil
}

func (la *Launcher) addSealightsEntryPoint(dotnetCli string, agent string) error {
	la.Log.Debug(fmt.Sprintf("Create file [%s] for cli", SealightsCli))

	cliFileName := filepath.Join(la.AgentDirAbsolute, SealightsCli)
	file, err := os.OpenFile(cliFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer file.Close()

	runCmd := fmt.Sprintf(`exec %s %s "$$@"`, dotnetCli, agent)

	file.WriteString(`#!/bin/sh` + "\n\n" + runCmd)

	return nil
}
