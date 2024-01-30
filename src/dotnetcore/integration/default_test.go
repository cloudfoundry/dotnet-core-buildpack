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

		app                   *cutlass.App
		latest6RuntimeVersion string
		latest6ASPNetVersion  string
		latest6SDKVersion     string
		latest7RuntimeVersion string
		latest8RuntimeVersion string
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "simple"))

		bpDir, err := cutlass.FindRoot()
		Expect(err).NotTo(HaveOccurred())

		latest6RuntimeVersion = GetLatestDepVersion(t, "dotnet-runtime", "6.0.x", bpDir)

		latest6ASPNetVersion = GetLatestDepVersion(t, "dotnet-aspnetcore", "6.0.x", bpDir)

		latest6SDKVersion = GetLatestDepVersion(t, "dotnet-sdk", "6.0.x", bpDir)

		latest7RuntimeVersion = GetLatestDepVersion(t, "dotnet-runtime", "7.0.x", bpDir)

		latest8RuntimeVersion = GetLatestDepVersion(t, "dotnet-runtime", "8.0.x", bpDir)
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("deploying a source-based app", func() {
		it("builds and runs the app and accepts SIGTERM and exits gracefully", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest6SDKVersion)))
			Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest6RuntimeVersion)))
			Expect(app.GetBody("/")).To(ContainSubstring("Welcome to .NET 6"))

			Expect(app.Stop()).To(Succeed())
			Eventually(func() string { return app.Stdout.String() }, 30*time.Second, 1*time.Second).Should(ContainSubstring("Application is shutting down..."))
		})

		context("with dotnet sdk 6 in global json", func() {
			context("when the sdk exists", func() {
				it.Before(func() {
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "simple_global_json_6"), "global.json", "sdk_version", latest6SDKVersion)
				})

				it("displays a simple text homepage", func() {
					PushAppAndConfirm(t, app)
					Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest6SDKVersion)))
					Expect(app.GetBody("/")).To(ContainSubstring("Welcome to .NET 6"))
				})
			})
			context("when the sdk is missing", func() {
				var (
					constructedVersion string
					baseFeatureLine    int
					proceed            bool
				)

				it.Before(func() {
					version, err := semver.NewVersion(latest6SDKVersion)
					Expect(err).ToNot(HaveOccurred())

					if version.Patch()%100 != 0 {
						proceed = true
					}

					baseFeatureLine = int((version.Patch() / 100) * 100)
					constructedVersion = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), baseFeatureLine)
					app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "simple_global_json_6"), "global.json", "sdk_version", constructedVersion)
				})

				it("logs a warning about using source_apps SDK", func() {
					PushAppAndConfirm(t, app)
					if proceed {
						Expect(app.Stdout.String()).To(ContainSubstring(fmt.Sprintf("SDK %s in global.json is not available", constructedVersion)))
						Expect(app.Stdout.String()).To(ContainSubstring("falling back to latest version in version line"))
					}
					Expect(app.GetBody("/")).To(ContainSubstring("Welcome to .NET 6"))
				})
			})
		})

		context("with buildpack.yml and global.json files", func() {
			it.Before(func() {
				app = ReplaceFileTemplate(t, filepath.Join(settings.FixturesPath, "source_apps", "multi_version_sources"), "global.json", "sdk_version", latest6SDKVersion)
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

		context("when an app has a Microsoft.AspNetCore.App", func() {

			context("with version 6", func() {
				it.Before(func() {
					app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "aspnet_package_reference_6"))
					app.Disk = "2G"
				})

				it("publishes and runs, installing the correct runtime and aspnetcore version with a warning", func() {
					PushAppAndConfirm(t, app)
					Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest6ASPNetVersion)))
					Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest6RuntimeVersion)))
					Eventually(app.Stdout.String()).Should(ContainSubstring("A PackageReference to Microsoft.AspNetCore.App is not necessary when targeting .NET Core 3.0 or higher."))
					Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
				})
			})
		})

		context("when the app has Microsoft.AspNetCore.All", func() {
			context("with version 6", func() {
				it.Before(func() {
					app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "source_6.0"))
					app.Disk = "1G"
				})

				it("publishes and runs, installing the a roll forward runtime and aspnetcore versions", func() {
					PushAppAndConfirm(t, app)
					Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest6RuntimeVersion)))
					Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest6ASPNetVersion)))
					Expect(app.GetBody("/")).To(ContainSubstring("building Web apps with ASP.NET Core"))
				})
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
				Expect(app.GetBody("/")).To(ContainSubstring("Welcome to .NET 6"))
			})
		})

		context("with .NET Core 7", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "source-app-7"))
				app.Disk = "2G"
				app.Memory = "1G"
			})

			it("builds and runs successfully", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest7RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("<title>source_app_7</title>"))
			})
		})

		context("with .NET Core 8", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "source_8"))
				app.Disk = "2G"
				app.Memory = "1G"
			})

			it("builds and runs successfully", func() {
				PushAppAndConfirm(t, app)
				Eventually(app.Stdout.String()).Should(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
				Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
			})
		})

		context("with use_legacy_openssl specified in buildpack.yml", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "simple_legacy_openssl"))
			})

			it("activates openssl legacy provider and builds/runs successfully", func() {
				Expect(app.Push()).To(Succeed())
				Expect(app.Stdout.String()).To(ContainSubstring("Loading legacy SSL provider"))
				Eventually(app.Stdout.String()).Should(ContainSubstring("name: OpenSSL Legacy Provider"))
			})
		})
	})

	context("deploying a framework-dependent app", func() {
		context("with libgdiplus", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "util", "libgdiplus", "bin", "Release", "net6.0", "ubuntu.18.04-x64", "publish"))
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
				Expect(app.GetBody("/")).To(ContainSubstring("Welcome to .NET 6"))
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
