package supply_test

import (
	"bytes"
	"dotnetcore/config"
	"dotnetcore/project"
	"dotnetcore/supply"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	gomock "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	var (
		err           error
		buildDir      string
		cacheDir      string
		depsDir       string
		depsIdx       string
		supplier      *supply.Supplier
		logger        *libbuildpack.Logger
		buffer        *bytes.Buffer
		mockCtrl      *gomock.Controller
		mockManifest  *MockManifest
		mockInstaller *MockInstaller
		mockCommand   *MockCommand
		installNode   func(string, string)
		installBower  func(string, string)
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "dotnetcore-buildpack.build.")
		Expect(err).To(BeNil())

		cacheDir, err = ioutil.TempDir("", "dotnetcore-buildpack.cache.")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "dotnetcore-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "9"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockInstaller = NewMockInstaller(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)

		args := []string{buildDir, cacheDir, depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})
		project := project.New(stager.BuildDir(), filepath.Join(depsDir, depsIdx), depsIdx)
		cfg := &config.Config{}

		supplier = &supply.Supplier{
			Stager:    stager,
			Manifest:  mockManifest,
			Installer: mockInstaller,
			Log:       logger,
			Command:   mockCommand,
			Project:   project,
			Config:    cfg,
		}

		installNode = func(dep, nodeDir string) {
			subDir := fmt.Sprintf("node-v%s-linux-x64", "6.12.0")
			err := os.MkdirAll(filepath.Join(nodeDir, subDir, "bin"), 0755)
			Expect(err).To(BeNil())
		}

		installBower = func(dep, bowerDir string) {
			subDir := fmt.Sprintf("bower-v%s-linux-x64", "1.8.2")
			err := os.MkdirAll(filepath.Join(bowerDir, subDir, "bin"), 0755)
			Expect(err).To(BeNil())
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(cacheDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("InstallBower", func() {
		var bowerInstallDir string
		BeforeEach(func() {
			bowerInstallDir = filepath.Join(depsDir, depsIdx, "node", "bin")
			Expect(os.MkdirAll(bowerInstallDir, 0755)).To(Succeed())
			csprojXml := `<Project Sdk="Microsoft.NET.Sdk.Web">
												<Target Name="PrepublishScript" BeforeTargets="PrepareForPublish">
													<Exec Command="bower install" />
												</Target>
											</Project>`
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "test_app.csproj"), []byte(csprojXml), 0644)).To(Succeed())
			Expect(err).To(BeNil())
			mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "bower", "-v").AnyTimes().Return(fmt.Errorf("error"))
			mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "npm", "install", "-g", gomock.Any()).AnyTimes().Return(nil)
		})
		Context("Not a published project and bower command necessary", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "npm", "-v").AnyTimes()
			})
			It("Installs bower", func() {
				mockInstaller.EXPECT().FetchDependency(libbuildpack.Dependency{Name: "bower", Version: "1.8.2"}, gomock.Any()).Return(nil)
				mockManifest.EXPECT().AllDependencyVersions("bower").AnyTimes().Return([]string{"1.8.2"})
				Expect(supplier.InstallBower()).To(Succeed())
			})
		})
		Context("It is a published project and bower command necessary", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "test_app.runtimeconfig.json"), []byte("any text"), 0644)).To(Succeed())
				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "npm", "-v").AnyTimes()
			})
			It("Does not install bower", func() {
				Expect(supplier.InstallBower()).To(Succeed())
			})
		})
		Context("NPM is NOT installed and bower command necessary", func() {
			It("Does not install bower", func() {
				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "npm", "-v").AnyTimes().Return(fmt.Errorf("error"))
				Expect(supplier.InstallBower()).ToNot(Succeed())
			})
		})
	})

	Describe("InstallNode", func() {
		var nodeInstallDir string
		var nodeTmpDir string
		var csprojXml string
		BeforeEach(func() {
			nodeInstallDir = filepath.Join(depsDir, depsIdx, "node")
			nodeTmpDir, err = ioutil.TempDir("", "dotnetcore-buildpack.tmp")
			Expect(err).To(BeNil())
			csprojXml = `<Project Sdk="Microsoft.NET.Sdk.Web">
												<Target Name="PrepublishScript" BeforeTargets="PrepareForPublish">
													<Exec Command="npm install" />
													<Exec Command="bower install" />
												</Target>
											</Project>`
		})

		AfterEach(func() {
			Expect(os.RemoveAll(nodeTmpDir)).To(Succeed())
		})

		Context("Node is not installed", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "node", "-v").AnyTimes().Return(fmt.Errorf("error"))
			})

			Context("Install node environment variable is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("INSTALL_NODE", "true")).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.Unsetenv("INSTALL_NODE")).To(Succeed())
				})

				It("Installs node", func() {
					mockInstaller.EXPECT().InstallOnlyVersion("node", gomock.Any()).Do(installNode).Return(nil)
					mockManifest.EXPECT().AllDependencyVersions("node").Return([]string{"6.12.0"})
					Expect(supplier.InstallNode()).To(Succeed())
				})
			})

			Context("Not a published project and bower/npm commands necessary", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "test_app.csproj"), []byte(csprojXml), 0644)).To(Succeed())
				})

				It("Installs node", func() {
					mockInstaller.EXPECT().InstallOnlyVersion("node", gomock.Any()).Do(installNode).Return(nil)
					mockManifest.EXPECT().AllDependencyVersions("node").AnyTimes().Return([]string{"6.12.0"})
					Expect(supplier.InstallNode()).To(Succeed())
				})
			})

			Context("It is a published project and bower/npm commands necessary", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "test_app.csproj"), []byte(csprojXml), 0644)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "test_app.runtimeconfig.json"), []byte("any text"), 0644)).To(Succeed())
				})

				It("Does not install node", func() {
					Expect(supplier.InstallNode()).To(Succeed())
				})
			})
		})

		Context("Node is installed", func() {
			BeforeEach(func() {
				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "node", "-v").AnyTimes()
			})

			It("Does not re-install node", func() {
				mockInstaller.EXPECT().InstallOnlyVersion("node", nodeTmpDir).Times(0)
				Expect(supplier.InstallNode()).To(Succeed())
			})
		})
	})

	Describe("InstallDotnet", func() {
		var defaultDep = libbuildpack.Dependency{Name: "dotnet", Version: "3.4.5"}

		Context("global.json", func() {
			Context("with sdk/version", func() {
				Context("that is in the buildpack", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "global.json"), []byte(`{"sdk": {"version": "6.7.8"}}`), 0644)).To(Succeed())
						mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{"6.7.8"})
					})

					It("installs the requested version", func() {
						dep := libbuildpack.Dependency{Name: "dotnet", Version: "6.7.8"}
						mockInstaller.EXPECT().InstallDependency(dep, filepath.Join(depsDir, depsIdx, "dotnet"))

						Expect(supplier.InstallDotnet()).To(Succeed())
					})
				})

				Context("that is missing, but matches existing version lines", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "global.json"), []byte(`{"sdk": {"version": "1.2.3"}}`), 0644)).To(Succeed())
						mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{"1.1.1", "1.2.5", "1.2.6", "1.3.7"})
					})

					It("installs the latest of the same version line", func() {
						dep := libbuildpack.Dependency{Name: "dotnet", Version: "1.2.6"}
						mockInstaller.EXPECT().InstallDependency(dep, filepath.Join(depsDir, depsIdx, "dotnet"))

						Expect(supplier.InstallDotnet()).To(Succeed())
					})
				})

				Context("that is missing, and does not match existing version lines", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(buildDir, "global.json"), []byte(`{"sdk": {"version": "1.2.3"}}`), 0644)).To(Succeed())
						mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{"1.1.1", "1.3.7"})
					})

					It("installs the default version", func() {
						mockManifest.EXPECT().DefaultVersion("dotnet").Return(defaultDep, nil)
						mockInstaller.EXPECT().InstallDependency(defaultDep, filepath.Join(depsDir, depsIdx, "dotnet"))

						Expect(supplier.InstallDotnet()).To(Succeed())
					})
				})
			})

			Context("without sdk/version", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "global.json"), []byte(`{}`), 0644)).To(Succeed())
				})

				It("installs the default version", func() {
					mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{})
					mockManifest.EXPECT().DefaultVersion("dotnet").Return(defaultDep, nil)
					mockInstaller.EXPECT().InstallDependency(defaultDep, filepath.Join(depsDir, depsIdx, "dotnet"))

					Expect(supplier.InstallDotnet()).To(Succeed())
				})
			})

			Context("with malformed sdk/version", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "global.json"), []byte(`hi mom`), 0644)).To(Succeed())
				})

				It("installs an error", func() {
					mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{})
					Expect(supplier.InstallDotnet()).ToNot(Succeed())
				})
			})

		})

		Context("fsproj", func() {
			BeforeEach(func() {
				Expect(os.Mkdir(filepath.Join(buildDir, "inner"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "inner", "example.fsproj"), []byte(""), 0644)).To(Succeed())
			})

			It("returns the fsharp compatible dotnet version", func() {
				mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{"1.0.4", "1.1.6", "1.1.7", "1.1.5", "2.0.0"})

				fSharpDep := libbuildpack.Dependency{Name: "dotnet", Version: "1.1.7"}
				mockInstaller.EXPECT().InstallDependency(fSharpDep, filepath.Join(depsDir, depsIdx, "dotnet"))

				Expect(supplier.InstallDotnet()).To(Succeed())
			})
		})

		Context("no known version", func() {
			It("returns the default version", func() {
				mockManifest.EXPECT().AllDependencyVersions("dotnet").Return([]string{})
				mockManifest.EXPECT().DefaultVersion("dotnet").Return(defaultDep, nil)
				mockInstaller.EXPECT().InstallDependency(defaultDep, filepath.Join(depsDir, depsIdx, "dotnet"))

				Expect(supplier.InstallDotnet()).To(Succeed())
			})
		})
	})
})
