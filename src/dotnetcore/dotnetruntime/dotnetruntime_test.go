package dotnetruntime_test

import (
	"bytes"
	"dotnetcore/dotnetruntime"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=dotnetruntime.go --destination=mocks_dotnetruntime_test.go --package=dotnetruntime_test

var _ = Describe("Dotnetruntime", func() {
	var (
		err           error
		depDir        string
		buildDir      string
		subject       *dotnetruntime.DotnetRuntime
		mockCtrl      *gomock.Controller
		mockInstaller *MockInstaller
		mockManifest  *MockManifest
		manifest      *libbuildpack.Manifest
		buffer        *bytes.Buffer
		logger        *libbuildpack.Logger
	)

	BeforeEach(func() {
		depDir, err = ioutil.TempDir("", "dotnetcore-buildpack.deps.")
		buildDir, err = ioutil.TempDir("", "dotnetcore-buildpack.build.")
		Expect(err).To(BeNil())

		mockCtrl = gomock.NewController(GinkgoT())
		mockInstaller = NewMockInstaller(mockCtrl)
		mockManifest = NewMockManifest(mockCtrl)

		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		Expect(ioutil.WriteFile(filepath.Join(buildDir, "manifest.yml"), []byte("---"), 0644)).To(Succeed())
		manifest, err = libbuildpack.NewManifest(buildDir, logger, time.Now())
		Expect(err).To(BeNil())
		Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"), []byte("---"), 0644)).To(Succeed())

		subject = dotnetruntime.New(depDir, buildDir, mockInstaller, mockManifest, logger)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.RemoveAll(depDir)).To(Succeed())
		Expect(os.RemoveAll(buildDir)).To(Succeed())
	})

	Describe("Install", func() {
		Context("Versions installed == [1.2.3, 4.5.6]", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(depDir, "dotnet-sdk", "shared", "Microsoft.NETCore.App", "1.2.3"), 0755)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(depDir, "dotnet-sdk", "shared", "Microsoft.NETCore.App", "4.5.6"), 0755)).To(Succeed())
			})

			Context("when required version is discovered via .runtimeconfig.json", func() {
				Context("Versions required == [4.5.6]", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.runtimeconfig.json"),
							[]byte(`{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "4.5.6" }, "applyPatches": false } }`), 0644)).To(Succeed())
					})

					It("does not install the runtime again", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "4.5.6"}, gomock.Any()).Times(0)
						Expect(subject.Install(filepath.Join(buildDir, "foo"))).To(Succeed())
					})
				})

				Context("Versions required == [7.8.9]", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.runtimeconfig.json"),
							[]byte(`{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "7.8.9" }, "applyPatches": false } }`), 0644)).To(Succeed())
					})

					It("installs the additional runtime", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "7.8.9"}, filepath.Join(depDir, "dotnet-sdk"))
						Expect(subject.Install(filepath.Join(buildDir, "foo.csproj"))).To(Succeed())
					})
				})
			})

			Context("when required versions are discovered via restored packages", func() {
				Context("Versions required == [4.5.6]", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(depDir, ".nuget", "packages", "microsoft.netcore.app", "4.5.6"), 0755)).To(Succeed())
					})

					It("does not install the runtime again", func() {
						mockManifest.EXPECT().AllDependencyVersions("dotnet-runtime").Return([]string{"4.5.6"})
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "4.5.6"}, gomock.Any()).Times(0)
						Expect(subject.Install(filepath.Join(buildDir, "foo.csproj"))).To(Succeed())
					})
				})

				Context("Versions required == [7.8.9]", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(depDir, ".nuget", "packages", "microsoft.netcore.app", "7.8.9"), 0755)).To(Succeed())
					})

					It("installs the additional runtime", func() {
						mockManifest.EXPECT().AllDependencyVersions("dotnet-runtime").Return([]string{"7.8.9"})
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "7.8.9"}, filepath.Join(depDir, "dotnet-sdk"))
						Expect(subject.Install(filepath.Join(buildDir, "foo.csproj"))).To(Succeed())
					})
				})
			})
		})
	})
})
