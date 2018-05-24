package main

import (
	"dotnetcore/config"
	"dotnetcore/dotnetframework"
	"dotnetcore/finalize"
	"dotnetcore/project"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	// _ "dotnetcore/hooks"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

func main() {
	logfile, err := ioutil.TempFile("", "cloudfoundry.dotnetcore-buildpack.finalize")
	defer logfile.Close()
	if err != nil {
		logger := libbuildpack.NewLogger(os.Stdout)
		logger.Error("Unable to create log file: %s", err.Error())
		os.Exit(8)
	}

	stdout := io.MultiWriter(os.Stdout, logfile)
	logger := libbuildpack.NewLogger(stdout)

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(9)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(10)
	}

	stager := libbuildpack.NewStager(os.Args[1:], logger, manifest)

	if err = manifest.ApplyOverride(stager.DepsDir()); err != nil {
		logger.Error("Unable to apply override.yml files: %s", err)
		os.Exit(17)
	}

	if err := stager.SetStagingEnvironment(); err != nil {
		logger.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(11)
	}

	configYml := struct {
		Config config.Config `yaml:"config"`
	}{}
	if err := libbuildpack.NewYAML().Load(filepath.Join(stager.DepDir(), "config.yml"), &configYml); err != nil {
		logger.Error("Unable to read supply time config.yml: %s", err.Error())
		os.Exit(15)
	}

	dotnetframework := dotnetframework.New(stager.DepDir(), stager.BuildDir(), libbuildpack.NewInstaller(manifest), logger)
	f := finalize.Finalizer{
		Stager:          stager,
		Log:             logger,
		Command:         &libbuildpack.Command{},
		DotnetFramework: dotnetframework,
		Config:          &configYml.Config,
		Project:         project.New(stager.BuildDir(), stager.DepDir(), stager.DepsIdx()),
	}

	if err := finalize.Run(&f); err != nil {
		os.Exit(12)
	}

	if err := libbuildpack.RunAfterCompile(stager); err != nil {
		logger.Error("After Compile: %s", err.Error())
		os.Exit(13)
	}

	if err := stager.SetLaunchEnvironment(); err != nil {
		logger.Error("Unable to setup launch environment: %s", err.Error())
		os.Exit(14)
	}

	stager.StagingComplete()
}
