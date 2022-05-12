package sealights

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

const AgentName = "SL.DotNet.dll"
const DefaultAgentMode = "testListener"

type Launcher struct {
	Log       *libbuildpack.Logger
	Options   *SealightsOptions
	AgentDir  string
	DotNetDir string
}

func NewLauncher(log *libbuildpack.Logger, options *SealightsOptions, agentInstallationDir string, dotnetInstallationDir string) *Launcher {
	return &Launcher{Log: log, Options: options, AgentDir: agentInstallationDir, DotNetDir: dotnetInstallationDir}
}

func (la *Launcher) ModifyStartParameters(stager *libbuildpack.Stager) error {
	la.updateAgentPath(stager)

	releaseInfo := NewReleaseInfo(stager.BuildDir())

	startCommand := releaseInfo.GetStartCommand()
	newStartCommand := la.updateStartCommand(startCommand)
	err := releaseInfo.SetStartCommand(newStartCommand)
	if err != nil {
		return err
	}

	la.Log.Info(fmt.Sprintf("Sealights: Start command updated. From '%s' to '%s'", startCommand, newStartCommand))

	return nil
}

func (la *Launcher) updateAgentPath(stager *libbuildpack.Stager) {
	if strings.HasPrefix(la.AgentDir, stager.BuildDir()) {
		clearPath := strings.TrimPrefix(la.AgentDir, stager.BuildDir())
		la.AgentDir = filepath.Join(".", clearPath)
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

	var sb strings.Builder
	options := la.Options

	agent := filepath.Join(la.AgentDir, AgentName)
	dotnetCli := "dotnet"
	if la.DotNetDir != "" {
		dotnetCli = filepath.Join(la.DotNetDir, "dotnet")
	}

	agentMode := DefaultAgentMode
	if options.Mode != "" {
		agentMode = options.Mode
	}

	sb.WriteString(fmt.Sprintf("%s %s %s", dotnetCli, agent, agentMode))

	if options.TokenFile != "" {
		sb.WriteString(fmt.Sprintf(" --tokenfile %s", options.TokenFile))
	} else {
		sb.WriteString(fmt.Sprintf(" --token %s", options.Token))
	}

	if options.BsIdFile != "" {
		sb.WriteString(fmt.Sprintf(" --buildSessionIdFile %s", options.BsIdFile))
	} else {
		sb.WriteString(fmt.Sprintf(" --buildSessionId %s", options.BsId))
	}

	if options.ProfilerLogDir != "" {
		sb.WriteString(fmt.Sprintf(" --profilerLogDir %s", options.ProfilerLogDir))
	}

	if options.ProfilerLogLevel != "" {
		sb.WriteString(fmt.Sprintf(" --profilerLogLevel %s", options.ProfilerLogLevel))
	}

	if options.Tags != "" {
		sb.WriteString(fmt.Sprintf(" --tags %s", options.Tags))
	}

	if options.Tools != "" {
		sb.WriteString(fmt.Sprintf(" --tools %s", options.Tools))
	}

	if options.IgnoreCertificateErrors == "true" {
		sb.WriteString(" --ignoreCertificateErrors true")
	}

	if options.NotCli == "true" {
		sb.WriteString(" --notCli true")
	}

	if options.AppName != "" {
		sb.WriteString(fmt.Sprintf(" --appName %s", options.AppName))
	}

	if options.BranchName != "" {
		sb.WriteString(fmt.Sprintf(" --branchName %s", options.BranchName))
	}

	if options.BuildName != "" {
		sb.WriteString(fmt.Sprintf(" --buildName %s", options.BuildName))
	}

	if options.IncludeNamespace != "" {
		sb.WriteString(fmt.Sprintf(" --includeNamespace %s", options.IncludeNamespace))
	}

	if options.WorkspacePath != "" {
		sb.WriteString(fmt.Sprintf(" --workspacePath %s", options.WorkspacePath))
	}

	if options.IgnoreGeneratedCode != "" {
		sb.WriteString(fmt.Sprintf(" --ignoreGeneratedCode %s", options.IgnoreGeneratedCode))
	}

	if options.TestStage != "" {
		sb.WriteString(fmt.Sprintf(" --testStage %s", options.TestStage))
	}

	if options.Proxy != "" {
		sb.WriteString(fmt.Sprintf(" --proxy %s", options.Proxy))
		sb.WriteString(fmt.Sprintf(" --proxyUsername %s", options.ProxyUsername))
		sb.WriteString(fmt.Sprintf(" --proxyPassword %s", options.ProxyPassword))
	}

	sb.WriteString(" --workingDir ${PWD}")

	if agentMode == DefaultAgentMode {
		target, args := la.getTargetArgs(command)
		sb.WriteString(fmt.Sprintf(" --target %s --targetArgs \"%s\"", target, args))
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

	if la.Options.Target != "" {
		target = la.Options.Target
	}

	if la.Options.TargetArgs != "" {
		args = la.Options.TargetArgs
	}

	if strings.HasPrefix(args, "--") {
		args = fmt.Sprintf(" %s", args)
	}

	return
}
