package sealights

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

const WindowsProfilerId = "01CA2C22-DC03-4FF5-8350-59E32A3536BA"
const WindowsAgentName = "SL.DotNet.exe"

const LinuxProfilerId = "3B1DAA64-89D4-4999-ABF4-6A979B650B7D"
const LinuxAgentName = "SL.DotNet"

type Launcher struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	AgentDirAbsolute   string
	AgentDirForRuntime string
}

func NewLauncher(log *libbuildpack.Logger, options *SealightsOptions, agentInstallationDir string, buildDir string) *Launcher {
	agentDirForRuntime := filepath.Join("${HOME}", agentInstallationDir)
	agentDirAbsolute := filepath.Join(buildDir, agentInstallationDir)
	return &Launcher{Log: log, Options: options, AgentDirForRuntime: agentDirForRuntime, AgentDirAbsolute: agentDirAbsolute}
}

func (la *Launcher) ModifyStartParameters(stager *libbuildpack.Stager) error {
	la.updateAgentPath(stager)

	releaseInfo := NewReleaseInfo(stager.BuildDir())

	startCommand := releaseInfo.GetStartCommand()
	newStartCommand := la.updateStartCommand(startCommand)

	shouldApply := la.Options.Verb != "" || la.Options.CustomCommand != ""
	if shouldApply {
		err := releaseInfo.SetStartCommand(newStartCommand)
		if err != nil {
			return err
		}

		la.Log.Info(fmt.Sprintf("Sealights: Start command updated. From '%s' to '%s'", startCommand, newStartCommand))
	} else {
		la.Log.Warning("Sealights. Verb or Custom Command are missed - start command will not be modified")
	}

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

	parts := strings.SplitAfterN(originalCommand, "&& ", 2)

	newCmd := parts[0] + la.buildCommandLine(parts[1])

	return newCmd
}

// Get command line that will launch sealights agent with required options.
// Examples:
// SL.DotNet [verb] [options]
// SL.DotNet [verb] [options] && source sealights.envrc && [start target app]
// [customCommand]
func (la *Launcher) buildCommandLine(command string) string {
	if la.Options.CustomCommand != "" {
		return la.Options.CustomCommand
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s", la.agentFillPath(), la.Options.Verb))

	for key, value := range la.Options.SlArguments {
		sb.WriteString(fmt.Sprintf(" --%s %s", key, value))
	}

	testListenerSessionKey, sessionKeyExists := la.Options.SlArguments["testListenerSessionKey"]
	if sessionKeyExists {
		exportEnvCmd, _ := la.addProfilerConfiguration(la.AgentDirForRuntime, testListenerSessionKey)

		// if testListenerSessionKey is provided, selected mode is background test listener
		// and target application should be started after the sealights agent
		sb.WriteString(fmt.Sprintf(" && %s && env && %s", exportEnvCmd, command))

		// resulting launch command should have only one 'exec' keyword
		// for the last subsequence part
		return sb.String()
	} else {
		return "exec " + sb.String()
	}
}

// Create file sealights.envrc with all the required env variables to make
// the profiler to attach to the target application
func (la *Launcher) addProfilerConfiguration(agentPath string, collectorId string) (string, error) {
	agentEnvFileName := "sealights.envrc"
	exportCommand := "export"
	executeCommand := "source"
	profilerLib_x64 := "libSL.DotNet.ProfilerLib.Linux.so"
	profilerLib_x86 := "libSL.DotNet.ProfilerLib.Linux.so"
	profilerId := LinuxProfilerId

	if runtime.GOOS == "windows" {
		agentEnvFileName = "sealights.bat"
		exportCommand = "set"
		executeCommand = "call"
		profilerLib_x64 = "SL.DotNet.ProfilerLib_x64.dll"
		profilerLib_x86 = "SL.DotNet.ProfilerLib_x86.dll"
		profilerId = WindowsProfilerId
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

	agentProfilerLibx86 := filepath.Join(la.AgentDirForRuntime, profilerLib_x86)
	agentProfilerLibx64 := filepath.Join(la.AgentDirForRuntime, profilerLib_x64)

	fileContent := ""

	fileContent += fmt.Sprintf("%s Cor_Profiler={%s}\n", exportCommand, profilerId)
	fileContent += fmt.Sprintf("%s Cor_Enable_Profiling=1\n", exportCommand)
	fileContent += fmt.Sprintf("%s Cor_Profiler_Path=%s\n", exportCommand, agentProfilerLibx64)
	fileContent += fmt.Sprintf("%s COR_PROFILER_PATH_32=%s\n", exportCommand, agentProfilerLibx86)
	fileContent += fmt.Sprintf("%s COR_PROFILER_PATH_64=%s\n", exportCommand, agentProfilerLibx64)
	fileContent += fmt.Sprintf("%s CORECLR_ENABLE_PROFILING=1\n", exportCommand)
	fileContent += fmt.Sprintf("%s CORECLR_PROFILER={%s}\n", exportCommand, profilerId)
	fileContent += fmt.Sprintf("%s CORECLR_PROFILER_PATH_32=%s\n", exportCommand, agentProfilerLibx86)
	fileContent += fmt.Sprintf("%s CORECLR_PROFILER_PATH_64=%s\n", exportCommand, agentProfilerLibx64)
	fileContent += fmt.Sprintf("%s SeaLights_CollectorId=%s\n", exportCommand, collectorId)

	if la.Options.EnableProfilerLogs == "true" {
		fileContent += fmt.Sprintf("%s SL_LogDir=%s\n", exportCommand, la.AgentDirForRuntime)
		fileContent += fmt.Sprintf("%s SL_LogLevel=6\n", exportCommand)
	}

	if _, err = file.WriteString(fileContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", executeCommand, homeBasedEnvFile), nil
}

func (la *Launcher) agentFillPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(la.AgentDirForRuntime, WindowsAgentName)
	} else {
		return filepath.Join(la.AgentDirForRuntime, LinuxAgentName)
	}
}
