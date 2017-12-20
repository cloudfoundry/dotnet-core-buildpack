package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var (
		app         *cutlass.App
		fixtureName string
	)
	AfterEach(func() { app = DestroyApp(app) })

	JustBeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", fixtureName))
	})

	Context("Deploying an app with multiple projects", func() {
		BeforeEach(func() {
			fixtureName = "multiple_projects_msbuild"
		})
		It("compiles both apps", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello, I'm a string!"))
			Eventually(app.Stdout.String, 10*time.Second).Should(ContainSubstring("Hello from a secondary project!"))
		})
	})
})
