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

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "nancy_kestrel_msbuild"))
	})

	It("displays a simple text homepage", func() {
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("Hello from Nancy running on CoreCLR"))
	})
})
