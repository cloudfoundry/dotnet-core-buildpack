package project_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/dotnet-core-buildpack/src/dotnetcore/project"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=project.go --destination=mocks_project_test.go --package=project_test

var _ = Describe("Project", func() {
	var (
		err           error
		buildDir      string
		depsDir       string
		depsIdx       string
		depsPath      string
		subject       *project.Project
		mockCtrl      *gomock.Controller
		mockManifest  *MockManifest
		mockInstaller *MockInstaller
		logger        *libbuildpack.Logger
		buffer        *bytes.Buffer
	)

	createRuntimeConfig := func(dep, version string) {
		content := `{ "runtimeOptions": { "framework": { "name": "%s", "version": "%s" }, "applyPatches": false } }`
		Expect(ioutil.WriteFile(filepath.Join(buildDir, "test.runtimeconfig.json"), []byte(fmt.Sprintf(content, dep, version)), 0644)).To(Succeed())
	}

	createDepsJSON := func(dep, version string, emptyContent bool) {
		if emptyContent {
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "test.deps.json"), []byte(`{ "libraries": {} }`), 0644)).To(Succeed())
		} else {
			content := `{ "libraries": { "%s/%s": { "name": "Microsoft.NETCore.App", "version": "4.5.6" } } }`
			Expect(ioutil.WriteFile(filepath.Join(buildDir, "test.deps.json"), []byte(fmt.Sprintf(content, dep, version)), 0644)).To(Succeed())
		}
	}

	installRuntimeConfig := func(dep, aspNetCoreVersion, runtimeVersion string) {
		content := `{ "runtimeOptions": { "framework": { "name": "Microsoft.NETCore.App", "version": "%s" }, "applyPatches": false } }`
		path := fmt.Sprintf(filepath.Join(depsDir, depsIdx, "dotnet-sdk", "shared", "%s", "%s", "%s.runtimeconfig.json"), dep, aspNetCoreVersion, dep)
		Expect(os.MkdirAll(filepath.Dir(path), 0777)).To(Succeed())
		Expect(ioutil.WriteFile(path, []byte(fmt.Sprintf(content, runtimeVersion)), 0666)).To(Succeed())
	}

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "dotnet-core-buildpack.build.")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "dotnetcore-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "9"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		depsPath = filepath.Join(depsDir, depsIdx, "dotnet-sdk")

		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockInstaller = NewMockInstaller(mockCtrl)

		subject = project.New(buildDir, filepath.Join(depsDir, depsIdx), depsIdx, mockManifest, mockInstaller, logger)
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
		Expect(err).To(BeNil())
	})

	Describe("StartCommand", func() {
		Context("The project is published", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "fred.runtimeconfig.json"), []byte(""), 0644)).To(Succeed())
			})

			Context("An executable for the project exists", func() {
				//before: make a 'fred' executable.
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "fred"), []byte(""), 0755)).To(Succeed())
				})

				It("returns ${HOME}/project", func() {
					startCmd, err := subject.StartCommand()
					Expect(err).To(BeNil())
					Expect(startCmd).To(Equal(filepath.Join("${HOME}", "fred")))
				})
			})

			Context("An executable for the project does NOT exist, but a dll does", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "fred.dll"), []byte(""), 0755)).To(Succeed())
				})

				It("returns ${HOME}/project.dll", func() {
					startCmd, err := subject.StartCommand()
					Expect(err).To(BeNil())
					Expect(startCmd).To(Equal(filepath.Join("${HOME}", "fred.dll")))
				})
			})
			Context("An executable for the project does NOT exist, and neither does a dll", func() {
				It("returns an empty string", func() {
					startCmd, err := subject.StartCommand()
					Expect(err).To(BeNil())
					Expect(startCmd).To(Equal(""))
				})
			})
		})

		Context("The project is NOT published", func() {
			Context("The csproj file does not have an AssemblyName tag", func() {
				BeforeEach(func() {
					Expect(os.MkdirAll(filepath.Join(buildDir, "subdir"), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "subdir", "fred.csproj"), []byte("<Project></Project>"), 0644)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "dotnet_publish"), 0755)).To(Succeed())
				})

				Context("An executable for the project exists", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "dotnet_publish", "fred"), []byte(""), 0755)).To(Succeed())
					})

					It("returns a path to the project executable", func() {
						startCmd, err := subject.StartCommand()
						Expect(err).To(BeNil())
						Expect(startCmd).To(Equal(filepath.Join("${DEPS_DIR}", depsIdx, "dotnet_publish", "fred")))
					})
				})

				Context("An executable for the project does NOT exist, but a dll does", func() {
					BeforeEach(func() {
						Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "dotnet_publish", "fred.dll"), []byte(""), 0755)).To(Succeed())
					})

					It("returns the path to the project.dll", func() {
						startCmd, err := subject.StartCommand()
						Expect(err).To(BeNil())
						Expect(startCmd).To(Equal(filepath.Join("${DEPS_DIR}", depsIdx, "dotnet_publish", "fred.dll")))
					})

				})

				Context("An executable for the project does NOT exist, and neither does a dll", func() {
					It("returns an empty string", func() {
						startCmd, err := subject.StartCommand()
						Expect(err).To(BeNil())
						Expect(startCmd).To(Equal(""))
					})
				})
			})

			Context("The csproj file has an AssemblyName tag", func() {
				BeforeEach(func() {
					Expect(os.MkdirAll(filepath.Join(buildDir, "subdir"), 0755)).To(Succeed())
					csprojContents := `
<Project Sdk="Microsoft.NET.Sdk.Web">
	<PropertyGroup>
		<AssemblyName>f.red.csproj</AssemblyName>
	</PropertyGroup>
</Project>`
					Expect(ioutil.WriteFile(filepath.Join(buildDir, "subdir", "fred.csproj"), []byte(csprojContents), 0644)).To(Succeed())
					Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "dotnet_publish"), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "dotnet_publish", "f.red"), []byte(""), 0755)).To(Succeed())
				})

				It("returns a start command with the AssemblyName instead of filename", func() {
					startCmd, err := subject.StartCommand()
					Expect(err).To(BeNil())
					Expect(startCmd).To(Equal(filepath.Join("${DEPS_DIR}", depsIdx, "dotnet_publish", "f.red")))
				})
			})
		})

		Context("mainPath could not be determined", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, "dotnet_publish"), 0755)).To(Succeed())
			})

			It("returns an empty string", func() {
				startCmd, err := subject.StartCommand()
				Expect(err).To(BeNil())
				Expect(startCmd).To(Equal(""))
			})
		})
	})

	Describe("GetVersionFromDepsJSON", func() {
		Context("when a .deps.json does contain aspnetcore.app", func() {
			BeforeEach(func() {
				createDepsJSON("Microsoft.AspNetCore.App", "2.1.1", false)
			})

			It("Returns the associated version", func() {
				version, err := subject.GetVersionFromDepsJSON("Microsoft.AspNetCore.App")
				Expect(err).To(BeNil())
				Expect(version).To(Equal("2.1.1"))
			})
		})

		Context("when a .deps.json does not contain aspnetcore.app", func() {
			BeforeEach(func() {
				createDepsJSON("Totally.Fake.Library", "2.1.1", false)
			})

			It("returns an error", func() {
				_, err := subject.GetVersionFromDepsJSON("Microsoft.AspNetCore.App")
				Expect(err).Should(MatchError("could not find library Microsoft.AspNetCore.App"))
			})
		})

		Context("when a .deps.json is not present", func() {
			It("returns an error", func() {
				_, err := subject.GetVersionFromDepsJSON("Microsoft.AspNetCore.App")
				Expect(err).Should(MatchError("multiple or no *.deps.json files present"))
			})
		})
	})

	Describe("FindMatchingFrameworkVersion", func() {
		Context("when applyPatches is false", func() {
			applyPatches := false

			Context("and the manifest has the exact version", func() {
				BeforeEach(func() {
					mockManifest.EXPECT().AllDependencyVersions("dotnet-runtime").Return([]string{"4.5.6", "4.5.7"})
				})

				It("Returns the exact version", func() {
					version, err := subject.FindMatchingFrameworkVersion("dotnet-runtime", "4.5.6", &applyPatches)
					Expect(err).To(BeNil())
					Expect(version).To(Equal("4.5.6"))
				})
			})
		})

		Context("when applyPatches is true", func() {
			applyPatches := true

			BeforeEach(func() {
				mockManifest.EXPECT().AllDependencyVersions("dotnet-runtime").Return([]string{"4.5.6", "4.5.7"})
			})

			It("Returns the same major.minor version with the highest available patch", func() {
				version, err := subject.FindMatchingFrameworkVersion("dotnet-runtime", "4.5.6", &applyPatches)
				Expect(err).To(BeNil())
				Expect(version).To(Equal("4.5.7"))
			})
		})
	})

	Describe("ProjectFilePaths", func() {
		BeforeEach(func() {
			for _, name := range []string{
				"first.csproj",
				"other.txt",
				"dir/second.csproj",
				".cloudfoundry/other.csproj",
				"dir/other.txt",
				"a/b/first.vbproj",
				"b/c/first.fsproj",
				"c/d/other.txt",
			} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			}
		})

		It("returns csproj, fsproj and vbproj files (excluding .cloudfoundry)", func() {
			Expect(subject.ProjectFilePaths()).To(ConsistOf([]string{
				filepath.Join(buildDir, "first.csproj"),
				filepath.Join(buildDir, "dir", "second.csproj"),
				filepath.Join(buildDir, "a", "b", "first.vbproj"),
				filepath.Join(buildDir, "b", "c", "first.fsproj"),
			}))
		})
	})

	Describe("IsPublished", func() {
		BeforeEach(func() {
			for _, name := range []string{
				"first.csproj",
				"c/d/other.txt",
			} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			}
		})

		Context("*.runtimeconfig.json exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "fred.runtimeconfig.json"), []byte(""), 0644)).To(Succeed())
			})

			It("returns true", func() {
				Expect(subject.IsPublished()).To(BeTrue())
			})
		})

		Context("*.runtimeconfig.json does NOT exist", func() {
			It("returns false", func() {
				Expect(subject.IsPublished()).To(BeFalse())
			})
		})
	})

	Describe("IsFDD", func() {
		BeforeEach(func() {
			for _, name := range []string{
				"first.csproj",
				"c/d/other.txt",
			} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			}
		})

		Context("*.runtimeconfig.json exists", func() {
			It("returns true when frameworkName is set", func() {
				createRuntimeConfig("anything", "1.1.1")
				Expect(subject.IsFDD()).To(BeTrue())
			})

			It("returns true when frameworkName is not set", func() {
				createRuntimeConfig("", "")
				Expect(subject.IsFDD()).To(BeFalse())
			})
		})

		Context("*.runtimeconfig.json does NOT exist", func() {
			It("returns false", func() {
				Expect(subject.IsFDD()).To(BeFalse())
			})
		})
	})

	Describe("IsSourceBased", func() {
		BeforeEach(func() {
			for _, name := range []string{
				"first.csproj",
				"c/d/other.txt",
			} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			}
		})

		Context("*.runtimeconfig.json exists", func() {
			BeforeEach(func() {
				createRuntimeConfig("", "")
			})

			It("returns false", func() {
				Expect(subject.IsSourceBased()).To(BeFalse())
			})
		})

		Context("*.runtimeconfig.json does NOT exist", func() {
			It("returns true", func() {
				Expect(subject.IsSourceBased()).To(BeTrue())
			})
		})
	})

	Describe("IsFsharp", func() {
		BeforeEach(func() {
			for _, name := range []string{
				"first.csproj",
				"c/d/other.txt",
			} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			}
		})

		Context(".fsproj file exists", func() {
			BeforeEach(func() {
				name := "a/c/something.fsproj"
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			})

			It("returns true", func() {
				Expect(subject.IsFsharp()).To(BeTrue())
			})
		})

		Context(".fsproj file does NOT exist", func() {
			It("returns false", func() {
				Expect(subject.IsFsharp()).To(BeFalse())
			})
		})

		Context(".fsproj file exists inside deps directory (.cloudfoundry)", func() {
			BeforeEach(func() {
				name := ".cloudfoundry/0/a/b/something.fsproj"
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
			})

			It("returns false", func() {
				Expect(subject.IsFsharp()).To(BeFalse())
			})
		})
	})

	Describe("MainPath", func() {
		Context("There is a runtimeconfig file present", func() {
			BeforeEach(func() {

				Expect(ioutil.WriteFile(filepath.Join(buildDir, "fred.runtimeconfig.json"), []byte(""), 0644)).To(Succeed())
			})

			It("returns the runtimeconfig file", func() {
				configFile, err := subject.MainPath()
				Expect(err).To(BeNil())
				Expect(configFile).To(Equal(filepath.Join(buildDir, "fred.runtimeconfig.json")))
			})
		})

		Context("No project path in paths", func() {
			It("returns an empty string", func() {
				path, err := subject.MainPath()
				Expect(err).To(BeNil())
				Expect(path).To(Equal(""))
			})
		})

		Context("Exactly one project path in paths", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(buildDir, "subdir"), 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "subdir", "first.csproj"), []byte(""), 0644)).To(Succeed())
			})
			It("returns that one path", func() {
				path, err := subject.MainPath()
				Expect(err).To(BeNil())
				Expect(path).To(Equal(filepath.Join(buildDir, "subdir", "first.csproj")))
			})
		})

		Context("More than one project path in paths", func() {
			BeforeEach(func() {
				for _, name := range []string{
					"first.csproj",
					"other.txt",
					"dir/second.csproj",
					".cloudfoundry/other.csproj",
					"dir/other.txt",
					"a/b/first.vbproj",
					"b/c/first.fsproj",
					"c/d/other.txt",
				} {
					Expect(os.MkdirAll(filepath.Dir(filepath.Join(buildDir, name)), 0755)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(buildDir, name), []byte(""), 0644)).To(Succeed())
				}
			})

			Context("There is a .deployment file present", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(buildDir, ".deployment"), []byte("[config]\nproject = ./a/b/first.vbproj"), 0644)).To(Succeed())
				})
				It("returns the path specified in the .deployment file.", func() {
					path, err := subject.MainPath()
					Expect(err).To(BeNil())
					Expect(path).To(Equal(filepath.Join(buildDir, "a", "b", "first.vbproj")))
				})
			})

			Context("There is NOT a .deployment file present", func() {

				It("Returns an error", func() {
					_, err := subject.MainPath()
					Expect(err).ToNot(BeNil())
				})
			})
		})
	})

	Describe("FDDInstallFrameworks", func() {
		Context("when the app specifies Microsoft.NETCore.App in .runtimeconfig.json", func() {
			BeforeEach(func() {
				createRuntimeConfig("Microsoft.NETCore.App", "7.8.9")
			})

			Context("when it does not find Microsoft.AspNetCore.App in deps.json", func() {
				BeforeEach(func() {
					createDepsJSON("", "", true)
				})

				It("installs the dotnet-runtime", func() {
					mockInstaller.
						EXPECT().
						InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "7.8.9"}, depsPath)

					Expect(subject.FDDInstallFrameworks()).To(Succeed())
				})
			})

			Context("when it finds the Microsoft.AspNetCore.App in deps.json", func() {
				BeforeEach(func() {
					createDepsJSON("Microsoft.AspNetCore.App", "2.3.4", true)
				})

				It("installs the dotnet-runtime and dotnet-aspnetcore", func() {
					mockInstaller.
						EXPECT().
						InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "7.8.9"}, depsPath)
					mockInstaller.
						EXPECT().
						InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "2.3.4"}, depsPath)

					Expect(subject.FDDInstallFrameworks()).To(Succeed())
				})
			})

			Context("when the version of Microsoft.AspNetCore.App found in deps.json is less than 2.1.0", func() {
				BeforeEach(func() {
					createDepsJSON("Microsoft.AspNetCore.App", "2.0.0", true)
				})

				It("installs dotnet-runtime and does not install dotnet-aspnetcore", func() {
					mockInstaller.
						EXPECT().
						InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "7.8.9"}, depsPath).
						Times(1)
					mockInstaller.
						EXPECT().
						InstallDependency(gomock.Any(), gomock.Any()).
						Times(0)

					Expect(subject.FDDInstallFrameworks()).To(Succeed())
				})
			})
		})

		Context("when the app specifies Microsoft.AspNetCore.App in .runtimeconfig.json", func() {
			BeforeEach(func() {
				createRuntimeConfig("Microsoft.AspNetCore.App", "6.7.8")
			})

			It("installs the dotnet-aspnetcore from the app's runtimeconfig.json and dotnet-runtime from Microsoft.AspNetCore.App.runtimeconfig.json", func() {
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "6.7.8"}, depsPath).
					Do(func(d libbuildpack.Dependency, s string) {
						installRuntimeConfig("Microsoft.AspNetCore.App", "6.7.8", "1.2.3")
					})
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "1.2.3"}, depsPath)

				Expect(subject.FDDInstallFrameworks()).To(Succeed())
			})
		})
		Context("when the app specifies Microsoft.AspNetCore.All in .runtimeconfig.json", func() {
			BeforeEach(func() {
				createRuntimeConfig("Microsoft.AspNetCore.All", "6.7.8")
			})

			It("installs the dotnet-aspnetcore from the app's runtimeconfig.json and dotnet-runtime from Microsoft.AspNetCore.All.runtimeconfig.json", func() {
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "6.7.8"}, depsPath).
					Do(func(d libbuildpack.Dependency, s string) {
						installRuntimeConfig("Microsoft.AspNetCore.All", "6.7.8", "1.2.3")
					})
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "1.2.3"}, depsPath)

				Expect(subject.FDDInstallFrameworks()).To(Succeed())
			})
		})

		Context("when Microsoft.AspNetCore.App is in .runtimeconfig.json", func() {
			BeforeEach(func() {
				createRuntimeConfig("Microsoft.AspNetCore.App", "2.0.0")
			})

			It("fails when the dotnet-aspnetcore version is less than 2.1.0", func() {
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "2.0.0"}, gomock.Any()).
					Return(errors.New("could not find dotnet-aspnetcore version 2.0.0"))
				Expect(subject.FDDInstallFrameworks()).NotTo(Succeed())
			})
		})
	})

	Describe("SourceInstallDotnetRuntime", func() {
		Context("when the runtime version is only specified under <TargetFramework> in the csproj", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"),
					[]byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">
	<PropertyGroup>
		<TargetFramework>netcoreapp6.7</TargetFramework>
	</PropertyGroup>
</Project>`), 0644)).To(Succeed())
			})

			It("installs the latest runtime for that minor", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-runtime").Return([]string{"4.5.6", "6.7.8", "6.7.9", "6.8.9"})
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "6.7.9"}, depsPath)

				Expect(subject.SourceInstallDotnetRuntime()).To(Succeed())
			})
		})

		Context("when the exact version is specified under RuntimeFrameworkVersion in the csproj", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"),
					[]byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">
	<PropertyGroup>
		<TargetFramework>netcoreapp6.6</TargetFramework>
		<RuntimeFrameworkVersion>6.7.8</RuntimeFrameworkVersion>
	</PropertyGroup>
</Project>`), 0644)).To(Succeed())
			})

			It("installs the runtime", func() {
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "6.7.8"}, depsPath)

				Expect(subject.SourceInstallDotnetRuntime()).To(Succeed())
			})
		})

		Context("when a floating version is specified under RuntimeFrameworkVersion in the csproj", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"),
					[]byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">
	<PropertyGroup>
		<TargetFramework>netcoreapp6.6</TargetFramework>
		<RuntimeFrameworkVersion>6.7.*</RuntimeFrameworkVersion>
	</PropertyGroup>
</Project>`), 0644)).To(Succeed())
			})

			It("installs the runtime floating on patch", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-runtime").Return([]string{"4.5.6", "6.7.8", "6.7.9", "6.8.9"})
				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: "6.7.9"}, depsPath)

				Expect(subject.SourceInstallDotnetRuntime()).To(Succeed())
			})
		})
	})

	Describe("SourceInstallDotnetAspNetCore", func() {
		Context("when the Microsoft.AspNetCore.App version is discovered via .csproj", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"),
					[]byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">
  <ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.App" Version="1.2.3"/>
  </ItemGroup>
</Project>`), 0666)).To(Succeed())
			})

			It("installs the dotnet-aspnetcore specified by PackageReference floating on patch", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-aspnetcore").Return([]string{"1.2.3", "1.2.4", "6.7.9", "6.8.9"}).
					Times(2)

				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "1.2.4"}, gomock.Any()).
					Times(1)

				Expect(subject.SourceInstallDotnetAspNetCore()).To(Succeed())
			})
		})

		Context("when the dotnet-aspnetcore version is not found in the csproj", func() {
			BeforeEach(func() {
				ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"), []byte("<valid/>"), 0644)
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, ".nuget", "packages", "microsoft.aspnetcore.app", "4.5.6"), 0755)).To(Succeed())
			})

			It("installs the dotnet-aspnetcore found in the nuget cache", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-aspnetcore").Return([]string{"4.5.6"})

				mockInstaller.
					EXPECT().
					InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: "4.5.6"}, depsPath).
					Times(1)

				Expect(subject.SourceInstallDotnetAspNetCore()).To(Succeed())
			})
		})

		Context("when the dotnet-aspnetcore version found in the nuget cache is less than 2.1", func() {
			BeforeEach(func() {
				ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"), []byte("<valid/>"), 0644)
				Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx, ".nuget", "packages", "microsoft.aspnetcore.app", "4.5.6"), 0755)).To(Succeed())
			})

			It("does not install the dotnet-aspnetcore", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-aspnetcore").Return([]string{"2.0.0"})

				mockInstaller.
					EXPECT().
					InstallDependency(gomock.Any(), gomock.Any()).
					Times(0)

				Expect(subject.SourceInstallDotnetAspNetCore()).To(Succeed())
			})
		})

		Context("when the dotnet-aspnetcore version is less than 2.1", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "foo.csproj"),
					[]byte(`
<Project Sdk="Microsoft.NET.Sdk.Web">
  <ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.App" Version="2.0.0"/>
  </ItemGroup>
</Project>`), 0666)).To(Succeed())
			})

			It("it will not install old 2.0 version of dotnet-aspnetcore", func() {
				mockManifest.
					EXPECT().
					AllDependencyVersions("dotnet-aspnetcore").
					Times(0)

				mockInstaller.
					EXPECT().
					InstallDependency(gomock.Any(), gomock.Any()).
					Times(0)

				Expect(subject.SourceInstallDotnetAspNetCore()).To(Succeed())
			})
		})
	})
})
