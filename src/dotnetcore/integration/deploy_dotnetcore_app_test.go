package integration_test

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	var (
		latest20RuntimeVersion                           string
		latest21RuntimeVersion, previous21RuntimeVersion string
		latest21ASPNetVersion, previous21ASPNetVersion   string
		latest20SDKVersion                               string
		latest21SDKVersion, previous21SDKVersion         string
	)

	BeforeEach(func() {
		latest20RuntimeVersion = GetLatestDepVersion("dotnet-runtime", "2.0.x", bpDir)

		latest21RuntimeVersion = GetLatestDepVersion("dotnet-runtime", "2.1.x", bpDir)
		previous21RuntimeVersion = GetLatestDepVersion("dotnet-runtime", fmt.Sprintf("<%s", latest21RuntimeVersion), bpDir)

		latest21ASPNetVersion = GetLatestDepVersion("dotnet-aspnetcore", "2.1.x", bpDir)
		previous21ASPNetVersion = GetLatestDepVersion("dotnet-aspnetcore", fmt.Sprintf("<%s", latest21ASPNetVersion), bpDir)

		latest20SDKVersion = GetLatestDepVersion("dotnet-sdk", "2.0.x", bpDir)

		latest21SDKVersion = GetLatestDepVersion("dotnet-sdk", "2.1.x", bpDir)
		previous21SDKVersion = GetLatestDepVersion("dotnet-sdk", fmt.Sprintf("<%s", latest21SDKVersion), bpDir)
	})

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("deploying simple web app with dotnet-runtime 1.0", func() {
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs2")
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "dotnet1.0"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
		})
	})

	Context("deploying simple web app with dotnet-runtime 2.0", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "dotnet2"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})

		//AssertUsesProxyDuringStagingIfPresent("dotnet2")
	})

	Context("deploying simple web app with dotnet 2.0 using dotnet 2.0 sdk", func() {
		BeforeEach(func() {
			app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "dotnet2_with_global_json"), "global.json", "sdk_version", latest20SDKVersion)
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest20SDKVersion)))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})
	})

	Context("deploying with a buildpack.yml and global.json files", func() {
		Context("when SDK versions don't match", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "with_buildpack_yml"), "global.json", "sdk_version", latest20SDKVersion)
			})

			It("installs the specific version from buildpack.yml instead of global.json", func() {
				app = ReplaceFileTemplate(app.Path, "buildpack.yml", "sdk_version", previous21SDKVersion)
				app.Push()

				Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", previous21SDKVersion)))
			})

			It("installs the floated version from buildpack.yml instead of global.json", func() {
				app = ReplaceFileTemplate(app.Path, "buildpack.yml", "sdk_version", "2.1.x")
				app.Push()

				Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest21SDKVersion)))
			})
		})

		Context("when SDK version from buildpack.yml is not available", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "with_buildpack_yml"), "buildpack.yml", "sdk_version", "2.0.0-preview7")
			})

			It("fails due to missing SDK", func() {
				Expect(app.Push()).ToNot(Succeed())

				Eventually(app.Stdout.String).Should(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
				Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"))
			})
		})
	})

	Context("deploying an mvc app with node prerendering", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "asp_prerender_node"))
			app.Disk = "2G"
			app.Memory = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("1 + 2 = 3"))
		})
	})

	Context("deploying simple web app with missing sdk", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "missing_sdk"))
		})

		It("Logs a warning about using default SDK", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("SDK 2.0.0-preview-007 in global.json is not available"))
			Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
		})
	})

	Context("deploying an msbuild app with RuntimeIdentfier", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained_msbuild"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(MatchRegexp("Removing dotnet-sdk"))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})

	Context("ASP.Netcore App 2.1 source based app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "aspnetcore21_source"))

			app.Disk = "2G"
			app.Memory = "2G"
		})

		It("publishes and runs, using the correct runtime and aspnetcore versions", func() {
			PushAppAndConfirm(app)
			Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
			Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))

			By("accepts SIGTERM and exits gracefully")
			Expect(app.Stop()).ToNot(HaveOccurred())
			time.Sleep(1 * time.Second) // Wait here to flush the log process buffer ¯\_(ツ)_/¯
			Eventually(app.Stdout.String(), 30*time.Second, time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
		})
	})

	Context("simple source-based netcoreapp2", func() {
		Context("runtime version explicitly defined in csproj", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "netcoreapp2_explicit_runtime_csproj"), "netcoreapp2.csproj", "runtime_version", previous21RuntimeVersion)

				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, using exact runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", previous21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("runtime version floated in csproj", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "netcoreapp2_float_runtime_csproj"))
				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, using latest patch runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("runtime version not defined in csproj", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "netcoreapp2_no_runtime_csproj"))
				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, using latest patch runtimes in the nuget cache", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest20RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("dotnet-runtime %s is already installed", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})
	})

	Context("with runtimeconfig.json", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "float_runtimeconfig"))
		})

		It("installs the latest patch of dotnet runtime from the runtimeconfig.json", func() {
			PushAppAndConfirm(app)
			Eventually(app.Stdout.String()).Should(MatchRegexp(fmt.Sprintf("Installing dotnet-runtime %[1]s|dotnet-runtime %[1]s is already installed", latest21RuntimeVersion)))
		})
	})

	Context("with runtimeconfig.json and applyPatches false", func() {
		BeforeEach(func() {
			app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "apply_patches_false"), "dotnet.runtimeconfig.json", "framework_version", previous21ASPNetVersion)
		})

		It("installs the exact version of dotnet aspnetcore from the runtimeconfig.json", func() {
			PushAppAndConfirm(app)
			Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", previous21ASPNetVersion)))
		})
	})

	Context("for a non-published app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_dot_in_name"))
			app.Memory = "1G"
			app.Disk = "2G"
		})

		It("successfully pushes an app with an AssemblyName", func() {
			PushAppAndConfirm(app)
		})
	})
})
