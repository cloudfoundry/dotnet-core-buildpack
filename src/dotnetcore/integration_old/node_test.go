package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/agouti"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNode(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		app        *cutlass.App

		agoutiDriver *agouti.WebDriver
		page         *agouti.Page
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "node_apps", "angular_dotnet"))
		app.Disk = "2G"
		app.Memory = "2G"

		var err error
		agoutiDriver = agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}))
		err = agoutiDriver.Start()
		Expect(err).To(BeNil())

		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		app = DestroyApp(t, app)
		Expect(page.Destroy()).To(Succeed())

		err := agoutiDriver.Stop()
		Expect(err).To(BeNil())
	})

	context("deploying an angular app", func() {
		it("displays a simple text homepage", func() {
			PushAppAndConfirm(t, app)
			url, err := app.GetUrl("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(page.Navigate(url)).To(Succeed())
			Eventually(page.HTML, 30*time.Second).Should(ContainSubstring("Hello, world!"))
		})
	})
}
