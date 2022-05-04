package sealights

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

const PackageArchiveName = "sealights-agent.tar.gz"
const DefaultLabId = "agents"
const DefaultVersion = "latest"
const AgentDir = "sealights"
const DotnetDir = "dotnet-sdk"

type AgentInstaller struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	MaxDownloadRetries int
}

func NewAgentInstaller(log *libbuildpack.Logger, options *SealightsOptions) *AgentInstaller {
	return &AgentInstaller{Log: log, Options: options, MaxDownloadRetries: 3}
}

func (agi *AgentInstaller) InstallAgent(stager *libbuildpack.Stager) (string, error) {
	installationPath := filepath.Join(stager.BuildDir(), AgentDir)
	archivePath, err := agi.downloadPackage()
	if err != nil {
		return "", err
	}

	err = agi.extractPackage(archivePath, installationPath)
	if err != nil {
		return "", err
	}

	return filepath.Join("${HOME}", AgentDir), nil
}

func (agi *AgentInstaller) InstallDependency(stager *libbuildpack.Stager) (string, error) {
	dependencyPath := filepath.Join(stager.BuildDir(), AgentDir, DotnetDir)
	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		agi.Log.Error("Unable to determine buildpack directory: %s", err.Error())
		return "", err
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, agi.Log, time.Now())
	if err != nil {
		agi.Log.Error("Unable to load buildpack manifest: %s", err.Error())
		return "", err
	}

	sdkVersion, runtimeVersion := agi.selectDotnetVersions(manifest)
	depinstaller := libbuildpack.NewInstaller(manifest)

	if err = depinstaller.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-sdk", Version: sdkVersion},
		dependencyPath,
	); err != nil {
		agi.Log.Error("Sealights. Failed to install dotnet sdk")
		return "", err
	}

	if err = depinstaller.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
		dependencyPath,
	); err != nil {
		agi.Log.Error("Sealights. Failed to install dotnet runtime")
		return "", err
	}

	return filepath.Join("${HOME}", AgentDir, DotnetDir), nil
}

func (agi *AgentInstaller) selectDotnetVersions(manifest *libbuildpack.Manifest) (sdkVersion string, runtimeVersion string) {
	sdkVersions := manifest.AllDependencyVersions("dotnet-sdk")
	sdkVersion, _ = libbuildpack.FindMatchingVersion("6.0.2x", sdkVersions)
	if sdkVersion == "" {
		sdkVersion = "6.0.202"
	}

	runtimeVersions := manifest.AllDependencyVersions("dotnet-runtime")
	runtimeVersion, _ = libbuildpack.FindMatchingVersion("6.0.x", runtimeVersions)
	if runtimeVersion == "" {
		runtimeVersion = "6.0.3"
	}

	return
}

func (agi *AgentInstaller) downloadPackage() (string, error) {
	url := agi.getDownloadUrl()

	agi.Log.Debug("Sealights. Download package started. From '%s'", url)

	tempAgentFile := filepath.Join(os.TempDir(), PackageArchiveName)
	err := downloadFileWithRetry(url, tempAgentFile, agi.MaxDownloadRetries)
	if err != nil {
		agi.Log.Error("Sealights. Failed to download package.")
		return "", err
	}

	agi.Log.Debug("Sealights. Download finished.")
	return tempAgentFile, nil
}

func (agi *AgentInstaller) extractPackage(source string, target string) error {
	agi.Log.Debug("Sealights. Extract package from '%s' to '%s'", source, target)

	err := libbuildpack.ExtractTarGz(source, target)
	if err != nil {
		agi.Log.Error("Sealights. Failed to extract package.")
		return err
	}

	agi.Log.Debug("Sealights. Package extracted.")
	return nil
}

func (agi *AgentInstaller) getDownloadUrl() string {
	if agi.Options.CustomAgentUrl != "" {
		return agi.Options.CustomAgentUrl
	}

	labId := DefaultLabId
	if agi.Options.LabId != "" {
		labId = agi.Options.LabId
	}

	version := DefaultVersion
	if agi.Options.Version != "" {
		version = agi.Options.Version
	}

	url := fmt.Sprintf("https://%s.sealights.co/dotnetcore/sealights-dotnet-agent-%s.tar.gz", labId, version)

	return url
}

func downloadFileWithRetry(url string, filePath string, MaxDownloadRetries int) error {
	const baseWaitTime = 3 * time.Second

	var err error
	for i := 0; i < MaxDownloadRetries; i++ {
		err = downloadFile(url, filePath)
		if err == nil {
			return nil
		}

		waitTime := baseWaitTime + time.Duration(math.Pow(2, float64(i)))*time.Second
		time.Sleep(waitTime)
	}

	return err
}

func downloadFile(url, destFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Could not download: %d", resp.StatusCode)
	}

	return writeToFile(resp.Body, destFile, 0666)
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
