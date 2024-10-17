package sealights

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type VcapServicesModel struct {
	Sealights SealightsOptions
}

type SealightsOptions struct {
	Version        string
	Verb           string
	CustomAgentUrl string
	CustomCommand  string
	Proxy          string
	ProxyUsername  string
	ProxyPassword  string
	UsePic         bool
	SlArguments    map[string]string
	SlEnvironment  map[string]string
}

type Configuration struct {
	Value  *SealightsOptions
	Log    *libbuildpack.Logger
	Stager *libbuildpack.Stager
}

func NewConfiguration(log *libbuildpack.Logger, stager *libbuildpack.Stager) *Configuration {
	configuration := Configuration{Log: log, Value: nil, Stager: stager}
	configuration.parseVcapServices()

	return &configuration
}

func (conf Configuration) UseSealights() bool {
	return conf.Value != nil
}

func (conf *Configuration) parseVcapServices() {

	var vcapServices map[string][]struct {
		Name        string                 `json:"name"`
		Credentials map[string]interface{} `json:"credentials"`
	}

	if err := json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &vcapServices); err != nil {
		conf.Log.Debug("Failed to unmarshal VCAP_SERVICES: %s", err)
		return
	}

	buildpackSpecificArguments := map[string]bool{
		"version":        true,
		"verb":           true,
		"customAgentUrl": true,
		"customCommand":  true,
		"usePic":         true,
		"cli":            true,
		"env":            true,
	}

	for _, services := range vcapServices {
		for _, service := range services {
			if !strings.Contains(strings.ToLower(service.Name), "sealights") {
				continue
			}

			slEnvironment := getMap(service.Credentials, "env")
			if slEnvironment == nil {
				slEnvironment = make(map[string]string)
			}

			slArguments := getMap(service.Credentials, "cli")
			if slArguments == nil {
				slArguments = make(map[string]string)
			}

			// this validation required to make settings for version 1.5.0 back compatible with 1.4
			// there is no property "cli" in the old version of the libpack - all fields for cli comes directly from settings
			// so if env variables are set - all settings not from the new "cli" property will be used only by libpack itself
			if len(slEnvironment) == 0 {
				for parameterName, parameterValue := range service.Credentials {
					_, shouldBeSkipped := buildpackSpecificArguments[parameterName]
					if shouldBeSkipped {
						continue
					}

					slArguments[parameterName] = parameterValue.(string)
				}
			} else {
				conf.Log.Debug("Sealights. Option 'env' is provided - only options specified directly in the 'cli' field will be propagated to a command line")
			}

			options := &SealightsOptions{
				Version:        getValue[string](service.Credentials, "version"),
				Verb:           getValue[string](service.Credentials, "verb"),
				CustomAgentUrl: getValue[string](service.Credentials, "customAgentUrl"),
				CustomCommand:  getValue[string](service.Credentials, "customCommand"),
				Proxy:          getValue[string](service.Credentials, "proxy"),
				ProxyUsername:  getValue[string](service.Credentials, "proxyUsername"),
				ProxyPassword:  getValue[string](service.Credentials, "proxyPassword"),
				UsePic:         getValue[bool](service.Credentials, "usePic"),
				SlArguments:    slArguments,
				SlEnvironment:  slEnvironment,
			}

			// write warning in case token or session is not provided
			tokenVariables := []string{"token", "tokenFile", "SL_TOKEN", "SL_TOKENFILE"}
			isTokenProvided := conf.isAnyVariableProvided(tokenVariables, *options)
			if !isTokenProvided {
				conf.Log.Warning("The Sealights token has not been provided.")
			}

			_, picEnabled := options.SlEnvironment["SL_PROFILER_INITIALIZECOLLECTOR"]
			if picEnabled {
				options.UsePic = true
			}

			if options.UsePic {
				conf.Log.Info("Sealights. PIC mode enabled")
			}

			_, toolsProvided := options.SlArguments["tools"]
			if !toolsProvided {
				options.SlArguments["tools"] = conf.buildToolName()
			}

			_, tagsProvided := options.SlArguments["tags"]
			if !tagsProvided {
				options.SlArguments["tags"] = conf.buildToolName()
			}

			if options.Verb == "" && !options.UsePic {
				options.Verb = "startBackgroundTestListener"
				conf.Log.Debug("Sealights. Verb has not been set. Continue with 'startBackgroundTestListener'")
			}

			_, collectorIdPorvided := options.SlArguments["testListenerSessionKey"]
			if collectorIdPorvided {
				conf.Log.Warning("Sealights. Option 'testListenerSessionKey' isn't supported in this environment")
			}

			conf.Value = options
			return
		}
	}
}

func (conf *Configuration) isAnyVariableProvided(variableName []string, options SealightsOptions) bool {
	for _, key := range variableName {
		_, variableProvided := options.SlArguments[key]
		if variableProvided {
			return true
		}

		_, variableProvided = options.SlEnvironment[key]
		if variableProvided {
			return true
		}
	}

	return false
}

func (conf *Configuration) buildToolName() string {
	ver, err := conf.Stager.BuildpackVersion()
	if err != nil {
		conf.Log.Warning("Failed to get buildpack version")
		ver = "unknown"
	}

	return fmt.Sprintf("sl-pcf-%s", ver)
}

func getValue[T any](dict map[string]interface{}, key string) T {
	var result T

	if value, ok := dict[key].(T); ok {
		return value
	}

	return result
}

func getMap(dict map[string]interface{}, key string) map[string]string {
	var result = make(map[string]string)

	for key, value := range getValue[map[string]interface{}](dict, key) {
		result[key] = value.(string)
	}

	return result
}
