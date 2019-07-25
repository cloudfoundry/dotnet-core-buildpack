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
		var fixture string
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				fixture = "fdd_asp_vendored_1.0"
			} else {
				fixture = "fdd_asp_vendored_2.1"
			}
			app = cutlass.New(filepath.Join(bpDir, "fixtures", fixture))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		AssertNoInternetTraffic(fixture)
	})

	Context("The app is self contained", func() {
		var fixture string
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				fixture = "self_contained_1.0"
			} else {
				fixture = "self_contained_2.1"
			}
			app = cutlass.New(filepath.Join(bpDir, "fixtures", fixture))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		AssertNoInternetTraffic(fixture)
	})

	Context("The app is self contained and a preview version", func() {
		var fixture string
		BeforeEach(func() {
			if os.Getenv("CF_STACK") == "cflinuxfs2" {
				Skip("Dotnet3 only works on cflinuxfs3")
			}

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "self_contained_3.0_preview"))
			app.Disk = "2G"
		})

		It("displays a simple text homepage", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Welcome"))
		})

		AssertNoInternetTraffic(fixture)
	})
})
