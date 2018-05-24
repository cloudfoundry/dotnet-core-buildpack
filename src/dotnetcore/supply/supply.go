package supply

import (
	"crypto/md5"
	"dotnetcore/config"
	"dotnetcore/project"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	if err := s.InstallDotnet(); err != nil {
		s.Log.Error("Unable to install Dotnet: %s", err.Error())
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
		return err
	}
	if shouldInstallNode {
		if err := s.Installer.InstallOnlyVersion("node", s.Stager.DepDir()); err != nil {
			return err
		}
		version := s.Manifest.AllDependencyVersions("node")[0]
		if err := os.Rename(filepath.Join(s.Stager.DepDir(), fmt.Sprintf("node-v%s-linux-x64", version)), filepath.Join(s.Stager.DepDir(), "node")); err != nil {
			return err
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
		return false, err
	} else if isPublished {
		return false, nil
	}

	return s.commandsInProjFiles([]string{"npm", "bower"})
}

func (s *Supplier) commandsInProjFiles(commands []string) (bool, error) {
	projFiles, err := s.Project.Paths()
	if err != nil {
		return false, err
	}

	for _, projFile := range projFiles {
		projFile = filepath.Join(s.Stager.BuildDir(), projFile)
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
			return false, err
		}
		if err := xml.Unmarshal(projFileContent, &obj); err != nil {
			return false, err
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
	allVersions := s.Manifest.AllDependencyVersions("dotnet")

	installVersion, err := s.globalJsonSdkVersion()
	if err != nil {
		return "", err
	}

	if contains(allVersions, installVersion) {
		return installVersion, nil
	}

	if installVersion != "" {
		s.Log.Warning("SDK %s not available", installVersion)
		installVersion = majorMinorOnly(installVersion)
		installVersion, err = libbuildpack.FindMatchingVersion(installVersion, allVersions)
		if err == nil {
			s.Log.Info("using latest version in version line")
			return installVersion, nil
		}
	}

	if found, err := s.Project.IsFsharp(); err != nil {
		return "", err
	} else if found {
		s.Log.Info("using the default FSharp SDK")
		return libbuildpack.FindMatchingVersion("1.1.x", allVersions)
	}

	dep, err := s.Manifest.DefaultVersion("dotnet")
	if err != nil {
		return "", err
	}
	s.Log.Info("using the default SDK")
	return dep.Version, nil
}

func (s *Supplier) InstallDotnet() error {
	installVersion, err := s.pickVersionToInstall()
	if err != nil {
		return err
	}
	s.Config.DotnetSdkVersion = installVersion

	if err := s.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet", Version: installVersion}, filepath.Join(s.Stager.DepDir(), "dotnet")); err != nil {
		return err
	}

	return s.Stager.AddBinDependencyLink(filepath.Join(s.Stager.DepDir(), "dotnet", "dotnet"), "dotnet")
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
