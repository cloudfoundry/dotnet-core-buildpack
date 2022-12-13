package finalize

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/config"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/project"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/kr/text"
)

var stackToRuntimeRID = map[string]string{
	"cflinuxfs3": "ubuntu.18.04-x64",
	"cflinuxfs4": "ubuntu.22.04-x64",
}

type Project interface {
	IsPublished() (bool, error)
	StartCommand() (string, error)
	InstallFrameworks() error
	ProjFilePaths() ([]string, error)
	MainPath() (string, error)
	IsFDD() (bool, error)
}

type Stager interface {
	BuildDir() string
	DepsIdx() string
	DepDir() string
	WriteProfileD(string, string) error
}

type Command interface {
	Run(*exec.Cmd) error
}

type Finalizer struct {
	Stager  Stager
	Log     *libbuildpack.Logger
	Command Command
	Config  *config.Config
	Project *project.Project
}

func Run(f *Finalizer) error {
	f.Log.BeginStep("Finalizing Dotnet Core")
	isFrameworkDependent, err := f.Project.IsFDD()
	if err != nil {
		return err
	}

	isSourceBased, err := f.Project.IsSourceBased()
	if err != nil {
		return err
	}

	stack := os.Getenv("CF_STACK")
	stackRID := stackToRuntimeRID[stack]
	if stackRID == "" {
		f.Log.Error("Unsupported stack: %s", stack)
		return err
	}

	if isSourceBased {
		if err := f.Project.SourceInstallDotnetRuntime(); err != nil {
			f.Log.Error("Unable to install dotnet-runtime: %s", err.Error())
			return err
		}

		if err := f.Project.SourceInstallDotnetAspNetCore(); err != nil {
			f.Log.Error("Unable to install dotnet-aspnetcore: %s", err.Error())
			return err
		}

		if err := f.DotnetPublish(stackRID); err != nil {
			f.Log.Error("Unable to run dotnet publish: %s", err.Error())
			return err
		}
	}

	if isFrameworkDependent {
		if err := f.Project.FDDInstallFrameworks(); err != nil {
			f.Log.Error("Unable to install frameworks: %s", err.Error())
			return err
		}
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

	isFDD, err := f.Project.IsFDD()
	if err != nil {
		return err
	}

	startCmd, err := f.Project.StartCommand()
	if err != nil {
		return err
	}

	if !(isFDD || strings.HasSuffix(startCmd, ".dll")) {
		dirsToRemove = append(dirsToRemove, "dotnet-sdk")
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
		files, err := os.ReadDir(filepath.Join(f.Stager.DepDir(), name))
		if err != nil {
			return err
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				return err
			}
			if (info.Mode() & os.ModeSymlink) != 0 {
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
	scriptContents := fmt.Sprintf(`
export ASPNETCORE_URLS=http://0.0.0.0:${PORT}
export DOTNET_ROOT=%s
`, filepath.Join("/home", "vcap", "deps", f.Stager.DepsIdx(), "dotnet-sdk"))

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
		"default_process_types": {"web": fmt.Sprintf("cd %s && exec %s --server.urls http://0.0.0.0:${PORT}", directory, startCmd)},
	}, nil
}

func (f *Finalizer) DotnetPublish(stackRID string) error {
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
	env = append(env, "PATH="+filepath.Join(filepath.Dir(mainProject), "node_modules", ".bin")+":"+os.Getenv("PATH"))

	publishPath := filepath.Join(f.Stager.DepDir(), "dotnet_publish")
	if err := os.MkdirAll(publishPath, 0755); err != nil {
		return err
	}
	args := []string{"publish", mainProject, "-o", publishPath, "-c", f.publicConfig()}
	args = append(args, "-r", stackRID)
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
