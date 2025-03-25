package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testCache(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

			name           string
			source         string
			Regexp         = `\[.*/dotnet-sdk.*\.tar\.xz\]`
			DownloadRegexp = "Download " + Regexp
			CopyRegexp     = "Copy " + Regexp
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

		it("uses the cache for manifest dependencies when deployed twice", func() {
			deploy := platform.Deploy.
				WithEnv(map[string]string{
					"BP_DEBUG": "true",
				})

			_, logs, err := deploy.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			Expect(logs).To(ContainLines(MatchRegexp(DownloadRegexp)))
			Expect(logs).NotTo(ContainLines(MatchRegexp(CopyRegexp)))

			_, logs, err = deploy.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			Expect(logs).NotTo(ContainLines(MatchRegexp(DownloadRegexp)))
			Expect(logs).To(ContainLines(MatchRegexp(CopyRegexp)))
		})
	}
}
