package sealights

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

const ReleaseFileName = "dotnet-core-buildpack-release-step.yml"
const StartCommandType = "web"

// file format:
// default_process_types:
//     web: cd ${DEPS_DIR}/0/dotnet_publish && exec ./app --server.urls http://0.0.0.0:${PORT}
type ReleaseData struct {
	DefaultProcessTypes map[string]string `yaml:"default_process_types"`
}

type ReleaseInfo struct {
	Data     ReleaseData
	FilePath string
}

func NewReleaseInfo(buildDirectory string) *ReleaseInfo {
	releaseFilePath := filepath.Join(buildDirectory, "tmp", ReleaseFileName)
	releaseData, _ := parseReleaseData(releaseFilePath)
	return &ReleaseInfo{Data: releaseData, FilePath: releaseFilePath}
}

func (rel *ReleaseInfo) GetStartCommand() string {
	return rel.Data.DefaultProcessTypes[StartCommandType]
}

func (rel *ReleaseInfo) SetStartCommand(newCommand string) error {
	rel.Data.DefaultProcessTypes[StartCommandType] = newCommand
	return writeReleaseData(rel.FilePath, rel.Data)
}

func parseReleaseData(releaseFilePath string) (ReleaseData, error) {
	var releaseData ReleaseData
	err := libbuildpack.NewYAML().Load(releaseFilePath, &releaseData)
	return releaseData, err
}

func writeReleaseData(releaseFilePath string, releaseData ReleaseData) error {
	return libbuildpack.NewYAML().Write(releaseFilePath, releaseData)
}
