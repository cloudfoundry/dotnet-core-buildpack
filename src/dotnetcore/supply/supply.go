package supply

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	if err := s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "touch", "/tmp/checkpoint"); err != nil {
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
	if err := s.InstallDotnetSdk(); err != nil {
		s.Log.Error("Unable to install Dotnet SDK: %s", err.Error())
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

func (s *Supplier) shouldInstallBower() (bool, error) {
	err := s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "bower", "-v")
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

	dir, err := ioutil.TempDir("", "dotnet-core_buildpack-bower")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := s.Installer.FetchDependency(dep, filepath.Join(dir, "bower.tar.gz")); err != nil {
		return err
	}

	if err := s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "npm", "install", "-g", filepath.Join(dir, "bower.tar.gz")); err != nil {
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

	if err := s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "npm", "-v"); err != nil {
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
		if err := s.Installer.InstallOnlyVersion("node", s.Stager.DepDir()); err != nil {
			return fmt.Errorf("Attempted to install node, but failed: %v", err)
		}
		version := s.Manifest.AllDependencyVersions("node")[0]
		oldfilename := filepath.Join(s.Stager.DepDir(), fmt.Sprintf("node-v%s-linux-x64", version))
		newfilename := filepath.Join(s.Stager.DepDir(), "node")
		if err := os.Rename(oldfilename, newfilename); err != nil {
			return fmt.Errorf("Could not rename '%s' to '%s': %v", oldfilename, newfilename, err)
		}
		return s.Stager.LinkDirectoryInDepDir(filepath.Join(s.Stager.DepDir(), "node", "bin"), "bin")
	}
	return nil
}

func (s *Supplier) shouldInstallNode() (bool, error) {
	err := s.Command.Execute(s.Stager.BuildDir(), ioutil.Discard, ioutil.Discard, "node", "-v")
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

		projFileContent, err := ioutil.ReadFile(projFile)
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
func majorMinorOnly(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) == 3 {
		parts[2] = "x" // ignore patch version
		return strings.Join(parts, ".")
	}
	return version
}

func (s *Supplier) pickVersionToInstall() (string, error) {
	allVersions := s.Manifest.AllDependencyVersions("dotnet-sdk")

	buildpackVersion, err := s.buildpackYamlSdkVersion()
	if err != nil {
		return "", err
	}
	if buildpackVersion != "" {
		version, err := libbuildpack.FindMatchingVersion(buildpackVersion, allVersions)
		if err != nil {
			s.Log.Warning("SDK %s in buildpack.yml is not available", buildpackVersion)
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
		installVersion, err := libbuildpack.FindMatchingVersion(majorMinorOnly(globalJSONVersion), allVersions)
		if err == nil {
			s.Log.Info("falling back to latest version in version line")
			return installVersion, nil
		}
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

func (s *Supplier) installRuntimeIfNeeded() error {
	runtimeVersionPath := filepath.Join(s.Stager.DepDir(), "dotnet-sdk", "RuntimeVersion.txt")

	exists, err := libbuildpack.FileExists(runtimeVersionPath)
	if err != nil {
		return err
	} else if exists {
		version, err := ioutil.ReadFile(runtimeVersionPath)
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

func (s *Supplier) suppliedVersion(allVersions []string) (string, error) {
	buildpackVersion, err := s.buildpackYamlSdkVersion()
	if err != nil {
		return "", err
	}

	if buildpackVersion != "" {
		version, err := libbuildpack.FindMatchingVersion(buildpackVersion, allVersions)
		if err != nil {
			s.Log.Warning("SDK %s in buildpack.yml is not available", buildpackVersion)
		}
		return version, err
	}

	globalJSONVersion, err := s.globalJsonSdkVersion()
	if err != nil {
		return "", err
	}
	if globalJSONVersion == "" {
		return "", nil
	}

	if contains(allVersions, globalJSONVersion) {
		return globalJSONVersion, nil
	}
	s.Log.Warning("SDK %s in global.json is not available", globalJSONVersion)
	installVersion, err := libbuildpack.FindMatchingVersion(majorMinorOnly(globalJSONVersion), allVersions)
	if err != nil {
		return "", nil
	}
	s.Log.Info("falling back to latest version in version line")
	return installVersion, nil
}

func (s *Supplier) buildpackYamlSdkVersion() (string, error) {
	if found, err := libbuildpack.FileExists(filepath.Join(s.Stager.BuildDir(), "buildpack.yml")); err != nil || !found {
		return "", err
	}

	obj := struct {
		DotnetCore struct {
			Version string `yaml:"sdk"`
		} `yaml:"dotnet-core"`
	}{}
	if err := libbuildpack.NewYAML().Load(filepath.Join(s.Stager.BuildDir(), "buildpack.yml"), &obj); err != nil {
		return "", err
	}

	return obj.DotnetCore.Version, nil
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
