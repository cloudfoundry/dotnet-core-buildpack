package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testDefault(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			fixture               string
			name                  string
			latest9RuntimeVersion string
			latest8RuntimeVersion string
			latest8SDKVersion     string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())

			latest9RuntimeVersion, err = GetLatestDepVersion(t, "dotnet-runtime", "9.0.x")
			Expect(err).NotTo(HaveOccurred())

			latest8RuntimeVersion, err = GetLatestDepVersion(t, "dotnet-runtime", "8.0.x")
			Expect(err).NotTo(HaveOccurred())

			latest8SDKVersion, err = GetLatestDepVersion(t, "dotnet-sdk", "8.0.x")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("deploying a source-based app", func() {
			it.Before(func() {
				var err error
				fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.Execute(name, fixture)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest8SDKVersion)))
				Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
				Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 8")))
			})

			context("with dotnet sdk 8 in global json", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_global_json_8"))
					Expect(err).NotTo(HaveOccurred())
				})

				context("when the sdk exists", func() {
					it.Before(func() {
						Expect(ReplaceFileTemplate(t, fixture, "global.json", "sdk_version", latest8SDKVersion)).To(Succeed())
					})

					it("displays a simple text homepage", func() {
						deployment, logs, err := platform.Deploy.Execute(name, fixture)
						Expect(err).NotTo(HaveOccurred())

						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest8SDKVersion)))
						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
						Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 8")))
					})
				})

				context("when the sdk is missing", func() {
					var (
						constructedVersion string
						baseFeatureLine    int
						proceed            bool
					)

					it.Before(func() {
						version, err := semver.NewVersion(latest8SDKVersion)
						Expect(err).ToNot(HaveOccurred())

						if version.Patch()%100 != 0 {
							proceed = true
						}

						baseFeatureLine = int((version.Patch() / 100) * 100)
						constructedVersion = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), baseFeatureLine)

						Expect(ReplaceFileTemplate(t, fixture, "global.json", "sdk_version", constructedVersion)).To(Succeed())
					})

					it("logs a warning about using source_apps SDK", func() {
						deployment, logs, err := platform.Deploy.Execute(name, fixture)
						Expect(err).NotTo(HaveOccurred())

						if proceed {
							Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest8SDKVersion)))
							Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
						}
						Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 8")))
					})
				})
			})

			context("with buildpack.yml and global.json files", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "multi_version_sources"))
					Expect(err).NotTo(HaveOccurred())

					Expect(ReplaceFileTemplate(t, fixture, "global.json", "sdk_version", latest8SDKVersion)).To(Succeed())
					Expect(ReplaceFileTemplate(t, fixture, "buildpack.yml", "sdk_version", "2.0.0-preview7")).To(Succeed())
				})

				context("when SDK version from buildpack.yml is not available", func() {
					it("fails due to missing SDK", func() {
						_, logs, err := platform.Deploy.Execute(name, fixture)
						Expect(err).To(HaveOccurred())

						Expect(logs).To(ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"))
						Expect(logs).To(ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"))
					})
				})
			})

			context("when an app has a Microsoft.AspNetCore.App", func() {
				context("with version 8", func() {
					it.Before(func() {
						var err error
						fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "aspnet_package_reference_8"))
						Expect(err).NotTo(HaveOccurred())
					})

					it("publishes and runs, installing the correct runtime and aspnetcore version with a warning", func() {
						deployment, logs, err := platform.Deploy.Execute(name, fixture)
						Expect(err).NotTo(HaveOccurred())

						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-aspnetcore %s", latest8RuntimeVersion)))
						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
						Expect(logs).To(ContainSubstring("A PackageReference to Microsoft.AspNetCore.App is not necessary when targeting .NET Core 3.0 or higher."))
						Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
					})
				})
			})

			context("when the app has Microsoft.AspNetCore.All", func() {
				context("with version 8", func() {
					it.Before(func() {
						var err error
						fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "source_8.0"))
						Expect(err).NotTo(HaveOccurred())
					})

					it("publishes and runs, installing the a roll forward runtime and aspnetcore versions", func() {
						deployment, logs, err := platform.Deploy.Execute(name, fixture)
						Expect(err).NotTo(HaveOccurred())

						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latest8SDKVersion)))
						Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest8RuntimeVersion)))
						Eventually(deployment).Should(Serve(ContainSubstring("building Web apps with ASP.NET Core")))
					})
				})
			})

			context("with AssemblyName specified", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "with_dot_in_name"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("successfully pushes an app with an AssemblyName", func() {
					deployment, _, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
				})
			})

			context("with libgdiplus", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "util", "libgdiplus"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("displays a simple text homepage", func() {
					deployment, logs, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainSubstring("Installing libgdiplus"))
					Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
				})
			})

			context("with .NET Core 8", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "source_8.0"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("builds and runs successfully", func() {
					deployment, _, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 8")))
				})
			})

			context("with .NET Core 9", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "source_9.0"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("builds and runs successfully", func() {
					deployment, logs, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latest9RuntimeVersion)))
					Eventually(deployment).Should(Serve(ContainSubstring("building Web apps with ASP.NET Core")))
				})
			})

			context("with BP_USE_LEGACY_OPENSSL set to `true`", func() {
				it.Before(func() {
					// this feature is not available on cflinuxfs3, because the stack already supports the legacy ssl provider
					SkipOnCflinuxfs3(t)

					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_legacy_openssl"))
					Expect(err).NotTo(HaveOccurred())
				})

		it("activates openssl legacy provider and builds/runs successfully", func() {
			deployment, logs, err := platform.Deploy.
				WithEnv(map[string]string{
					"BP_OPENSSL_ACTIVATE_LEGACY_PROVIDER": "true",
				}).
				Execute(name, fixture)
			Expect(err).NotTo(HaveOccurred())

			// Check that the legacy SSL provider was loaded during build
			Expect(logs).To(ContainSubstring("Loading legacy SSL provider"))
			Eventually(func() string {
					logs, _ := deployment.RuntimeLogs()
					return logs 
					}, "10s", "1s").Should(Or(ContainSubstring("name: OpenSSL Legacy Provider"),
				))
			})
		})
		})

		context("deploying a framework-dependent app", func() {
			context("with libgdiplus", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "util", "libgdiplus", "bin", "Release", "net8.0", "linux-x64", "publish"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("displays a simple text homepage", func() {
					deployment, logs, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainSubstring("Installing libgdiplus"))
					Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
				})
			})

			context("with .NET Core 8", func() {
				it.Before(func() {
					var err error
					fixture, err = switchblade.Source(filepath.Join(fixtures, "fde_apps", "fde_8.0"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("builds and runs successfully", func() {
					deployment, _, err := platform.Deploy.Execute(name, fixture)
					Expect(err).NotTo(HaveOccurred())

					Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 8")))
				})
			})
		})

		context("deploying a self contained msbuild app with RuntimeIdentfier", func() {
			it.Before(func() {
				var err error
				fixture, err = switchblade.Source(filepath.Join(fixtures, "self_contained_apps", "msbuild"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("displays a simple text homepage", func() {
				deployment, logs, err := platform.Deploy.Execute(name, fixture)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(MatchRegexp("Removing dotnet-sdk"))
				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context("deploying .NET Core 8 self-contained app", func() {
			it.Before(func() {
				var err error
				fixture, err = switchblade.Source(filepath.Join(fixtures, "self_contained_apps", "self_contained_executable_8.0"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs successfully", func() {
				deployment, _, err := platform.Deploy.Execute(name, fixture)
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Hello, world!")))
			})
		})
	}
}
