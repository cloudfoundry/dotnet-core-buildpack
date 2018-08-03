package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/agouti"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		app = DestroyApp(app)
		Expect(page.Destroy()).To(Succeed())
	})

	Context("Deploying an angular app using msbuild and dotnet 1.X", func() {
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs2")
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "angular_msbuild_dotnet1"))
		})

		It("displays a javascript homepage", func() {
			PushAppAndConfirm(app)

			url, err := app.GetUrl("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(page.Navigate(url)).To(Succeed())
			Eventually(page.HTML, 30*time.Second).Should(ContainSubstring("My First Angular 2 App"))
		})
	})

	Context("Deploying an angular app using msbuild and dotnet 2.X", func() {
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs3")
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "angular_msbuild_dotnet2"))
		})

		It("displays a javascript homepage", func() {
			PushAppAndConfirm(app)

			url, err := app.GetUrl("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(page.Navigate(url)).To(Succeed())
			Eventually(page.HTML, 30*time.Second).Should(ContainSubstring("Hello, world from Dotnet Core 2.1"))
		})
	})
})
