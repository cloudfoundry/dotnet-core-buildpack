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
		latest21RuntimeVersion, previous21RuntimeVersion string
		latest21ASPNetVersion, previous21ASPNetVersion   string
		latest20SDKVersion                               string
		latest21SDKVersion, previous21SDKVersion         string
	)

	BeforeEach(func() {
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

	Context("deploying a source-based app", func() {
		Context("with dotnet-runtime 1.0", func() {
			BeforeEach(func() {
				SkipUnlessStack("cflinuxfs2")
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "source_web_1.0"))
			})

			It("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
			})
		})

		Context("with dotnet-runtime 2.0", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_source_web_2.0"))
			})

			It("displays a simple text homepage", func() {
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
			})
		})

		Context("with dotnet sdk 2.0 in global json", func() {
			Context("when the sdk exists", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "source_2.0_global_json_templated"), "global.json", "sdk_version", latest20SDKVersion)
				})

				It("displays a simple text homepage", func() {
					PushAppAndConfirm(app)

					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest20SDKVersion)))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
				})

			})

			Context("when the sdk is missing", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "source_2.0_global_json_templated"), "global.json", "sdk_version", "2.0.0-preview-007")
				})

				It("Logs a warning about using default SDK", func() {
					PushAppAndConfirm(app)
					Expect(app.Stdout.String()).To(ContainSubstring("SDK 2.0.0-preview-007 in global.json is not available"))
					Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 2.0"))
				})
			})
		})

		Context("with buildpack.yml and global.json files", func() {
			Context("when SDK versions don't match", func() {
				BeforeEach(func() {
					app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "with_buildpack_yml_templated"), "global.json", "sdk_version", latest20SDKVersion)
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
					app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "with_buildpack_yml_templated"), "buildpack.yml", "sdk_version", "2.0.0-preview7")
				})

				It("fails due to missing SDK", func() {
					Expect(app.Push()).ToNot(Succeed())

					Eventually(app.Stdout.String).Should(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
					Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"))
				})
			})
		})

		Context("with node prerendering", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "source_prerender_node"))
				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("displays a simple text homepage", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("1 + 2 = 3"))
			})
		})

		Context("when RuntimeFrameworkVersion is explicitly defined in csproj", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "source_2.1_explicit_runtime_templated"), "netcoreapp2.csproj", "runtime_version", previous21RuntimeVersion)
				// app = ReplaceFileTemplate(app.Path, "buildpack.yml", "sdk_version", previous21SDKVersion)

				app.Disk = "2G"
				app.Memory = "2G"
				fmt.Printf("previous21runtiem: %s", previous21RuntimeVersion)
				// fmt.Printf("previous21sdk: %s", previous21SDKVersion)
			})

			It("publishes and runs, using exact runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", previous21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when RuntimeFrameworkVersion is floated in csproj", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "source_2.1_float_runtime"))
				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, using latest patch runtime", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when the app has Microsoft.AspNetCore.All version 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "source_aspnetcore_all_2.1"))
				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, using the TargetFramework for the runtime version and the latest 2.1 patch of dotnet-aspnetcore", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Sample pages using ASP.NET Core MVC"))
			})
		})

		Context("when the app has Microsoft.AspNetCore.App version 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "source_aspnetcore_app_2.1"))

				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, installing the correct runtime and aspnetcore versions", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))

				By("accepts SIGTERM and exits gracefully")
				Expect(app.Stop()).To(Succeed())
				Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
			})
		})

		Context("with AssemblyName specified", func() {
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

	Context("deploying an FDD app", func() {
		Context("with Microsoft.AspNetCore.App 2.1", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "fdd_aspnetcore_2.1"))

				app.Disk = "2G"
				app.Memory = "2G"
			})

			It("publishes and runs, and floats the runtime and aspnetcore versions by default", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest21ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest21RuntimeVersion)))

				By("accepts SIGTERM and exits gracefully")
				Expect(app.Stop()).ToNot(HaveOccurred())
				Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
			})
		})

		Context("with Microsoft.AspNetCore.App 2.1 and applyPatches false", func() {
			BeforeEach(func() {
				app = ReplaceFileTemplate(filepath.Join(bpDir, "fixtures", "fdd_apply_patches_false_2.1_templated"), "dotnet.runtimeconfig.json", "framework_version", previous21ASPNetVersion)
			})

			It("installs the exact version of dotnet-aspnetcore from the runtimeconfig.json", func() {
				PushAppAndConfirm(app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", previous21ASPNetVersion)))
			})
		})
	})

	Context("deploying a self contained msbuild app with RuntimeIdentfier", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained_msbuild"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(MatchRegexp("Removing dotnet-sdk"))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})

})
