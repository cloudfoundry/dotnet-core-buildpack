package sealights

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

const WindowsPackageName = "sealights-dotnet-agent-windows-self-contained.zip"
const LinuxPackageName = "sealights-dotnet-agent-linux-self-contained.tar.gz"

const DefaultVersion = "latest"
const AgentDir = "sealights"
const DotnetDir = "dotnet-sdk"
const VersionFileName = "version.txt"

const AgentDownloadUrlFormat = "https://agents.sealights.co/dotnetcore/%s/%s"

type AgentInstaller struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	MaxDownloadRetries int
}

func NewAgentInstaller(log *libbuildpack.Logger, options *SealightsOptions) *AgentInstaller {
	return &AgentInstaller{Log: log, Options: options, MaxDownloadRetries: 3}
}

func (agi *AgentInstaller) InstallAgent(stager *libbuildpack.Stager) (string, string, error) {
	packageName := getPackageNameByPlatform()
	installationPath := filepath.Join(stager.BuildDir(), AgentDir)
	archivePath, err := agi.downloadPackage(packageName)
	if err != nil {
		return "", "", err
	}

	err = agi.extractPackage(archivePath, installationPath)
	if err != nil {
		return "", "", err
	}

	agentVersion := agi.readAgentVersion(installationPath)

	return AgentDir, agentVersion, nil
}

func (agi *AgentInstaller) downloadPackage(packageName string) (string, error) {
	url := agi.getDownloadUrl(packageName)

	agi.Log.Debug("Sealights. Download package started. From '%s'", url)

	tempAgentFile := filepath.Join(os.TempDir(), packageName)
	err := agi.downloadFileWithRetry(url, tempAgentFile, agi.MaxDownloadRetries)
	if err != nil {
		agi.Log.Error("Sealights. Failed to download package.")
		return "", err
	}

	agi.Log.Debug("Sealights. Download finished.")
	return tempAgentFile, nil
}

func (agi *AgentInstaller) extractPackage(source string, target string) error {
	agi.Log.Debug("Sealights. Extract package from '%s' to '%s'", source, target)

	var err error
	var isZip = strings.HasSuffix(source, ".zip")
	if isZip {
		err = libbuildpack.ExtractZip(source, target)
	} else {
		err = libbuildpack.ExtractTarGz(source, target)
	}

	if err != nil {
		agi.Log.Error("Sealights. Failed to extract package.")
		return err
	}

	agi.Log.Debug("Sealights. Package extracted.")
	return nil
}

func (agi *AgentInstaller) getDownloadUrl(packageName string) string {
	if agi.Options.CustomAgentUrl != "" {
		return agi.Options.CustomAgentUrl
	}

	version := DefaultVersion
	if agi.Options.Version != "" {
		version = agi.Options.Version
	}

	// resulting url example:
	// https://agents.sealights.co/dotnetcore/latest/sealights-dotnet-agent-linux-self-contained.tar.gz
	url := fmt.Sprintf(AgentDownloadUrlFormat, version, packageName)

	return url
}

func (agi *AgentInstaller) downloadFileWithRetry(url string, filePath string, MaxDownloadRetries int) error {
	const baseWaitTime = 3 * time.Second

	var err error
	for i := 0; i < MaxDownloadRetries; i++ {
		err = agi.downloadFile(url, filePath)
		if err == nil {
			return nil
		}

		waitTime := baseWaitTime + time.Duration(math.Pow(2, float64(i)))*time.Second
		time.Sleep(waitTime)
	}

	return err
}

func (agi *AgentInstaller) downloadFile(agentUrl string, destFile string) error {
	client := agi.createClient()

	resp, err := client.Get(agentUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("could not download: %d", resp.StatusCode)
	}

	return writeToFile(resp.Body, destFile, 0666)
}

// Create simple client or client with proxy, based on the settings
func (agi *AgentInstaller) createClient() *http.Client {
	if agi.Options.Proxy != "" {
		proxyUrl, _ := url.Parse(agi.Options.Proxy)

		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{
					Scheme: proxyUrl.Scheme,
					User:   url.UserPassword(agi.Options.ProxyUsername, agi.Options.ProxyPassword),
					Host:   proxyUrl.Host,
				}),
			},
		}
	} else {
		return &http.Client{}
	}
}

func (agi *AgentInstaller) readAgentVersion(installationPath string) string {
	data, err := os.ReadFile(filepath.Join(installationPath, VersionFileName))
	if err != nil {
		agi.Log.Warning("Failed to get agent version: %v", err)
		return "unknown"
	}

	return string(data)
}

func writeToFile(source io.Reader, destFile string, mode os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(destFile), 0755)
	if err != nil {
		return err
	}

	fh, err := os.OpenFile(destFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fh, source)
	if err != nil {
		return err
	}

	return nil
}

func getPackageNameByPlatform() string {
	if runtime.GOOS == "windows" {
		return WindowsPackageName
	} else {
		return LinuxPackageName
	}
}
