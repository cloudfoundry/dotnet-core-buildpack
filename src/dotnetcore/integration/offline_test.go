package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name   string
			source string
			err    error
		)

		it.Before(func() {
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("when pushing a simple app", func() {
			it.Before(func() {
				var err error
				name, err = switchblade.RandomName()
				Expect(err).NotTo(HaveOccurred())

				source, err = switchblade.Source(filepath.Join(fixtures, "vendored", "fdd_dotnet_6"))
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				Expect(platform.Delete.Execute(name)).To(Succeed())
			})

			it.Focus("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				Expect(logs.String()).To(SatisfyAll(
					ContainSubstring("Supplying Dotnet Core"),
					MatchRegexp(CopyRegexp),
					Not(MatchRegexp(DownloadRegexp)),
				))

				Eventually(deployment).Should(Serve(ContainSubstring("building Web apps with ASP.NET Core")))
			})
		})
	}
}
