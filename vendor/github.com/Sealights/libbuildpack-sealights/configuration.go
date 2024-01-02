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
	SlArguments    map[string]string
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
	}

	for _, services := range vcapServices {
		for _, service := range services {
			if !strings.Contains(strings.ToLower(service.Name), "sealights") {
				continue
			}

			queryString := func(key string) string {
				if value, ok := service.Credentials[key].(string); ok {
					return value
				}
				return ""
			}

			slArguments := map[string]string{}
			for parameterName, parameterValue := range service.Credentials {
				_, shouldBeSkipped := buildpackSpecificArguments[parameterName]
				if shouldBeSkipped {
					continue
				}

				slArguments[parameterName] = parameterValue.(string)
			}

			options := &SealightsOptions{
				Version:        queryString("version"),
				Verb:           queryString("verb"),
				CustomAgentUrl: queryString("customAgentUrl"),
				CustomCommand:  queryString("customCommand"),
				Proxy:          queryString("proxy"),
				ProxyUsername:  queryString("proxyUsername"),
				ProxyPassword:  queryString("proxyPassword"),
				SlArguments:    slArguments,
			}

			// write warning in case token or session is not provided
			_, tokenProvided := options.SlArguments["token"]
			_, tokenFileProvided := options.SlArguments["tokenFile"]
			if !tokenProvided && !tokenFileProvided {
				conf.Log.Warning("Sealights access token isn't provided")
			}

			_, sessionProvided := options.SlArguments["buildSessionId"]
			_, sessionFileProvided := options.SlArguments["buildSessionIdFile"]
			if !sessionProvided && !sessionFileProvided {
				conf.Log.Warning("Sealights build session id isn't provided")
			}

			_, toolsProvided := options.SlArguments["tools"]
			if !toolsProvided {
				options.SlArguments["tools"] = conf.buildToolName()
			}

			if options.Verb == "" {
				options.Verb = "startBackgroundTestListener"
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

func (conf *Configuration) buildToolName() string {
	lang := conf.Stager.BuildpackLanguage()
	ver, err := conf.Stager.BuildpackVersion()
	if err != nil {
		conf.Log.Warning("Failed to get buildpack version")
		ver = "unknown"
	}

	return fmt.Sprintf("pcf-%s-%s", lang, ver)
}
