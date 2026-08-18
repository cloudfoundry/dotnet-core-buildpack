package finalize_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/config"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/finalize"
	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/project"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
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
		DeferCleanup(os.RemoveAll, buildDir)

		depsDir, err = os.MkdirTemp("", "dotnet-core-buildpack.deps.")
		Expect(err).To(BeNil())
		DeferCleanup(os.RemoveAll, depsDir)

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

		stackRID = "linux-x64"
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

	Describe("WriteProfileD", func() {
		var scriptPath string

		BeforeEach(func() {
			scriptPath = filepath.Join(depsDir, depsIdx, "profile.d", "startup.sh")
		})

		It("writes the ASPNETCORE_URLS and DOTNET_ROOT exports", func() {
			Expect(finalizer.WriteProfileD()).To(Succeed())

			contents, err := os.ReadFile(scriptPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`export ASPNETCORE_URLS="${ASPNETCORE_URLS:-http://0.0.0.0:${PORT}}"`))
			Expect(string(contents)).To(ContainSubstring("export DOTNET_ROOT=" + filepath.Join("/home", "vcap", "deps", depsIdx, "dotnet-sdk")))
		})

		It("writes the DOTNET_GCHeapHardLimit calculation block", func() {
			Expect(finalizer.WriteProfileD()).To(Succeed())

			contents, err := os.ReadFile(scriptPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring("MEMORY_LIMIT"))
			Expect(string(contents)).To(ContainSubstring("DOTNET_GCHeapHardLimit"))
			Expect(string(contents)).To(ContainSubstring("DOTNET_GCHeapHardLimitPercent"))
		})

		runScript := func(env []string) string {
			Expect(finalizer.WriteProfileD()).To(Succeed())

			cmd := exec.Command("bash", "-c", "source "+scriptPath+` && printf '%s' "$DOTNET_GCHeapHardLimit"`)
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			return string(output)
		}

		Context("MEMORY_LIMIT is set to 1024m and no overrides are present", func() {
			It("exports DOTNET_GCHeapHardLimit at 75 percent of the container limit", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m"})).To(Equal("0x30000000"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimitPercent is set to 50", func() {
			It("exports DOTNET_GCHeapHardLimit at the requested percentage", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=512m", "DOTNET_GCHeapHardLimitPercent=50"})).To(Equal("0x10000000"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimitPercent is invalid", func() {
			It("falls back to the 75 percent default", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimitPercent=not-a-number"})).To(Equal("0x30000000"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimitPercent is out of range", func() {
			It("falls back to the 75 percent default", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimitPercent=150"})).To(Equal("0x30000000"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimitPercent has a leading zero that bash would otherwise treat as invalid octal", func() {
			It("does not abort and falls back gracefully instead of erroring", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimitPercent=08"})).To(Equal("0x51eb851"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimitPercent has a leading zero that bash would otherwise treat as valid octal", func() {
			It("interprets the value as decimal, not octal", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimitPercent=050"})).To(Equal("0x20000000"))
			})
		})

		Context("MEMORY_LIMIT is set to 0m", func() {
			It("does not export DOTNET_GCHeapHardLimit", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=0m"})).To(Equal(""))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimit is already set within range", func() {
			It("respects the user-provided value", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimit=0x10000000"})).To(Equal("0x10000000"))
			})
		})

		Context("MEMORY_LIMIT is set and DOTNET_GCHeapHardLimit exceeds it", func() {
			It("clamps to the container memory limit and logs a warning", func() {
				Expect(finalizer.WriteProfileD()).To(Succeed())

				cmd := exec.Command("bash", "-c", "source "+scriptPath+` && printf '%s' "$DOTNET_GCHeapHardLimit"`)
				cmd.Env = []string{"MEMORY_LIMIT=1024m", "DOTNET_GCHeapHardLimit=0x100000000"}
				output, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("WARNING: DOTNET_GCHeapHardLimit (4294967296 bytes) exceeds MEMORY_LIMIT (1073741824 bytes)"))
				Expect(string(output)).To(HaveSuffix("0x40000000"))
			})
		})

		Context("MEMORY_LIMIT is not set", func() {
			It("does not export DOTNET_GCHeapHardLimit", func() {
				Expect(runScript([]string{})).To(Equal(""))
			})
		})

		Context("MEMORY_LIMIT does not match the expected <int>m format", func() {
			It("does not export DOTNET_GCHeapHardLimit", func() {
				Expect(runScript([]string{"MEMORY_LIMIT=512M"})).To(Equal(""))
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
