package sealights

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

const WindowsAgentName = "SL.DotNet.exe"
const LinuxAgentName = "SL.DotNet"
const GlobalVariablesFile = "sealights-env.sh"

type Launcher struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	AgentDirAbsolute   string
	AgentDirForRuntime string
	Stager             *libbuildpack.Stager
}

func NewLauncher(log *libbuildpack.Logger, options *SealightsOptions, agentInstallationDir string, stager *libbuildpack.Stager) *Launcher {
	agentDirForRuntime := filepath.Join("${HOME}", agentInstallationDir)
	agentDirAbsolute := filepath.Join(stager.BuildDir(), agentInstallationDir)
	return &Launcher{Log: log, Options: options, AgentDirForRuntime: agentDirForRuntime, AgentDirAbsolute: agentDirAbsolute, Stager: stager}
}

func (la *Launcher) ModifyStartParameters(stager *libbuildpack.Stager) error {
	la.updateAgentPath(stager)

	releaseInfo := NewReleaseInfo(stager.BuildDir())

	startCommand := releaseInfo.GetStartCommand()
	newStartCommand := la.updateStartCommand(startCommand)

	la.setEnvVariablesGlobally()

	shouldApply := la.Options.Verb != "" || la.Options.CustomCommand != ""
	if shouldApply {
		err := releaseInfo.SetStartCommand(newStartCommand)
		if err != nil {
			return err
		}

		logMessage := fmt.Sprintf("Sealights: Start command updated. From '%s' to '%s'", startCommand, newStartCommand)
		la.Log.Info(maskSensitiveData(logMessage))
	} else {
		la.Log.Debug("Sealights. Start command will not be modified")
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

	agentExecutable := la.agentFullPath()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s", agentExecutable, la.Options.Verb))

	for key, value := range la.Options.SlArguments {
		sb.WriteString(fmt.Sprintf(" --%s %s", key, value))
	}

	// background test listener require to set environment variables
	// before starting the target process
	if la.Options.Verb == "startBackgroundTestListener" {
		exportEnvCmd, _ := la.addProfilerConfiguration(la.AgentDirForRuntime)

		// if testListenerSessionKey is provided, selected mode is background test listener
		// and target application should be started after the sealights agent
		sb.WriteString(fmt.Sprintf(" && %s && %s", exportEnvCmd, command))

		// resulting launch command should have only one 'exec' keyword
		// for the last subsequence part
		return sb.String()
	} else {
		return "exec " + sb.String()
	}
}

// Create file sealights.envrc with all the required env variables to make
// the profiler to attach to the target application
func (la *Launcher) addProfilerConfiguration(agentPath string) (string, error) {
	executeCommand := "source"
	if runtime.GOOS == "windows" {
		executeCommand = "call"
	}

	agentEnvFileName := la.agentEnvFileName()

	agentEnvFile := filepath.Join(la.AgentDirAbsolute, agentEnvFileName)
	homeBasedEnvFile := filepath.Join(la.AgentDirForRuntime, agentEnvFileName)

	envManager := NewEnvManager(la.Log, la.Options)
	envVariebles := envManager.GetVariables(la.AgentDirForRuntime)

	err := envManager.WriteIntoFile(agentEnvFile, envVariebles)
	if err != nil {
		return "", err
	}

	la.Log.Debug(fmt.Sprintf("Create file %s", agentEnvFileName))

	return fmt.Sprintf("%s %s", executeCommand, homeBasedEnvFile), nil
}

func (la *Launcher) agentFullPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(la.AgentDirForRuntime, WindowsAgentName)
	} else {
		return filepath.Join(la.AgentDirForRuntime, LinuxAgentName)
	}
}

func (la *Launcher) agentEnvFileName() string {
	if runtime.GOOS == "windows" {
		return "sealights.bat"
	} else {
		return "sealights.envrc"
	}
}

func (la *Launcher) setEnvVariablesGlobally() {
	envManager := NewEnvManager(la.Log, la.Options)
	var envVariables map[string]string
	if la.Options.UsePic {
		// set all variables important for the profiler
		envVariables = envManager.GetVariables(la.AgentDirForRuntime)
	} else {
		// set only dlls provided directly in options
		envVariables = la.Options.SlArguments
	}

	if runtime.GOOS == "windows" {
		for key, value := range envVariables {
			os.Setenv(key, value)
		}
	} else {
		localEnvFile := filepath.Join(la.AgentDirAbsolute, GlobalVariablesFile)
		err := envManager.WriteIntoFile(localEnvFile, envVariables)
		if err != nil {
			la.Log.Error("Sealights. Failed to create local env file")
		}

		sealightsEnvPath := filepath.Join(la.Stager.DepDir(), "profile.d", GlobalVariablesFile)
		la.Log.Debug("Copy %s to %s", localEnvFile, sealightsEnvPath)
		if err = libbuildpack.CopyFile(localEnvFile, sealightsEnvPath); err != nil {
			la.Log.Error("Sealights. Failed to copy file to profile.d")
		}
	}
}

func maskSensitiveData(input string) string {
	re := regexp.MustCompile(`(--proxyPassword\s|--token\s).*?(\s--|\s&&)`)
	output := re.ReplaceAllString(input, "$1********$2")

	return output
}
