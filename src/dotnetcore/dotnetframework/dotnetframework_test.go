package dotnetframework_test

import (
	"bytes"
	"dotnetcore/dotnetframework"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=dotnetframework.go --destination=mocks_dotnetframework_test.go --package=dotnetframework_test

var _ = Describe("Dotnetframework", func() {
	var (
		err           error
		depDir        string
		buildDir      string
		subject       *dotnetframework.DotnetFramework
		mockCtrl      *gomock.Controller
		mockInstaller *MockInstaller
		buffer        *bytes.Buffer
		logger        *libbuildpack.Logger
	)

	BeforeEach(func() {
		depDir, err = ioutil.TempDir("", "dotnetcore-buildpack.deps.")
		buildDir, err = ioutil.TempDir("", "dotnetcore-buildpack.build.")
		Expect(err).To(BeNil())

		mockCtrl = gomock.NewController(GinkgoT())
		mockInstaller = NewMockInstaller(mockCtrl)

		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		subject = dotnetframework.New(depDir, buildDir, mockInstaller, logger)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.RemoveAll(depDir)).To(Succeed())
		Expect(os.RemoveAll(buildDir)).To(Succeed())
	})

	Describe("Install", func() {
		Context("Versions installed == [1.2.3, 4.5.6]", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(depDir, "dotnet", "shared", "Microsoft.NETCore.App", "1.2.3"), 0755)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(depDir, "dotnet", "shared", "Microsoft.NETCore.App", "4.5.6"), 0755)).To(Succeed())
			})

			Context("when installed version contains metadata in its semver", func() {
				BeforeEach(func() {
					Expect(os.MkdirAll(filepath.Join(depDir, ".nuget", "packages", "microsoft.netcore.app", "7.8.9"), 0755)).To(Succeed())
				})

				It("symlinks to plain semver", func() {
					mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-framework", Version: "7.8.9"}, filepath.Join(depDir, "dotnet")).Do(func(libbuildpack.Dependency, string) {
						Expect(os.MkdirAll(filepath.Join(depDir, "dotnet", "shared", "Microsoft.NETCore.App", "7.8.9-rtm-foo"), 0755)).To(Succeed())
					})
					Expect(subject.Install()).To(Succeed())
					dest, err := os.Readlink(filepath.Join(depDir, "dotnet", "shared", "Microsoft.NETCore.App", "7.8.9"))
					Expect(err).To(BeNil())
					Expect(dest).To(Equal(filepath.Join(depDir, "dotnet", "shared", "Microsoft.NETCore.App", "7.8.9-rtm-foo")))
				})
			})

			Context("when required version is discovered via .runtimeconfig.json", func() {
				Context("Versions required == [4.5.6]", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.runtimeconfig.json"),
							[]byte(`{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "4.5.6" } } }`), 0644)).To(Succeed())
					})

					It("does not install the framework again", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-framework", Version: "4.5.6"}, gomock.Any()).Times(0)
						Expect(subject.Install()).To(Succeed())
					})
				})

				Context("Versions required == [7.8.9]", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.runtimeconfig.json"),
							[]byte(`{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "7.8.9" } } }`), 0644)).To(Succeed())
					})

					It("installs the additional framework", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-framework", Version: "7.8.9"}, filepath.Join(depDir, "dotnet"))
						Expect(subject.Install()).To(Succeed())
					})
				})
			})

			Context("when required versions are discovered via restored packages", func() {
				Context("Versions required == [4.5.6]", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(depDir, ".nuget", "packages", "microsoft.netcore.app", "4.5.6"), 0755)).To(Succeed())
					})

					It("does not install the framework again", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-framework", Version: "4.5.6"}, gomock.Any()).Times(0)
						Expect(subject.Install()).To(Succeed())
					})
				})

				Context("Versions required == [7.8.9]", func() {
					BeforeEach(func() {
						Expect(os.MkdirAll(filepath.Join(depDir, ".nuget", "packages", "microsoft.netcore.app", "7.8.9"), 0755)).To(Succeed())
					})

					It("installs the additional framework", func() {
						mockInstaller.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "dotnet-framework", Version: "7.8.9"}, filepath.Join(depDir, "dotnet"))
						Expect(subject.Install()).To(Succeed())
					})
				})
			})
		})
	})
})
