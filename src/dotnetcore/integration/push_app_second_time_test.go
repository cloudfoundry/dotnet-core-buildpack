package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		SkipUnlessUncached()

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "fsharp_msbuild"))
		app.SetEnv("BP_DEBUG", "true")
	})

	DownloadRegexp := `Download \[.*/dotnet\..*\.tar\.xz\]`
	CopyRegexp := `Copy \[.*/dotnet\..*\.tar\.xz\]`

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
