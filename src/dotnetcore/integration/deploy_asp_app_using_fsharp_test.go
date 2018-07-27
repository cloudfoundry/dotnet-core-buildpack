package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("deploying a simple webapp written in F#", func() {
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs3" {
				Skip("dotnet 1.0.x SDK and Framework are not supported on cflinuxfs3")
			}
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "fsharp_msbuild"))
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(ContainSubstring("Installing dotnet 1."))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World from F#!"))
		})
	})
})
