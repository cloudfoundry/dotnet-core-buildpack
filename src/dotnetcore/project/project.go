package project

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/go-ini/ini"
)

type ConfigJSON struct {
	RuntimeOptions struct {
		Framework struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"framework"`
		ApplyPatches *bool `json:"applyPatches"`
	} `json:"runtimeOptions"`
}

type Manifest interface {
	AllDependencyVersions(string) []string
}

type Installer interface {
	InstallDependency(libbuildpack.Dependency, string) error
}

type Project struct {
	buildDir  string
	depDir    string
	depsIdx   string
	manifest  Manifest
	installer Installer
	Log       *libbuildpack.Logger
}

func New(buildDir, depDir, depsIdx string, manifest Manifest, installer Installer, logger *libbuildpack.Logger) *Project {
	return &Project{
		buildDir:  buildDir,
		depDir:    depDir,
		depsIdx:   depsIdx,
		manifest:  manifest,
		installer: installer,
		Log:       logger,
	}
}

func (p *Project) IsPublished() (bool, error) {
	path, err := p.RuntimeConfigPath()
	if err != nil {
		return false, err
	}
	return path != "", nil
}

func (p *Project) StartCommand() (string, error) {
	projectPath, err := p.MainPath()
	if err != nil {
		return "", err
	} else if projectPath == "" {
		return "", nil
	}
	runtimeConfigRe := regexp.MustCompile(`\.(runtimeconfig\.json)$`)
	projRe := regexp.MustCompile(`\.([a-z]+proj)$`)

	if runtimeConfigRe.MatchString(projectPath) {
		projectPath = runtimeConfigRe.ReplaceAllString(projectPath, "")
		projectPath = filepath.Base(projectPath)
	} else if projRe.MatchString(projectPath) {
		assemblyName, err := p.getAssemblyName(projectPath)
		if err != nil {
			return "", err
		}
		if assemblyName != "" {
			projectPath = projRe.ReplaceAllString(assemblyName, "")
		} else {
			projectPath = projRe.ReplaceAllString(projectPath, "")
			projectPath = filepath.Base(projectPath)
		}
	}

	return p.publishedStartCommand(projectPath)
}

func (p *Project) FindMatchingFrameworkVersion(name, version string, applyPatches *bool) (string, error) {
	var err error
	if applyPatches == nil || *applyPatches {
		version, err = p.getLatestPatch(name, version)
		if err != nil {
			return "", err
		}
	}
	return version, nil
}

func (p *Project) GetVersionFromDepsJSON(library string) (string, error) {
	depsJSONFiles, err := filepath.Glob(filepath.Join(p.buildDir, "*.deps.json"))
	if err != nil {
		return "", err
	}

	if len(depsJSONFiles) == 1 {
		return p.getVersionFromAssetFile(depsJSONFiles[0], library)
	}

	return "", fmt.Errorf("multiple or no *.deps.json files present")
}

type libraryMissingError struct {
	s string
}

func (e *libraryMissingError) Error() string {
	return e.s
}

func (p *Project) getVersionFromAssetFile(path, library string) (string, error) {
	depsBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(depsBytes, &result); err != nil {
		return "", err
	}

	if _, ok := result["libraries"]; !ok {
		return "", &libraryMissingError{fmt.Sprintf("could not find library %s", library)}
	}

	libraries := result["libraries"].(map[string]interface{})
	for key := range libraries {
		re := regexp.MustCompile(fmt.Sprintf(`(%s)\/(\d\.\d\.\d)`, library))
		matchedString := re.FindStringSubmatch(key)
		if matchedString != nil {
			return matchedString[2], nil
		}
	}

	return "", &libraryMissingError{fmt.Sprintf("could not find library %s", library)}
}

func (p *Project) versionsFromNugetPackages(dependency string, rollForward bool) ([]string, error) {
	depToAssembly := map[string]string{
		"dotnet-runtime":    "microsoft.netcore.app",
		"dotnet-aspnetcore": "microsoft.aspnetcore.app",
	}

	restoredVersionsDir := filepath.Join(p.depDir, ".nuget", "packages", depToAssembly[dependency])
	if exists, err := libbuildpack.FileExists(restoredVersionsDir); err != nil {
		return []string{}, err
	} else if !exists {
		return []string{}, nil
	}

	files, err := ioutil.ReadDir(restoredVersionsDir)
	if err != nil {
		return []string{}, err
	}

	// Use this map as a set for de-duplication later on
	versions := map[string]interface{}{}
	for _, f := range files {
		if rollForward {
			version, err := p.getLatestPatch(dependency, f.Name())
			if err != nil {
				return []string{}, nil
			}
			versions[version] = nil
		} else {
			versions[f.Name()] = nil
		}
	}

	var distinctVersions []string
	for v := range versions {
		distinctVersions = append(distinctVersions, v)
	}

	return distinctVersions, nil
}

func (p *Project) VersionFromProjFile(mainProjectFile, regex, name string) (string, error) {
	var version string

	proj, err := ioutil.ReadFile(mainProjectFile)
	if err != nil {
		return "", err
	}

	matches := regexp.MustCompile(regex).FindStringSubmatch(string(proj))
	if len(matches) == 2 {
		version = matches[1]
		if version[len(version)-1] == '*' {
			return p.getLatestPatch(name, version)
		}
	}

	return version, nil
}

func (p *Project) ProjectFilePaths() ([]string, error) {
	var paths []string

	err := filepath.Walk(p.buildDir, func(path string, _ os.FileInfo, err error) error {
		if strings.Contains(path, "/.cloudfoundry/") {
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".csproj") || strings.HasSuffix(path, ".vbproj") || strings.HasSuffix(path, ".fsproj") {
			paths = append(paths, path)
		}

		return nil
	})

	if err != nil {
		return []string{}, err
	}

	return paths, nil
}

func (p *Project) IsFsharp() (bool, error) {
	if paths, err := p.ProjectFilePaths(); err != nil {
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

func (p *Project) RuntimeConfigPath() (string, error) {
	if configFiles, err := filepath.Glob(filepath.Join(p.buildDir, "*.runtimeconfig.json")); err != nil {
		return "", err
	} else if len(configFiles) == 1 {
		return configFiles[0], nil
	} else if len(configFiles) > 1 {
		return "", fmt.Errorf("multiple *.runtimeconfig.json files present")
	}
	return "", nil
}

func (p *Project) MainPath() (string, error) {
	runtimeConfigFile, err := p.RuntimeConfigPath()
	if err != nil {
		return "", err
	} else if runtimeConfigFile != "" {
		return runtimeConfigFile, nil
	}

	paths, err := p.ProjectFilePaths()
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

		return "", fmt.Errorf("multiple paths: %v contain a project file, but no .deployment file was used", paths)
	}

	return "", nil
}

func (p *Project) IsFDD() (bool, error) {
	path, err := p.RuntimeConfigPath()
	if err != nil {
		return false, err
	}

	if path != "" {
		runtimeJSON, err := ParseRuntimeConfig(path)
		if err != nil {
			return false, err
		}

		if runtimeJSON.RuntimeOptions.Framework.Name != "" {
			return true, nil
		}
	}
	return false, nil
}

func (p *Project) IsSourceBased() (bool, error) {
	path, err := p.RuntimeConfigPath()
	if err != nil {
		return false, err
	}

	return path == "", nil
}

func (p *Project) FDDInstallFrameworks() error {
	path, err := p.RuntimeConfigPath()
	if err != nil {
		return err
	}

	runtimeConfig, err := ParseRuntimeConfig(path)
	if err != nil {
		return err
	}

	frameworkName := runtimeConfig.RuntimeOptions.Framework.Name
	frameworkVersion := runtimeConfig.RuntimeOptions.Framework.Version
	applyPatches := runtimeConfig.RuntimeOptions.ApplyPatches

	if frameworkName == "Microsoft.NETCore.App" {
		return p.installFrameworksNETCoreApp(frameworkName, frameworkVersion, applyPatches)
	} else if frameworkName == "Microsoft.AspNetCore.All" || frameworkName == "Microsoft.AspNetCore.App" {
		return p.installFrameworksAspNetCoreApp(frameworkName, frameworkVersion, applyPatches)
	}

	return fmt.Errorf("invalid framework [%s] specified in application runtime config file", frameworkName)
}

func (p *Project) installFrameworksNETCoreApp(frameworkName, frameworkVersion string, applyPatches *bool) error {
	runtimeVersion, err := p.FindMatchingFrameworkVersion("dotnet-runtime", frameworkVersion, applyPatches)
	if err != nil {
		return err
	}

	exists, err := p.isInstalled(frameworkName, runtimeVersion)
	if err != nil {
		return err
	} else if exists {
		p.Log.Info("dotnet-runtime %s is already installed", runtimeVersion)
	} else {
		err = p.installer.InstallDependency(
			libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
			filepath.Join(p.depDir, "dotnet-sdk"),
		)
		if err != nil {
			return err
		}
	}

	aspNetCoreVersion, err := p.GetVersionFromDepsJSON("Microsoft.AspNetCore.App")
	if _, ok := err.(*libraryMissingError); err != nil && !ok {
		return err
	} else if ok {
		return nil
	}

	return p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspNetCoreVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
}

func (p *Project) installFrameworksAspNetCoreApp(frameworkName, frameworkVersion string, applyPatches *bool) error {
	aspNetCoreVersion, err := p.FindMatchingFrameworkVersion("dotnet-aspnetcore", frameworkVersion, applyPatches)
	if err != nil {
		return err
	}

	err = p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspNetCoreVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
	if err != nil {
		return err
	}

	aspNetCoreConfigJSON, err := ParseRuntimeConfig(filepath.Join(
		p.depDir,
		"dotnet-sdk",
		"shared",
		frameworkName,
		aspNetCoreVersion,
		fmt.Sprintf("%s.runtimeconfig.json", frameworkName),
	))
	if err != nil {
		return err
	}

	runtimeVersion, err := p.FindMatchingFrameworkVersion(
		"dotnet-runtime",
		aspNetCoreConfigJSON.RuntimeOptions.Framework.Version,
		aspNetCoreConfigJSON.RuntimeOptions.ApplyPatches,
	)
	if err != nil {
		return err
	}

	exists, err := p.isInstalled("Microsoft.NETCore.App", runtimeVersion)
	if err != nil {
		return err
	} else if exists {
		p.Log.Info("dotnet-runtime %s is already installed", runtimeVersion)
		return nil
	}

	return p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
}

func (p *Project) SourceInstallDotnetRuntime() error {
	mainPath, err := p.MainPath()
	if err != nil {
		return err
	}

	runtimeVersion, err := p.VersionFromProjFile(
		mainPath,
		"<RuntimeFrameworkVersion>(.*)</RuntimeFrameworkVersion>",
		"dotnet-runtime",
	)
	if err != nil {
		return err
	}

	if runtimeVersion == "" {
		runtimeVersions, err := p.versionsFromNugetPackages("dotnet-runtime", true)
		if err != nil {
			return err
		}

		for _, version := range runtimeVersions {
			if exists, err := p.isInstalled("Microsoft.NETCore.App", version); err != nil {
				return err
			} else if exists {
				p.Log.Info("dotnet-runtime %s is already installed", version)
				continue
			}

			err := p.installer.InstallDependency(
				libbuildpack.Dependency{Name: "dotnet-runtime", Version: version},
				filepath.Join(p.depDir, "dotnet-sdk"),
			)
			if err != nil {
				return err
			}
		}
	} else {
		err := p.installer.InstallDependency(
			libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
			filepath.Join(p.depDir, "dotnet-sdk"),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) SourceInstallDotnetAspNetCore() error {
	mainPath, err := p.MainPath()
	if err != nil {
		return err
	}

	aspnetcoreRegex := `"Microsoft.AspNetCore.App" Version="(.*)"`
	aspnetcoreVersion, err := p.VersionFromProjFile(mainPath, aspnetcoreRegex, "dotnet-aspnetcore")
	if err != nil {
		return err
	}

	if aspnetcoreVersion == "" {
		aspnetcoreVersions, err := p.versionsFromNugetPackages("dotnet-aspnetcore", true)
		if err != nil {
			return err
		}

		for _, version := range aspnetcoreVersions {
			err := p.installer.InstallDependency(
				libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: version},
				filepath.Join(p.depDir, "dotnet-sdk"),
			)
			if err != nil {
				return err
			}
		}
	} else {
		err := p.installer.InstallDependency(
			libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspnetcoreVersion},
			filepath.Join(p.depDir, "dotnet-sdk"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) isInstalled(library, version string) (bool, error) {
	libraryPath := filepath.Join(p.depDir, "dotnet-sdk", "shared", library, version)
	if exists, err := libbuildpack.FileExists(libraryPath); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	return false, nil
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
		return "", err
	} else if exists {
		if err := os.Chmod(filepath.Join(filepath.Join(publishedPath, projectPath)), 0755); err != nil {
			return "", err
		}
		return filepath.Join(runtimePath, projectPath), nil
	}

	if exists, err := libbuildpack.FileExists(filepath.Join(publishedPath, fmt.Sprintf("%s.dll", projectPath))); err != nil {
		return "", fmt.Errorf("checking if a %s.dll file exists: %v", projectPath, err)
	} else if exists {
		return fmt.Sprintf("%s.dll", filepath.Join(runtimePath, projectPath)), nil
	}
	return "", nil
}

func (p *Project) getAssemblyName(projectPath string) (string, error) {
	projFile, err := os.Open(projectPath)
	if err != nil {
		return "", err
	}
	defer projFile.Close()
	projBytes, err := ioutil.ReadAll(projFile)
	if err != nil {
		return "", err
	}

	proj := struct {
		PropertyGroup struct {
			AssemblyName string
		}
	}{}
	err = xml.Unmarshal(projBytes, &proj)
	if err != nil {
		return "", err
	}
	return proj.PropertyGroup.AssemblyName, nil
}

func (p *Project) getLatestPatch(name, version string) (string, error) {
	v := strings.Split(version, ".")
	v[2] = "x"
	versions := p.manifest.AllDependencyVersions(name)
	latestPatch, err := libbuildpack.FindMatchingVersion(strings.Join(v, "."), versions)
	if err != nil {
		return "", err
	}
	return latestPatch, nil
}

func ParseRuntimeConfig(runtimeConfigPath string) (ConfigJSON, error) {
	obj := ConfigJSON{}
	if err := libbuildpack.NewJSON().Load(runtimeConfigPath, &obj); err != nil {
		return ConfigJSON{}, err
	}
	return obj, nil
}
