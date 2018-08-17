package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("deploying a simple webapp written in F#", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "fsharp_msbuild"))
			app.Memory = "2G"
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World from F#!"))
		})
	})
})
