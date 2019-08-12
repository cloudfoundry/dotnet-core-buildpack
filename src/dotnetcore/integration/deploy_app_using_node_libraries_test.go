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
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
		Expect(page.Destroy()).To(Succeed())
	})

	Context("Deploying an angular app using msbuild and dotnet core 2.1", func() {
		BeforeEach(func() {
			SkipUnlessStack("cflinuxfs3")
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "angular_msbuild_dotnet_2.1"))
			app.Disk = "2G"
			app.Memory = "2G"
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
