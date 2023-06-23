package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testCache(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app            *cutlass.App
		Regexp         = `\[.*/dotnet-sdk.*\.tar\.xz\]`
		DownloadRegexp = "Download " + Regexp
		CopyRegexp     = "Copy " + Regexp
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "simple"))
		app.SetEnv("BP_DEBUG", "true")
		app.Buildpacks = []string{"dotnet_core_buildpack"}
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("uses the cache for manifest dependencies when deployed twice", func() {
		PushAppAndConfirm(t, app)
		Expect(app.Stdout.String()).To(MatchRegexp(DownloadRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(CopyRegexp))

		app.Stdout.Reset()
		PushAppAndConfirm(t, app)
		Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
	})
}
