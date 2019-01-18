package project

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gravityblast/go-jsmin"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/go-ini/ini"
)

type CSProj struct {
	PropertyGroup struct {
		TargetFramework         string `xml:"TargetFramework"`
		RuntimeFrameworkVersion string `xml:"RuntimeFrameworkVersion"`
		AssemblyName            string `xml:"AssemblyName"`
	}
	ItemGroups []struct {
		PackageReferences []struct {
			Include string `xml:"Include,attr"`
			Version string `xml:"Version,attr"`
		} `xml:"PackageReference"`
	} `xml:"ItemGroup"`
}

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
		proj, err := p.parseProj()
		if err != nil {
			return "", err
		}

		if proj.PropertyGroup.AssemblyName != "" {
			projectPath = projRe.ReplaceAllString(proj.PropertyGroup.AssemblyName, "")
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

func (p *Project) UsesLibrary(library string) (bool, error) {
	published, err := p.IsPublished()
	if err != nil {
		return false, err
	}

	if published {
		_, err := p.GetVersionFromDepsJSON(library)
		if _, libMissing := err.(*libraryMissingError); err != nil && !libMissing {
			return false, err
		} else if libMissing {
			return false, nil
		}
		return true, nil
	} else {
		proj, err := p.parseProj()
		if err != nil {
			return false, err
		}

		for _, ig := range proj.ItemGroups {
			for _, pr := range ig.PackageReferences {
				if pr.Include == library {
					return true, nil
				}
			}
		}
	}

	return false, nil
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
		runtimeJSON, err := parseRuntimeConfig(path)
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

	runtimeConfig, err := parseRuntimeConfig(path)
	if err != nil {
		return err
	}

	frameworkName := runtimeConfig.RuntimeOptions.Framework.Name
	frameworkVersion := runtimeConfig.RuntimeOptions.Framework.Version
	applyPatches := runtimeConfig.RuntimeOptions.ApplyPatches

	if frameworkName == "Microsoft.NETCore.App" {
		return p.fddInstallFrameworksNETCoreApp(frameworkName, frameworkVersion, applyPatches)
	} else if frameworkName == "Microsoft.AspNetCore.All" || frameworkName == "Microsoft.AspNetCore.App" {
		return p.fddInstallFrameworksAspNetCoreApp(frameworkName, frameworkVersion, applyPatches)
	}

	return fmt.Errorf("invalid framework [%s] specified in application runtime config file", frameworkName)
}

func (p *Project) SourceInstallDotnetRuntime() error {
	proj, err := p.parseProj()
	if err != nil {
		return err
	}

	runtimeVersion := proj.PropertyGroup.RuntimeFrameworkVersion
	if runtimeVersion != "" {
		matches := regexp.MustCompile(`\d\.\d\.\d`).FindStringSubmatch(runtimeVersion)
		if len(matches) != 1 {
			runtimeVersion, err = p.getLatestPatch("dotnet-runtime", runtimeVersion)
			if err != nil {
				return err
			}
		}
	} else {
		matches := regexp.MustCompile(`netcoreapp(.*)`).FindStringSubmatch(proj.PropertyGroup.TargetFramework)
		if len(matches) == 2 {
			runtimeVersionMinor := matches[1]
			runtimeVersion, err = p.getLatestPatch("dotnet-runtime", runtimeVersionMinor)
			if err != nil {
				return err
			}
		} else {
			return errors.New("could not find a version of dotnet-runtime to install")
		}
	}

	p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
	return nil
}

func (p *Project) SourceInstallDotnetAspNetCore() error {
	proj, err := p.parseProj()
	if err != nil {
		return err
	}

	aspnetcoreVersion := ""
	for _, ig := range proj.ItemGroups {
		for _, pr := range ig.PackageReferences {
			if pr.Include == "Microsoft.AspNetCore.App" || pr.Include == "Microsoft.AspNetCore.All" {
				aspnetcoreVersion = pr.Version
			}
		}
	}

	if aspnetcoreVersion == "" {
		aspnetcoreVersions, err := p.versionsFromNugetPackages("dotnet-aspnetcore", true)
		if err != nil {
			return err
		}

		for _, version := range aspnetcoreVersions {
			err := p.installAspNetCoreDependency(version, false)
			if err != nil {
				return err
			}
		}
	} else {
		err = p.installAspNetCoreDependency(aspnetcoreVersion, true)
		if err != nil {
			return err
		}
	}

	return nil
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

func (p *Project) installAspNetCoreDependency(version string, latestPatch bool) error {
	semverObj, err := semver.Parse(version)
	if err != nil {
		return err
	}

	if semverObj.Major <= 2 && semverObj.Minor < 1 {
		return nil
	}

	if latestPatch {
		version, err = p.getLatestPatch("dotnet-aspnetcore", version)
		if err != nil {
			return err
		}
	}

	return p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: version},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
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

func (p *Project) getLatestPatch(name, version string) (string, error) {
	v := strings.Split(version, ".")
	length := len(v)
	if length == 0 {
		return "", fmt.Errorf("could not find latest patch of %s: version %s not found", name, version)
	}

	if length <= 2 {
		v = append(v, "x")
	} else if length == 3 {
		v[2] = "x"
	} else {
		return "", fmt.Errorf("could not find latest patch of %s: could not parse version %s", name, version)
	}

	versions := p.manifest.AllDependencyVersions(name)
	latestPatch, err := libbuildpack.FindMatchingVersion(strings.Join(v, "."), versions)
	if err != nil {
		return "", err
	}
	return latestPatch, nil
}

func (p *Project) fddInstallFrameworksNETCoreApp(frameworkName, frameworkVersion string, applyPatches *bool) error {
	runtimeVersion, err := p.FindMatchingFrameworkVersion("dotnet-runtime", frameworkVersion, applyPatches)
	if err != nil {
		return err
	}

	if err = p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	); err != nil {
		return err
	}

	aspNetCoreVersion, err := p.GetVersionFromDepsJSON("Microsoft.AspNetCore.App")
	if _, ok := err.(*libraryMissingError); err != nil && !ok {
		return err
	} else if ok {
		return nil
	}

	return p.installAspNetCoreDependency(aspNetCoreVersion, false)
}

func (p *Project) fddInstallFrameworksAspNetCoreApp(frameworkName, frameworkVersion string, applyPatches *bool) error {
	aspNetCoreVersion, err := p.FindMatchingFrameworkVersion("dotnet-aspnetcore", frameworkVersion, applyPatches)
	if err != nil {
		return err
	}

	if err = p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspNetCoreVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	); err != nil {
		return err
	}

	aspNetCoreConfigJSON, err := parseRuntimeConfig(filepath.Join(
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

	return p.installer.InstallDependency(
		libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion},
		filepath.Join(p.depDir, "dotnet-sdk"),
	)
}

func (p *Project) parseProj() (CSProj, error) {
	mainPath, err := p.MainPath()
	if err != nil {
		return CSProj{}, err
	}

	projFile, err := os.Open(mainPath)
	if err != nil {
		return CSProj{}, err
	}
	defer projFile.Close()
	projBytes, err := ioutil.ReadAll(projFile)
	if err != nil {
		return CSProj{}, err
	}
	obj := CSProj{}

	if err := xml.Unmarshal(projBytes, &obj); err != nil {
		return CSProj{}, err
	}
	return obj, nil
}

func sanitizeJsonConfig(runtimeConfigPath string) ([]byte, error) {
	input, err := os.Open(runtimeConfigPath)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	output := &bytes.Buffer{}
	if err := jsmin.Min(input, output); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func parseRuntimeConfig(runtimeConfigPath string) (ConfigJSON, error) {
	obj := ConfigJSON{}

	buf, err := sanitizeJsonConfig(runtimeConfigPath)
	if err != nil {
		return obj, err
	}

	if err := json.Unmarshal(buf, &obj); err != nil {
		return obj, err
	}

	return obj, nil
}

type libraryMissingError struct {
	s string
}

func (e *libraryMissingError) Error() string {
	return e.s
}
