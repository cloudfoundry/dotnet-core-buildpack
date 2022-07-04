package sealights

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type VcapServicesModel struct {
	Sealights SealightsOptions
}

type SealightsOptions struct {
	Version          string
	Verb             string
	CustomAgentUrl   string
	CustomCommand    string
	LabId            string
	Token            string
	TokenFile        string
	BsId             string
	BsIdFile         string
	ParseArgsFromCmd string
	Proxy            string
	ProxyUsername    string
	ProxyPassword    string
	SlArguments      map[string]string
}

type Configuration struct {
	Value *SealightsOptions
	Log   *libbuildpack.Logger
}

func NewConfiguration(log *libbuildpack.Logger) *Configuration {
	configuration := Configuration{Log: log, Value: nil}
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
				Version:          queryString("version"),
				Verb:             queryString("verb"),
				CustomAgentUrl:   queryString("customAgentUrl"),
				CustomCommand:    queryString("customCommand"),
				Token:            queryString("token"),
				TokenFile:        queryString("tokenFile"),
				BsId:             queryString("buildSessionId"),
				BsIdFile:         queryString("buildSessionIdFile"),
				LabId:            queryString("labId"),
				ParseArgsFromCmd: queryString("parseArgsFromCmd"),
				Proxy:            queryString("proxy"),
				ProxyUsername:    queryString("proxyUsername"),
				ProxyPassword:    queryString("proxyPassword"),
				SlArguments:      slArguments,
			}

			isTokenProvided := options.Token != "" || options.TokenFile != ""
			if !isTokenProvided {
				conf.Log.Warning("Sealights access token isn't provided")
			}

			isSessionIdProvided := options.BsId != "" || options.BsIdFile != ""
			if !isSessionIdProvided {
				conf.Log.Warning("Sealights build session id isn't provided")
			}

			conf.Value = options
			return
		}
	}
}
