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
	Version                 string
	Mode                    string
	Token                   string
	TokenFile               string
	BsId                    string
	BsIdFile                string
	Target                  string
	WorkingDir              string
	TargetArgs              string
	ProfilerLogDir          string
	ProfilerLogLevel        string
	CustomAgentUrl          string
	LabId                   string
	Proxy                   string
	ProxyUsername           string
	ProxyPassword           string
	IgnoreCertificateErrors string
	Tools                   string
	Tags                    string
	NotCli                  string
	AppName                 string
	BranchName              string
	BuildName               string
	IncludeNamespace        string
	WorkspacePath           string
	IgnoreGeneratedCode     string
	TestStage               string
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

			options := &SealightsOptions{
				Version:                 queryString("version"),
				Mode:                    queryString("mode"), // default value is 'testListener'
				Token:                   queryString("token"),
				TokenFile:               queryString("tokenFile"),
				BsId:                    queryString("bsId"),
				BsIdFile:                queryString("bsIdFile"),
				Target:                  queryString("target"),
				WorkingDir:              queryString("workingDir"),
				TargetArgs:              queryString("targetArgs"),
				ProfilerLogDir:          queryString("profilerLogDir"),
				ProfilerLogLevel:        queryString("profilerLogLevel"),
				CustomAgentUrl:          queryString("customAgentUrl"),
				LabId:                   queryString("labId"),
				Proxy:                   queryString("proxy"),
				ProxyUsername:           queryString("proxyUsername"),
				ProxyPassword:           queryString("proxyPassword"),
				IgnoreCertificateErrors: queryString("ignoreCertificateErrors"),
				Tools:                   queryString("tools"),
				Tags:                    queryString("tags"),
				NotCli:                  queryString("notCli"),
				AppName:                 queryString("appName"),
				BranchName:              queryString("branchName"),
				BuildName:               queryString("buildName"),
				IncludeNamespace:        queryString("includeNamespace"),
				WorkspacePath:           queryString("workspacePath"),
				IgnoreGeneratedCode:     queryString("ignoreGeneratedCode"),
				TestStage:               queryString("testStage"),
			}

			isTokenProvided := options.Token != "" || options.TokenFile != ""
			if !isTokenProvided {
				conf.Log.Warning("Sealights access token isn't provided")
				return
			}

			isSessionIdProvided := options.BsId != "" || options.BsIdFile != ""
			if !isSessionIdProvided {
				conf.Log.Warning("Sealights build session id isn't provided")
				return
			}

			conf.Value = options
			return
		}
	}
}
