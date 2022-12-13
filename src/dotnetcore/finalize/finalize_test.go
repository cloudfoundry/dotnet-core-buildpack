package finalize_test

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/config"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/finalize"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/project"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=finalize.go --destination=mocks_finalize_test.go --package=finalize_test

var _ = Describe("Finalize", func() {
	var (
		err         error
		buildDir    string
		depsDir     string
		depsIdx     string
		finalizer   *finalize.Finalizer
		logger      *libbuildpack.Logger
		buffer      *bytes.Buffer
		mockCtrl    *gomock.Controller
		mockCommand *MockCommand
		stackRID    string
	)

	BeforeEach(func() {
		buildDir, err = os.MkdirTemp("", "dotnet-core-buildpack.build.")
		Expect(err).To(BeNil())

		depsDir, err = os.MkdirTemp("", "dotnet-core-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "9"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockCommand = NewMockCommand(mockCtrl)

		args := []string{buildDir, "", depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})
		project := project.New(stager.BuildDir(), filepath.Join(depsDir, depsIdx), depsIdx, &libbuildpack.Manifest{}, libbuildpack.NewInstaller(&libbuildpack.Manifest{}), logger)
		cfg := &config.Config{}

		finalizer = &finalize.Finalizer{
			Stager:  stager,
			Command: mockCommand,
			Log:     logger,
			Project: project,
			Config:  cfg,
		}

		stackRID = "ubuntu.18.04-x64"
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("DotnetPublish", func() {
		Context("The project is already published", func() {
			BeforeEach(func() {
				Expect(os.WriteFile(filepath.Join(buildDir, "test_app.runtimeconfig.json"), []byte("any text"), 0644)).To(Succeed())
			})
			It("Does not run dotnet publish", func() {
				Expect(finalizer.DotnetPublish(stackRID)).To(Succeed())
			})
		})
		Context("The project is NOT already published", func() {
			It("Runs dotnet publish", func() {
				mockCommand.EXPECT().Run(gomock.Any())
				Expect(finalizer.DotnetPublish(stackRID)).To(Succeed())
			})
		})
	})

	Describe("CleanStagingArea", func() {
		Context(`The .nuget directory exists with a symlink to it`, func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "bin"), 0755)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "lib"), 0755)).To(Succeed())
				for _, name := range []string{
					".nuget/fileA.txt",
					".nuget/fileB.txt",
					"other/file.txt",
				} {
					Expect(os.MkdirAll(filepath.Dir(filepath.Join(depsDir, depsIdx, name)), 0755)).To(Succeed())
					Expect(os.WriteFile(filepath.Join(depsDir, depsIdx, name), []byte(""), 0644)).To(Succeed())
					Expect(os.Symlink(filepath.Join(depsDir, depsIdx, name), filepath.Join(depsDir, depsIdx, "bin", filepath.Base(name)))).To(Succeed())
					Expect(os.Symlink(filepath.Join(depsDir, depsIdx, name), filepath.Join(depsDir, depsIdx, "lib", filepath.Base(name)))).To(Succeed())
				}
			})

			It("deletes .nuget directory", func() {
				Expect(finalizer.CleanStagingArea()).To(Succeed())

				Expect(filepath.Join(depsDir, depsIdx, ".nuget")).ToNot(BeADirectory())
				Expect(filepath.Join(depsDir, depsIdx, "other", "file.txt")).To(BeARegularFile())
			})

			It("deletes symlinks to .nuget directory from bin directory", func() {
				Expect(finalizer.CleanStagingArea()).To(Succeed())

				files, err := filepath.Glob(filepath.Join(depsDir, depsIdx, "bin", "*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(files).To(Equal([]string{filepath.Join(depsDir, depsIdx, "bin", "file.txt")}))
			})

			It("deletes symlinks to .nuget directory from lib directory", func() {
				Expect(finalizer.CleanStagingArea()).To(Succeed())

				files, err := filepath.Glob(filepath.Join(depsDir, depsIdx, "lib", "*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(files).To(Equal([]string{filepath.Join(depsDir, depsIdx, "lib", "file.txt")}))
			})
		})
	})
})
