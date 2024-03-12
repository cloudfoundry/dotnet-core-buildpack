package sealights

import (
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

const WindowsPackageName = "sealights-dotnet-agent-windows-self-contained.zip"
const LinuxPackageName = "sealights-dotnet-agent-linux-self-contained.tar.gz"
const WindowsPackageDir = "sealights-dotnet-agent-windows-self-contained"
const LinuxPackageDir = "sealights-dotnet-agent-linux-self-contained"

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
	installationPath := filepath.Join(stager.BuildDir(), AgentDir)
	archivePath, err := agi.downloadPackage()
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

func (agi *AgentInstaller) downloadPackage() (string, error) {
	url := agi.getDownloadUrl()

	agi.Log.Debug("Sealights. Download package started. From '%s'", url)

	tempAgentFile, err := agi.downloadFileWithRetry(url, agi.MaxDownloadRetries)
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
	var isZip = strings.HasSuffix(source, ".zip") || strings.HasSuffix(source, ".nupkg")
	if isZip {
		err = libbuildpack.ExtractZip(source, target)
	} else {
		err = libbuildpack.ExtractTarGz(source, target)
	}

	if err != nil {
		agi.Log.Error("Sealights. Failed to extract package.")
		return err
	}

	err = agi.extractContentIfNeeded(target)
	if err != nil {
		agi.Log.Error("Sealights. Failed to copy content from package")
		return err
	}

	agi.Log.Debug("Sealights. Package extracted.")
	return nil
}

func (agi *AgentInstaller) extractContentIfNeeded(target string) error {
	contentDirectory := filepath.Join(target, "content")
	found, err := libbuildpack.FileExists(contentDirectory)
	if err != nil {
		return err
	} else if found {
		// nuget package has different structure compare to
		// regular installation package. need to extract corresponding
		// agent from the content to align them

		singlePackage, err := libbuildpack.FileExists(filepath.Join(contentDirectory, "version.txt"))
		if err != nil {
			return err
		}

		agentDir := contentDirectory
		if !singlePackage {
			agentDir = filepath.Join(contentDirectory, getPackageDirByPlatform())
		}

		err = libbuildpack.MoveDirectory(agentDir, target)
		if err != nil {
			return err
		}

		agi.updateFilePermissions(target)
	}

	// remove "content" directory once it not needed
	os.RemoveAll(contentDirectory)

	return nil
}

func (agi *AgentInstaller) getDownloadUrl() string {
	if agi.Options.CustomAgentUrl != "" {
		return agi.Options.CustomAgentUrl
	}

	version := DefaultVersion
	if agi.Options.Version != "" {
		version = agi.Options.Version
	}

	packageName := getPackageNameByPlatform()

	// resulting url example:
	// https://agents.sealights.co/dotnetcore/latest/sealights-dotnet-agent-linux-self-contained.tar.gz
	url := fmt.Sprintf(AgentDownloadUrlFormat, version, packageName)

	return url
}

func (agi *AgentInstaller) downloadFileWithRetry(url string, MaxDownloadRetries int) (string, error) {
	const baseWaitTime = 3 * time.Second

	var err error
	var filePath string
	for i := 0; i < MaxDownloadRetries; i++ {
		filePath, err = agi.downloadFile(url)
		if err == nil {
			return filePath, nil
		}

		waitTime := baseWaitTime + time.Duration(math.Pow(2, float64(i)))*time.Second
		time.Sleep(waitTime)
	}

	return "", err
}

func (agi *AgentInstaller) downloadFile(agentUrl string) (string, error) {
	client := agi.createClient()

	resp, err := client.Get(agentUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("could not download: %d", resp.StatusCode)
	}

	fileName, err := guessFilename(resp)
	if err != nil {
		fileName = getPackageNameByPlatform()
	}

	destFile := filepath.Join(os.TempDir(), fileName)

	return destFile, writeToFile(resp.Body, destFile, 0666)
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

func (agi *AgentInstaller) updateFilePermissions(installationPath string) error {
	files, err := os.ReadDir(installationPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		os.Chmod(filepath.Join(installationPath, file.Name()), 0755)
	}

	return nil
}

func (agi *AgentInstaller) readAgentVersion(installationPath string) string {
	data, err := os.ReadFile(filepath.Join(installationPath, VersionFileName))
	if err != nil {
		agi.Log.Warning("Failed to get agent version: %v", err)
		return "unknown"
	}

	agentVersion := string(data)

	return strings.TrimSuffix(agentVersion, "\n")
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

func getPackageDirByPlatform() string {
	if runtime.GOOS == "windows" {
		return WindowsPackageDir
	} else {
		return LinuxPackageDir
	}
}

func guessFilename(resp *http.Response) (string, error) {
	filename := resp.Request.URL.Path

	cd := resp.Header.Get("Content-Disposition")

	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			val, ok := params["filename"]
			if ok {
				filename = val
			}
		}
	}

	// sanitize
	if filename == "" || strings.HasSuffix(filename, "/") || strings.Contains(filename, "\x00") {
		return "", errors.New("no file name")
	}

	filename = filepath.Base(path.Clean("/" + filename))
	if filename == "" || filename == "." || filename == "/" {
		return "", errors.New("no file name")
	}

	return filename, nil
}
