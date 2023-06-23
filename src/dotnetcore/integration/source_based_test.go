package integration_test

import (
	"github.com/cloudfoundry/switchblade"
	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"path/filepath"
	"testing"
)

type LatestVersions struct {
	latest6RuntimeVersion string
	latest6ASPNetVersion  string
	latest6SDKVersion     string
	latest7RuntimeVersion string
	latest7SDKVersion     string
}

func testSourceBased(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name   string
			source string
			err    error
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple"))

			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		it("builds and runs the app", func() {
			deployment, logs, err := platform.Deploy.
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			Expect(logs.String()).To(SatisfyAll(
				ContainSubstring("Supplying Dotnet Core"),
			))

			Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
		})

		context("when an app has a Microsoft.AspNetCore.App", func() {
			context("with version 6", func() {
				it.Before(func() {
					source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "aspnet_package_reference"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("publishes and runs, installing the correct runtime and aspnetcore version with a warning", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, source)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs.String()).To(SatisfyAll(
						ContainSubstring("Supplying Dotnet Core"),
						ContainSubstring("A PackageReference to Microsoft.AspNetCore.App is not necessary when targeting .NET Core 3.0 or higher."),
					))

					Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
				})
			})
		})

		context("when the app has Microsoft.AspNetCore.All", func() {
			context("with version 6", func() {
				it.Before(func() {
					source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "dotnet_6"))
					Expect(err).NotTo(HaveOccurred())
				})

				it("publishes and runs, installing the correct runtime and aspnetcore version with a warning", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, source)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs.String()).To(SatisfyAll(
						ContainSubstring("Supplying Dotnet Core"),
						ContainSubstring("A PackageReference to Microsoft.AspNetCore.App is not necessary when targeting .NET Core 3.0 or higher."),
					))

					Eventually(deployment).Should(Serve(ContainSubstring("building Web apps with ASP.NET Core")))
				})
			})
		})

		context("with AssemblyName specified", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "dot_in_name"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("successfully pushes an app with an AssemblyName", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context("with libgdiplus", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "libgdiplus"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("displays a simple text homepage", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
					ContainSubstring("Installing libgdiplus"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Hello World!")))
			})
		})

		context.Focus("with .NET Core 6", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "source_6.0"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs successfully", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
			})
		})

		context.Focus("with .NET Core 7", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "dotnet_7"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("builds and runs successfully", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("<title>source_app_7</title>")))
			})
		})
	}
}
