package integration_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App

	BeforeEach(SkipUnlessCached)

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
	})

	Context("The app is portable", func() {
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "asp_vendored"))
			} else {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "asp_vendored_dotnet2"))
			}
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		AssertNoInternetTraffic("asp_vendored")
	})

	Context("The app is self contained", func() {
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained"))
			} else {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained_dotnet2"))
			}
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		AssertNoInternetTraffic("self_contained")
	})
})
