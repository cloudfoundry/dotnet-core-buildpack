package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/go-ini/ini"
)

type Project struct {
	buildDir string
	depDir   string
	depsIdx  string
}

func New(buildDir, depDir, depsIdx string) *Project {
	return &Project{buildDir: buildDir, depDir: depDir, depsIdx: depsIdx}
}

func (p *Project) IsPublished() (bool, error) {
	if path, err := p.RuntimeConfigFile(); err != nil {
		return false, err
	} else {
		return path != "", nil
	}
}

func (p *Project) Paths() ([]string, error) {
	paths := []string{}
	if err := filepath.Walk(p.buildDir, func(path string, _ os.FileInfo, err error) error {
		if strings.Contains(path, "/.cloudfoundry/") {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ".csproj") || strings.HasSuffix(path, ".vbproj") || strings.HasSuffix(path, ".fsproj") {
			if path, err := filepath.Rel(p.buildDir, path); err != nil {
				return err
			} else {
				paths = append(paths, path)
			}
		}
		return nil
	}); err != nil {
		return []string{}, err
	}
	return paths, nil
}

func (p *Project) IsFsharp() (bool, error) {
	if paths, err := p.Paths(); err != nil {
		return false, err
	} else {
		for _, path := range paths {
			if strings.HasSuffix(path, ".fsproj") {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *Project) RuntimeConfigFile() (string, error) {
	if configFiles, err := filepath.Glob(filepath.Join(p.buildDir, "*.runtimeconfig.json")); err != nil {
		return "", err
	} else if len(configFiles) == 1 {
		return configFiles[0], nil
	} else if len(configFiles) > 1 {
		return "", fmt.Errorf("Multiple .runtimeconfig.json files present")
	}
	return "", nil
}

func (p *Project) MainPath() (string, error) {
	if runtimeConfigFile, err := p.RuntimeConfigFile(); err != nil {
		return "", err
	} else if runtimeConfigFile != "" {
		return runtimeConfigFile, nil
	}
	paths, err := p.Paths()
	if err != nil {
		return "", err
	}

	if len(paths) == 1 {
		return paths[0], nil
	} else if len(paths) > 1 {
		if exists, err := libbuildpack.FileExists(filepath.Join(p.buildDir, ".deployment")); err != nil {
			return "", err
		} else if exists {
			deployment, err := ini.Load(filepath.Join(p.buildDir, ".deployment"))
			if err != nil {
				return "", err
			}
			config, err := deployment.GetSection("config")
			if err != nil {
				return "", err
			}
			project, err := config.GetKey("project")
			if err != nil {
				return "", err
			}
			return filepath.Join(p.buildDir, strings.Trim(project.String(), ".")), nil
		}
		return "", fmt.Errorf("Multiple paths: %v contain a project file, but no .deployment file was used", paths)
	}
	return "", nil
}

func (p *Project) publishedStartCommand(projectPath string) (string, error) {
	var publishedPath string
	var runtimePath string

	if published, err := p.IsPublished(); err != nil {
		return "", err
	} else if published {
		publishedPath = p.buildDir
		runtimePath = "${HOME}"
	} else {
		publishedPath = filepath.Join(p.depDir, "dotnet_publish")
		runtimePath = filepath.Join("${DEPS_DIR}", p.depsIdx, "dotnet_publish")
	}

	if exists, err := libbuildpack.FileExists(filepath.Join(publishedPath, projectPath)); err != nil {
		return "", nil
	} else if exists {
		if err := os.Chmod(filepath.Join(filepath.Join(publishedPath, projectPath)), 0755); err != nil {
			return "", nil
		}
		return filepath.Join(runtimePath, projectPath), nil
	}

	if exists, err := libbuildpack.FileExists(filepath.Join(publishedPath, fmt.Sprintf("%s.dll", projectPath))); err != nil {
		return "", nil
	} else if exists {
		return fmt.Sprintf("%s.dll", filepath.Join(runtimePath, projectPath)), nil
	}
	return "", nil
}

func (p *Project) StartCommand() (string, error) {
	projectPath, err := p.MainPath()
	if err != nil {
		return "", err
	} else if projectPath == "" {
		return "", nil
	}

	re := regexp.MustCompile(`\.(runtimeconfig\.json|[a-z]+proj)$`)
	projectPath = re.ReplaceAllString(projectPath, "")
	projectPath = filepath.Base(projectPath)

	return p.publishedStartCommand(projectPath)
}
