package sealights

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/libbuildpack"
)

const WindowsProfilerId = "{01CA2C22-DC03-4FF5-8350-59E32A3536BA}"
const WingowsProfilerName32 = "SL.DotNet.ProfilerLib_x86.dll"
const WingowsProfilerName64 = "SL.DotNet.ProfilerLib_x64.dll"

const LinuxProfilerId = "{3B1DAA64-89D4-4999-ABF4-6A979B650B7D}"
const LinuxProfilerName = "libSL.DotNet.ProfilerLib.Linux.so"

const DefaultPort = "31031"

type PlatformProfilerParams struct {
	Name_32 string
	Name_64 string
	Id      string
}

type EnvManager struct {
	Options *SealightsOptions
	Log     *libbuildpack.Logger
}

func NewEnvManager(log *libbuildpack.Logger, options *SealightsOptions) *EnvManager {
	envManager := EnvManager{Log: log, Options: options}

	return &envManager
}

func (emng *EnvManager) GetVariables(runtimeDirectory string) map[string]string {
	profilerInfo := emng.getProfilerInfo()

	agentProfilerLibx86 := filepath.Join(runtimeDirectory, profilerInfo.Name_32)
	agentProfilerLibx64 := filepath.Join(runtimeDirectory, profilerInfo.Name_64)

	env_variables := map[string]string{}

	env_variables["Cor_Profiler"] = profilerInfo.Id
	env_variables["Cor_Enable_Profiling"] = "1"
	env_variables["Cor_Profiler_Path"] = agentProfilerLibx64
	env_variables["COR_PROFILER_PATH_32"] = agentProfilerLibx86
	env_variables["COR_PROFILER_PATH_64"] = agentProfilerLibx64
	env_variables["CORECLR_ENABLE_PROFILING"] = "1"
	env_variables["CORECLR_PROFILER"] = profilerInfo.Id
	env_variables["CORECLR_PROFILER_PATH_32"] = agentProfilerLibx86
	env_variables["CORECLR_PROFILER_PATH_64"] = agentProfilerLibx64
	env_variables["SL_AGENT_PORT"] = DefaultPort

	testListenerSessionKey, sessionKeyExists := emng.Options.SlArguments["testListenerSessionKey"]
	if sessionKeyExists {
		env_variables["SL_CollectorId"] = testListenerSessionKey
	}

	if emng.Options.UsePic {
		env_variables["SL_PROFILER_INITIALIZECOLLECTOR"] = "1"
		env_variables["SL_PROFILER_BLOCKING_CONNECTION_STARTUP"] = "ASYNC"
	}

	// put to the dictionary provided variables and
	// replace auto generated variables with provided
	for key, value := range emng.Options.SlEnvironment {
		env_variables[key] = value
	}

	return env_variables
}

func (emng *EnvManager) WriteIntoFile(filePath string, envVariables map[string]string) error {
	exportCommand := "export"
	if runtime.GOOS == "windows" {
		exportCommand = "set"
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		emng.Log.Error(fmt.Sprint(err))
		return err
	}

	defer file.Close()

	fileContent := ""

	for key, value := range envVariables {
		fileContent += fmt.Sprintf("%s %s=%s\n", exportCommand, key, value)
	}

	if _, err = file.WriteString(fileContent); err != nil {
		return err
	}

	return nil
}

func (emng *EnvManager) getProfilerInfo() *PlatformProfilerParams {
	if runtime.GOOS == "windows" {
		profilerParams := PlatformProfilerParams{
			Name_32: WingowsProfilerName32,
			Name_64: WingowsProfilerName64,
			Id:      WindowsProfilerId,
		}

		return &profilerParams
	} else {
		profilerParams := PlatformProfilerParams{
			Name_32: LinuxProfilerName,
			Name_64: LinuxProfilerName,
			Id:      LinuxProfilerId,
		}

		return &profilerParams
	}
}
