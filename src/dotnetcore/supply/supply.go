package supply

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/config"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/project"

	"github.com/cloudfoundry/libbuildpack"
)

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(string, string, ...string) (string, error)
}

type Manifest interface {
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	FetchDependency(libbuildpack.Dependency, string) error
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Stager interface {
	BuildDir() string
	CacheDir() string
	DepDir() string
	DepsIdx() string
	LinkDirectoryInDepDir(string, string) error
	AddBinDependencyLink(string, string) error
	WriteEnvFile(string, string) error
	WriteProfileD(string, string) error
	SetStagingEnvironment() error
}

type Supplier struct {
	Stager    Stager
	Manifest  Manifest
	Installer Installer
	Log       *libbuildpack.Logger
	Command   Command
	Config    *config.Config
	Project   *project.Project
}

func Run(s *Supplier) error {
	s.Log.BeginStep("Supplying Dotnet Core")

	if err := s.Command.Execute(s.Stager.BuildDir(), io.Discard, io.Discard, "touch", "/tmp/checkpoint"); err != nil {
		s.Log.Error("Unable to execute command: %s", err.Error())
		return err
	}

	if checksum, err := s.CalcChecksum(); err == nil {
		s.Log.Debug("BuildDir Checksum Before Supply: %s", checksum)
	}

	if err := s.InstallLibunwind(); err != nil {
		s.Log.Error("Unable to install Libunwind: %s", err.Error())
		return err
	}

	usesLibgdiplus, err := s.Project.UsesLibrary("System.Drawing.Common")
	if err != nil {
		s.Log.Error(`Error searching project for library "System.Drawing.Common": %s`, err.Error())
		return err
	}

	if usesLibgdiplus {
		if err := s.InstallLibgdiplus(); err != nil {
			s.Log.Error("Unable to install libgdiplus: %s", err.Error())
			return err
		}
	}

	if err := s.InstallDotnetSdk(); err != nil {
		s.Log.Error("Unable to install Dotnet SDK: %s", err.Error())
		return err
	}

	if err := s.LoadLegacySSLProvider(); err != nil {
		s.Log.Error("Unable to load the requested legacy SSL provider: %s", err.Error())
		return err
	}

	if err := s.InstallNode(); err != nil {
		s.Log.Error("Unable to install NodeJs: %s", err.Error())
		return err
	}

	if err := s.InstallBower(); err != nil {
		s.Log.Error("Unable to install Bower: %s", err.Error())
		return err
	}

	if err := s.Stager.SetStagingEnvironment(); err != nil {
		s.Log.Error("Unable to setup environment variables: %s", err.Error())
		return err
	}

	if checksum, err := s.CalcChecksum(); err == nil {
		s.Log.Debug("BuildDir Checksum After Supply: %s", checksum)
	}

	if filesChanged, err := s.Command.Output(s.Stager.BuildDir(), "find", ".", "-newer", "/tmp/checkpoint", "-not", "-path", "./.cloudfoundry/*", "-not", "-path", "./.cloudfoundry"); err == nil && filesChanged != "" {
		s.Log.Debug("Below files changed:")
		s.Log.Debug(filesChanged)
	}

	return nil
}

func (s *Supplier) InstallLibunwind() error {
	if err := s.Installer.InstallOnlyVersion("libunwind", filepath.Join(s.Stager.DepDir(), "libunwind")); err != nil {
		return err
	}

	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "libunwind", "lib"), "lib")
}

func (s *Supplier) InstallLibgdiplus() error {
	if err := s.Installer.InstallOnlyVersion("libgdiplus", filepath.Join(s.Stager.DepDir(), "libgdiplus")); err != nil {
		return err
	}

	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "libgdiplus", "lib"), "lib")
}

func (s *Supplier) shouldInstallBower() (bool, error) {
	err := s.Command.Execute(s.Stager.BuildDir(), io.Discard, io.Discard, "bower", "-v")
	if err == nil {
		return false, nil
	}

	if isPublished, err := s.Project.IsPublished(); err != nil {
		return false, err
	} else if isPublished {
		return false, nil
	}

	if commandsPresent, err := s.commandsInProjFiles([]string{"bower"}); err != nil {
		return false, err
	} else if commandsPresent {
		return true, nil
	}
	return false, nil
}

func (s *Supplier) bowerInstall() error {
	versions := s.Manifest.AllDependencyVersions("bower")
	dep := libbuildpack.Dependency{Name: "bower", Version: versions[0]}

	dir, err := os.MkdirTemp("", "dotnet-core_buildpack-bower")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := s.Installer.FetchDependency(dep, filepath.Join(dir, "bower.tar.gz")); err != nil {
		return err
	}

	if err := s.Command.Execute(s.Stager.BuildDir(), io.Discard, io.Discard, "npm", "install", "-g", filepath.Join(dir, "bower.tar.gz")); err != nil {
		return err
	}
	return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "node", "bin"), "bin")
}

func (s *Supplier) InstallBower() error {
	if shouldInstallBower, err := s.shouldInstallBower(); err != nil {
		return err
	} else if !shouldInstallBower {
		return nil
	}

	if err := s.Command.Execute(s.Stager.BuildDir(), io.Discard, io.Discard, "npm", "-v"); err != nil {
		return fmt.Errorf("Trying to install bower but NPM is not installed")
	}

	return s.bowerInstall()
}

func (s *Supplier) InstallNode() error {
	shouldInstallNode, err := s.shouldInstallNode()
	if err != nil {
		return fmt.Errorf("Could not decide whether to install node: %v", err)
	}
	if shouldInstallNode {
		version, err := libbuildpack.FindMatchingVersion("x", s.Manifest.AllDependencyVersions("node"))
		if err != nil {
			return err
		}

		dep := libbuildpack.Dependency{
			Name:    "node",
			Version: version,
		}

		nodePath := filepath.Join(s.Stager.DepDir(), "node")
		if err := s.Installer.InstallDependency(dep, nodePath); err != nil {
			return err
		}

		return s.Stager.LinkDirectoryInDepDir(filepath.Join(nodePath, "bin"), "bin")
	}
	return nil
}

func (s *Supplier) shouldInstallNode() (bool, error) {
	err := s.Command.Execute(s.Stager.BuildDir(), io.Discard, io.Discard, "node", "-v")
	if err == nil {
		return false, nil
	}

	if os.Getenv("INSTALL_NODE") != "" {
		return true, nil
	}

	if isPublished, err := s.Project.IsPublished(); err != nil {
		return false, fmt.Errorf("Could not determine if project is published: %v", err)
	} else if isPublished {
		return false, nil
	}

	return s.commandsInProjFiles([]string{"npm", "bower"})
}

func (s *Supplier) commandsInProjFiles(commands []string) (bool, error) {
	projFiles, err := s.Project.ProjectFilePaths()
	if err != nil {
		return false, fmt.Errorf("Could not get project file paths: %v", err)
	}

	for _, projFile := range projFiles {
		obj := struct {
			Sdk    string `xml:"Sdk,attr"`
			Target struct {
				Name          string `xml:"Name,attr"`
				BeforeTargets string `xml:"BeforeTargets,attr"`
				AfterTargets  string `xml:"AfterTargets,attr"`
				Exec          []struct {
					Command string `xml:"Command,attr"`
				} `xml:"Exec"`
			} `xml:"Target"`
		}{}

		projFileContent, err := os.ReadFile(projFile)
		if err != nil {
			return false, fmt.Errorf("Could not read project file: %v", err)
		}
		if err := xml.Unmarshal(projFileContent, &obj); err != nil {
			return false, fmt.Errorf("Could not unmarshal project file: %v", err)
		}

		targetNames := []string{"BeforeBuild", "BeforeCompile", "BeforePublish", "AfterBuild", "AfterCompile", "AfterPublish"}
		nameInTargetNames := false
		for _, name := range targetNames {
			if name == obj.Target.Name {
				nameInTargetNames = true
				break
			}
		}

		attrInTargetAttrs := obj.Target.BeforeTargets != "" || obj.Target.AfterTargets != ""

		if nameInTargetNames || attrInTargetAttrs {
			for _, ex := range obj.Target.Exec {
				command := ex.Command
				for _, cmd := range commands {
					if strings.Contains(command, cmd) {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil

}

// Turn a semver string into major.minor.x
// Will turn a.b.c into a.b.x
// Will not modify strings that don't match a.b.c
func sdkRollForward(version string, versions []string) (string, error) {
	// Filter versions that match the major.minor version
	versions = filterVersions(versions, version)

	var featureLine string
	var highestPatch string
	parts := strings.SplitN(version, ".", 3)
	if len(parts) == 3 {
		featureLine = parts[2][:1]
	}

	for _, v := range versions {
		versionSplit := strings.SplitN(v, ".", 3)
		if len(versionSplit) == 3 && versionSplit[2][:1] == featureLine {
			if highestPatch == "" {
				highestPatch = versionSplit[2][1:]
			} else {
				current, err := strconv.Atoi(highestPatch)
				comp, err := strconv.Atoi(versionSplit[2][1:])
				if err != nil {
					return "", err
				}
				if current < comp {
					highestPatch = versionSplit[2][1:]
				}
			}
		}
	}

	if highestPatch == "" {
		return "", fmt.Errorf("could not find sdk in same feature line as '%s'", version)
	}

	return fmt.Sprintf("%s.%s.%s%s", parts[0], parts[1], featureLine, highestPatch), nil
}

func filterVersions(versions []string, version string) []string {
	var filtered []string
	semver := strings.SplitN(version, ".", 3)
	major, minor := semver[0], semver[1]
	for _, v := range versions {
		if strings.HasPrefix(v, major+"."+minor) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (s *Supplier) pickVersionToInstall() (string, error) {
	allVersions := s.Manifest.AllDependencyVersions("dotnet-sdk")
	buildpackYamlVersion, err := s.parseBuildpackYamlVersion()
	if err != nil {
		return "", err
	}

	if buildpackYamlVersion != "" {
		version, err := project.FindMatchingVersionWithPreview(buildpackYamlVersion, allVersions)
		if err != nil {
			s.Log.Warning("SDK %s in buildpack.yml is not available", buildpackYamlVersion)
			return "", err
		}
		return version, err
	}

	globalJSONVersion, err := s.globalJsonSdkVersion()
	if err != nil {
		return "", err
	}

	if globalJSONVersion != "" {
		if contains(allVersions, globalJSONVersion) {
			return globalJSONVersion, nil
		}
		s.Log.Warning("SDK %s in global.json is not available", globalJSONVersion)
		installVersion, err := sdkRollForward(globalJSONVersion, allVersions)
		if err == nil {
			s.Log.Info("falling back to latest version in version line")
			return installVersion, nil
		}
		return "", err
	}

	dep, err := s.Manifest.DefaultVersion("dotnet-sdk")
	if err != nil {
		return "", err
	}
	s.Log.Info("using the default SDK")
	return dep.Version, nil
}

func (s *Supplier) InstallDotnetSdk() error {
	installVersion, err := s.pickVersionToInstall()
	if err != nil {
		return err
	}
	s.Config.DotnetSdkVersion = installVersion

	if err := s.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet-sdk", Version: installVersion}, filepath.Join(s.Stager.DepDir(), "dotnet-sdk")); err != nil {
		return err
	}

	if err := s.Stager.AddBinDependencyLink(filepath.Join(s.Stager.DepDir(), "dotnet-sdk", "dotnet"), "dotnet"); err != nil {
		return err
	}

	return s.installRuntimeIfNeeded()
}

func (s *Supplier) LoadLegacySSLProvider() error {
	loadLegacySSLProvider, err := s.parseBuildpackYamlOpenssl()
	if err != nil {
		return err
	}

	if loadLegacySSLProvider {
		// If a user sets the buidpack.yml to include legacy provider AND
		// includes their own openssl.cnf, just use the provided openssl.cnf
		opensslCnfFile := filepath.Join(s.Stager.BuildDir(), "openssl.cnf")
		exists, err := libbuildpack.FileExists(opensslCnfFile)
		if err != nil {
			// untested
			return err
		}

		s.Log.BeginStep("Loading legacy SSL provider")
		if !exists {
			content := `[provider_sect]
default = default_sect
legacy = legacy_sect

[default_sect]
activate = 1

[legacy_sect]
activate = 1`
			err := os.WriteFile(opensslCnfFile, []byte(content), 0644)
			if err != nil {
				s.Log.Info("Cannot write openssl.cnf file to build directory")
				// untested
				return err
			}
		} else {
			s.Log.Info("Application already contains openssl.cnf file")
		}
	}

	return nil
}

func (s *Supplier) installRuntimeIfNeeded() error {
	runtimeVersionPath := filepath.Join(s.Stager.DepDir(), "dotnet-sdk", "RuntimeVersion.txt")

	exists, err := libbuildpack.FileExists(runtimeVersionPath)
	if err != nil {
		return err
	} else if exists {
		version, err := os.ReadFile(runtimeVersionPath)
		if err != nil {
			return err
		}
		name := "dotnet-runtime"
		runtimeVersion, err := s.Project.FindMatchingFrameworkVersion(name, string(version), nil)
		if err != nil {
			return err
		}
		return s.Installer.InstallDependency(libbuildpack.Dependency{Name: name, Version: runtimeVersion}, filepath.Join(s.Stager.DepDir(), "dotnet-sdk"))
	}
	return nil
}

func (s *Supplier) globalJsonSdkVersion() (string, error) {
	if found, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "global.json")); err != nil || !found {
		return "", err
	}

	obj := struct {
		Sdk struct {
			Version string `json:"version"`
		} `json:"sdk"`
	}{}
	if err := libbuildpack.NewJSON().Load(filepath.Join(s.Stager.BuildDir(), "global.json"), &obj); err != nil {
		return "", err
	}
	return obj.Sdk.Version, nil
}

func (s *Supplier) CalcChecksum() (string, error) {
	h := md5.New()
	basepath := s.Stager.BuildDir()
	err := filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			relpath, err := filepath.Rel(basepath, path)
			if strings.HasPrefix(relpath, ".cloudfoundry/") {
				return nil
			}
			if err != nil {
				return err
			}
			if _, err := io.WriteString(h, relpath); err != nil {
				return err
			}
			if f, err := os.Open(path); err != nil {
				return err
			} else {
				if _, err := io.Copy(h, f); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type buildpackYaml struct {
	DotnetCore struct {
		Version string `yaml:"sdk"`
	} `yaml:"dotnet-core"`
	UseLegacyOpenssl bool `yaml:"use_legacy_openssl"`
}

func (s *Supplier) parseBuildpackYamlVersion() (string, error) {
	content, err := s.parseBuildpackYamlFile()
	if err != nil {
		return "", err
	}
	return content.DotnetCore.Version, nil
}

func (s *Supplier) parseBuildpackYamlOpenssl() (bool, error) {
	content, err := s.parseBuildpackYamlFile()
	if err != nil {
		return false, err
	}
	return content.UseLegacyOpenssl, nil
}

func (s *Supplier) parseBuildpackYamlFile() (buildpackYaml, error) {
	obj := buildpackYaml{}
	if found, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "buildpack.yml")); err != nil || !found {
		return obj, err
	}

	if err := libbuildpack.NewYAML().Load(filepath.Join(s.Stager.BuildDir(), "buildpack.yml"), &obj); err != nil {
		return obj, err
	}

	return obj, nil
}
