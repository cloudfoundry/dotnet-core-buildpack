package integration_test

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/semver"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testVersions(platform switchblade.Platform, fixtures, root string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name           string
			source         string
			err            error
			latestVersions LatestVersions
		)

		it.Before(func() {
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())

			latestVersions, err = GetLatestVersions(root)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("when there is a global.json", func() {
			context("with an unsupported version", func() {
				context("and sdk doesn't exist at all", func() {
					it.Before(func() {
						source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_global_json"))
						Expect(err).NotTo(HaveOccurred())

						data, err := os.ReadFile(filepath.Join(source, "global.json"))
						Expect(err).ToNot(HaveOccurred())
						data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", "sdk_version")), []byte("99.99.99"), 1)
						Expect(os.WriteFile(filepath.Join(source, "global.json"), data, 0644)).To(Succeed())
					})

					it.After(func() {
						Expect(os.RemoveAll(source)).To(Succeed())
					})

					it("displays a nice error messages and gracefully fails", func() {
						_, logs, err := platform.Deploy.
							Execute(name, source)
						Expect(err).To(HaveOccurred())

						Expect(logs.String()).To(SatisfyAll(
							ContainSubstring("SDK 99.99.99 in global.json is not available"),
							ContainSubstring("Unable to install Dotnet SDK: could not find sdk in same feature line as '99.99.99'"),
						))
					})
				})

				context("and sdk doesn't exist but fall to latest in version line", func() {
					var (
						constructedVersion string
						baseFeatureLine    int
						proceed            bool
					)

					it.Before(func() {
						version, err := semver.NewVersion(latestVersions.latest6SDKVersion)
						Expect(err).ToNot(HaveOccurred())

						if version.Patch()%100 != 0 {
							proceed = true
						}

						baseFeatureLine = int((version.Patch() / 100) * 100)
						constructedVersion = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), baseFeatureLine)

						source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_global_json"))
						Expect(err).NotTo(HaveOccurred())

						data, err := os.ReadFile(filepath.Join(source, "global.json"))
						Expect(err).ToNot(HaveOccurred())
						data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", "sdk_version")), []byte(constructedVersion), 1)
						Expect(os.WriteFile(filepath.Join(source, "global.json"), data, 0644)).To(Succeed())

					})

					it("logs a warning about using source_apps SDK", func() {
						if proceed {
							deployment, logs, err := platform.Deploy.
								Execute(name, source)
							Expect(err).NotTo(HaveOccurred())

							Expect(logs.String()).To(SatisfyAll(
								ContainSubstring(fmt.Sprintf("SDK %s in global.json is not available", constructedVersion)),
								ContainSubstring("falling back to latest version in version line"),
							))

							Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
						}
					})
				})

			})

			context("with a supported version", func() {
				it.Before(func() {
					source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_global_json"))
					Expect(err).NotTo(HaveOccurred())

					data, err := os.ReadFile(filepath.Join(source, "global.json"))
					Expect(err).ToNot(HaveOccurred())
					data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", "sdk_version")), []byte(latestVersions.latest7SDKVersion), 1)
					Expect(os.WriteFile(filepath.Join(source, "global.json"), data, 0644)).To(Succeed())
				})

				it.After(func() {
					Expect(os.RemoveAll(source)).To(Succeed())
				})

				it("deploy successfully", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, source)
					Expect(err).NotTo(HaveOccurred())

					Expect(logs.String()).To(SatisfyAll(
						ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latestVersions.latest7SDKVersion)),
						ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latestVersions.latest7RuntimeVersion)),
					))

					Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
				})
			})
		})

		context("with buildpack.yml and global.json files", func() {
			context("when SDK version from buildpack.yml is not available", func() {
				it.Before(func() {
					source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "multi_version_sources"))
					Expect(err).NotTo(HaveOccurred())

					data, err := os.ReadFile(filepath.Join(source, "buildpack.yml"))
					Expect(err).ToNot(HaveOccurred())
					data = bytes.Replace(data, []byte(fmt.Sprintf("<%%= %s %%>", "sdk_version")), []byte("2.0.0-preview7"), 1)
					Expect(os.WriteFile(filepath.Join(source, "buildpack.yml"), data, 0644)).To(Succeed())
				})

				it("fails due to missing SDK", func() {
					_, logs, err := platform.Deploy.
						Execute(name, source)
					Expect(err).To(HaveOccurred())

					Expect(logs.String()).To(SatisfyAll(
						ContainSubstring("SDK 2.0.0-preview7 in buildpack.yml is not available"),
						ContainSubstring("Unable to install Dotnet SDK: no match found for 2.0.0-preview7"),
					))
				})
			})
		})

		context("when there is no global.json file", func() {
			it.Before(func() {
				source, err = switchblade.Source(filepath.Join(fixtures, "source_apps", "simple_global_json"))
				Expect(err).NotTo(HaveOccurred())

				Expect(os.Remove(filepath.Join(source, "global.json"))).To(Succeed())
			})

			it("deploys with default versions", func() {

				deployment, logs, err := platform.Deploy.
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring(fmt.Sprintf("Installing dotnet-sdk %s", latestVersions.latest6SDKVersion)),
					ContainSubstring(fmt.Sprintf("Installing dotnet-runtime %s", latestVersions.latest6RuntimeVersion)),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("Welcome to .NET 6")))
			})
		})
	}
}
