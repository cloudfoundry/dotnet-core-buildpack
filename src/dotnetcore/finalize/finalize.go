package finalize

import (
	"dotnetcore/config"
	"dotnetcore/project"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/kr/text"
)

type Stager interface {
	BuildDir() string
	DepsIdx() string
	DepDir() string
	WriteProfileD(string, string) error
}

type Command interface {
	Run(*exec.Cmd) error
}

type DotnetFramework interface {
	Install() error
}

type Finalizer struct {
	Stager          Stager
	Log             *libbuildpack.Logger
	Command         Command
	DotnetFramework DotnetFramework
	Config          *config.Config
	Project         *project.Project
}

func Run(f *Finalizer) error {
	f.Log.BeginStep("Finalizing Dotnet Core")

	if err := f.DotnetRestore(); err != nil {
		f.Log.Error("Unable to run dotnet restore: %s", err.Error())
		return err
	}

	if err := f.DotnetFramework.Install(); err != nil {
		f.Log.Error("Unable to install required dotnet frameworks: %s", err.Error())
		return err
	}

	if err := f.DotnetPublish(); err != nil {
		f.Log.Error("Unable to run dotnet publish: %s", err.Error())
		return err
	}

	if err := f.CleanStagingArea(); err != nil {
		f.Log.Error("Unable to run CleanStagingArea: %s", err.Error())
		return err
	}

	if err := f.WriteProfileD(); err != nil {
		f.Log.Error("Unable to write profile.d: %s", err.Error())
		return err
	}

	data, err := f.GenerateReleaseYaml()
	if err != nil {
		f.Log.Error("Error generating release YAML: %s", err)
		return err
	}
	releasePath := filepath.Join(f.Stager.BuildDir(), "tmp", "dotnet-core-buildpack-release-step.yml")
	return libbuildpack.NewYAML().Write(releasePath, data)
}

func (f *Finalizer) CleanStagingArea() error {
	f.Log.BeginStep("Cleaning staging area")

	dirsToRemove := []string{"nuget", ".nuget", ".local", ".cache", ".config", ".npm"}

	if startCmd, err := f.Project.StartCommand(); err != nil {
		return err
	} else if !strings.HasSuffix(startCmd, ".dll") {
		dirsToRemove = append(dirsToRemove, "dotnet")
	}
	if os.Getenv("INSTALL_NODE") != "true" {
		dirsToRemove = append(dirsToRemove, "node")
	}

	for _, dir := range dirsToRemove {
		if found, err := libbuildpack.FileExists(filepath.Join(f.Stager.DepDir(), dir)); err != nil {
			return err
		} else if found {
			f.Log.Info("Removing %s", dir)
			if err := os.RemoveAll(filepath.Join(f.Stager.DepDir(), dir)); err != nil {
				return err
			}
			if err := f.removeSymlinksTo(filepath.Join(f.Stager.DepDir(), dir)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Finalizer) removeSymlinksTo(dir string) error {
	for _, name := range []string{"bin", "lib"} {
		files, err := ioutil.ReadDir(filepath.Join(f.Stager.DepDir(), name))
		if err != nil {
			return err
		}

		for _, file := range files {
			if (file.Mode() & os.ModeSymlink) != 0 {
				source := filepath.Join(f.Stager.DepDir(), name, file.Name())
				target, err := os.Readlink(source)
				if err != nil {
					return err
				}
				if strings.HasPrefix(target, dir) {
					if err := os.Remove(source); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (f *Finalizer) WriteProfileD() error {
	scriptContents := "export ASPNETCORE_URLS=http://0.0.0.0:${PORT}\n"

	return f.Stager.WriteProfileD("startup.sh", scriptContents)
}

func (f *Finalizer) GenerateReleaseYaml() (map[string]map[string]string, error) {
	startCmd, err := f.Project.StartCommand()
	if err != nil {
		return nil, err
	}
	directory := filepath.Dir(startCmd)
	startCmd = "./" + filepath.Base(startCmd)
	if strings.HasSuffix(startCmd, ".dll") {
		startCmd = "dotnet " + startCmd
	}
	return map[string]map[string]string{
		"default_process_types": {"web": fmt.Sprintf("cd %s && %s --server.urls http://0.0.0.0:${PORT}", directory, startCmd)},
	}, nil
}

func (f *Finalizer) DotnetRestore() error {
	if published, err := f.Project.IsPublished(); err != nil {
		return err
	} else if published {
		return nil
	}
	f.Log.BeginStep("Restore dotnet dependencies")
	env := f.shellEnvironment()
	paths, err := f.Project.Paths()
	if err != nil {
		return err
	}
	for _, path := range paths {
		cmd := exec.Command("dotnet", "restore", path)
		cmd.Dir = f.Stager.BuildDir()
		cmd.Env = env
		cmd.Stdout = indentWriter(os.Stdout)
		cmd.Stderr = indentWriter(os.Stderr)
		if err := f.Command.Run(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (f *Finalizer) DotnetPublish() error {
	if published, err := f.Project.IsPublished(); err != nil {
		return err
	} else if published {
		return nil
	}
	f.Log.BeginStep("Publish dotnet")

	mainProject, err := f.Project.MainPath()
	if err != nil {
		return err
	}

	env := f.shellEnvironment()
	env = append(env, "PATH="+filepath.Join(f.Stager.BuildDir(), filepath.Dir(mainProject), "node_modules", ".bin")+":"+os.Getenv("PATH"))

	publishPath := filepath.Join(f.Stager.DepDir(), "dotnet_publish")
	if err := os.MkdirAll(publishPath, 0755); err != nil {
		return err
	}
	args := []string{"publish", mainProject, "-o", publishPath, "-c", f.publicConfig()}
	if strings.HasPrefix(f.Config.DotnetSdkVersion, "2.") {
		args = append(args, "-r", "ubuntu.14.04-x64")
	}
	cmd := exec.Command("dotnet", args...)
	cmd.Dir = f.Stager.BuildDir()
	cmd.Env = env
	cmd.Stdout = indentWriter(os.Stdout)
	cmd.Stderr = indentWriter(os.Stderr)

	f.Log.Debug("Running command: %v", cmd)
	if err := f.Command.Run(cmd); err != nil {
		return err
	}

	return nil
}

func (f *Finalizer) publicConfig() string {
	if os.Getenv("PUBLISH_RELEASE_CONFIG") == "true" {
		return "Release"
	}

	return "Debug"
}

func (f *Finalizer) shellEnvironment() []string {
	env := os.Environ()
	for _, v := range []string{
		"DOTNET_SKIP_FIRST_TIME_EXPERIENCE=true",
		"DefaultItemExcludes=.cloudfoundry/**/*.*",
		"HOME=" + f.Stager.DepDir(),
	} {
		env = append(env, v)
	}
	return env
}

func indentWriter(writer io.Writer) io.Writer {
	return text.NewIndentWriter(writer, []byte("       "))
}
