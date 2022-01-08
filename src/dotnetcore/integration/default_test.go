package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		app *cutlass.App

		latest31RuntimeVersion, previous31RuntimeVersion string
		latest31ASPNetVersion, previous31ASPNetVersion   string
		latest31SDKVersion, previous31SDKVersion         string
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "simple"))

		bpDir, err := cutlass.FindRoot()
		Expect(err).NotTo(HaveOccurred())

		latest31RuntimeVersion = GetLatestDepVersion(t, "dotnet-runtime", "3.1.x", bpDir)
		previous31RuntimeVersion = GetLatestDepVersion(t, "dotnet-runtime", fmt.Sprintf("<%s", latest31RuntimeVersion), bpDir)

		latest31ASPNetVersion = GetLatestDepVersion(t, "dotnet-aspnetcore", "3.1.x", bpDir)
		previous31ASPNetVersion = GetLatestDepVersion(t, "dotnet-aspnetcore", fmt.Sprintf("<%s", latest31ASPNetVersion), bpDir)

		latest31SDKVersion = GetLatestDepVersion(t, "dotnet-sdk", "3.1.x", bpDir)
		previous31SDKVersion = GetLatestDepVersion(t, "dotnet-sdk", fmt.Sprintf("<%s", latest31SDKVersion), bpDir)
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("deploying a source-based app", func() {
		it("builds and runs the app and accepts SIGTERM and exits gracefully", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest31SDKVersion)))
			Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest31RuntimeVersion)))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 3.1"))

			Expect(app.Stop()).To(Succeed())
			Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
		})

		context("with dotnet sdk 3.1 in global json", func() {
			context("when the sdk exists", func() {
				it.Before(func() {
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "simple_global_json"), "global.json", "sdk_version", latest31SDKVersion)
				})

				it("displays a simple text homepage", func() {
					PushAppAndConfirm(t, app)
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest31SDKVersion)))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 3.1"))
				})
			})
			context("when the sdk is missing", func() {
				var constructedVersion string

				it.Before(func() {
					version, err := semver.NewVersion(latest31SDKVersion)
					Expect(err).ToNot(HaveOccurred())
					baseFeatureLine := (version.Patch() / 100) * 100
					constructedVersion = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), baseFeatureLine)
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "simple_global_json"), "global.json", "sdk_version", constructedVersion)
				})

				it("logs a warning about using source_apps SDK", func() {
					PushAppAndConfirm(t, app)
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("SDK %s in global.json is not available", constructedVersion)))
					Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello From Dotnet 3.1"))
				})
			})
		})

		context("with buildpack.yml and global.json files", func() {
			it.Before(func() {
				app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "multi_version_sources"), "global.json", "sdk_version", latest31SDKVersion)
			})

			context("when SDK versions don't match", func() {
				it("installs the specific version from buildpack.yml instead of global.json", func() {
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "multi_version_sources"), "buildpack.yml", "sdk_version", previous31SDKVersion)
					app.Push()
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", previous31SDKVersion)))
				})

				it("installs the floated version from buildpack.yml instead of global.json", func() {
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "multi_version_sources"), "buildpack.yml", "sdk_version", "3.1.x")
					app.Push()
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest31SDKVersion)))
				})
			})

			context("when SDK version from buildpack.yml is not available", func() {
				it("fails due to missing SDK", func() {
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "multi_version_sources"), "buildpack.yml", "sdk_version", "2.0.0-preview7")
					Expect(app.Push()).ToNot(Succeed())

					Eventually(app.Stdout.String).Should(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
					Eventually(app.Stdout.String).Should(ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"))
				})
			})
		})

		context("when RuntimeFrameworkVersion is explicitly defined in csproj", func() {
			it.Before(func() {
				app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "templated_runtime"), "templated_runtime.csproj", "runtime_version", previous31RuntimeVersion)
				app.Disk = "2G"
			})

			it("publishes and runs, using exact runtime", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", previous31RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
			})
		})

		context("when a 3.1 app has a Microsoft.AspNetCore.App version 3.1", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "aspnet_package_reference"))
				app.Disk = "2G"
			})

			it("publishes and runs, installing the correct runtime and aspnetcore version with a warning", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest31ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest31RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring("A PackageReference to Microsoft.AspNetCore.App is not necessary when targeting .NET Core 3.0 or higher."))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
			})
		})

		context("when the app has Microsoft.AspNetCore.All version 3.0", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "source_3.0"))
				app.Disk = "1G"
			})

			it("publishes and runs, installing the a roll forward runtime and aspnetcore versions", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest31RuntimeVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest31ASPNetVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("building Web apps with ASP.NET Core"))
			})
		})

		context("with AssemblyName specified", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "with_dot_in_name"))
				app.Disk = "2G"
			})

			it("successfully pushes an app with an AssemblyName", func() {
				PushAppAndConfirm(t, app)
			})
		})

		context("with libgdiplus", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "util", "libgdiplus"))
			})

			it("displays a simple text homepage", func() {
				PushAppAndConfirm(t, app)
				Expect(app.Stdout.String()).To(ContainSubstring("Installing libgdiplus"))
			})
		})

		context("with .NET Core 6", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "source_6.0"))
			})

			it("builds and runs successfully", func() {
				PushAppAndConfirm(t, app)
				Expect(app.GetBody("/")).To(ContainSubstring("Welcome"))
			})
		})
	})

	context("deploying a framework-dependent app", func() {
		context("with Microsoft.AspNetCore.App 3.1", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "fdd_apps", "simple"))
				app.Disk = "2G"
			})

			it("publishes and runs, and floats the runtime and aspnetcore versions by default", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest31SDKVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest31ASPNetVersion)))
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest31RuntimeVersion)))
				Expect(app.Stop()).ToNot(HaveOccurred())
				Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Goodbye, cruel world!"))
			})
		})

		context("with Microsoft.AspNetCore.App 3.1 and applyPatches false", func() {
			it.Before(func() {
				app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "fdd_apps", "templated_framework"), "templated_framework.runtimeconfig.json", "framework_version", previous31ASPNetVersion)
			})

			it("installs the exact version of dotnet-aspnetcore from the runtimeconfig.json", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", previous31ASPNetVersion)))
			})
		})

		context("with libgdiplus", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "util", "libgdiplus", "bin", "Release", "netcoreapp3.1", "linux-x64", "publish"))
			})

			it("displays a simple text homepage", func() {
				PushAppAndConfirm(t, app)
				Expect(app.Stdout.String()).To(ContainSubstring("Installing libgdiplus"))
			})
		})

		context("with .NET Core 6", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "fde_apps", "fde_6.0"))
			})

			it("builds and runs successfully", func() {
				PushAppAndConfirm(t, app)
				Expect(app.GetBody("/")).To(ContainSubstring("Welcome"))
			})
		})
	})

	context("deploying a self contained msbuild app with RuntimeIdentfier", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "self_contained_apps", "msbuild"))
		})

		it("displays a simple text homepage", func() {
			PushAppAndConfirm(t, app)
			Expect(app.Stdout.String()).To(MatchRegexp("Removing dotnet-sdk"))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})

	context("deploying .NET Core 6 self-contained app", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "self_contained_apps", "self_contained_executable_6.0"))
		})

		it("builds and runs successfully", func() {
			PushAppAndConfirm(t, app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello, world!"))
		})
	})
}
