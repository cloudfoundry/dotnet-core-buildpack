package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	const (
		DownloadRegexp = `Download \[.*/dotnet-sdk.*\.tar\.xz\]`
		CopyRegexp     = `Copy \[.*/dotnet-sdk.*\.tar\.xz\]`
	)

	var app *cutlass.App

	BeforeEach(func() {
		SkipUnlessUncached()

		app = cutlass.New(Fixtures("source_2.1_float_runtime"))
		app.SetEnv("BP_DEBUG", "true")
		app.Buildpacks = []string{"dotnet_core_buildpack"}
	})

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	It("uses the cache for manifest dependencies", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp(DownloadRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(CopyRegexp))

		app.Stdout.Reset()
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
	})
})
