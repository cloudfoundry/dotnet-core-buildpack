package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testVendored(t *testing.T, context spec.G, it spec.S) {
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "fdd_apps", "simple"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "self_contained_apps", "msbuild"))
}

func AssertNoInternetTraffic(t *testing.T, context spec.G, it spec.S, fixture string) {
	var Expect = NewWithT(t).Expect
	var app *cutlass.App

	context("when offline", func() {
		it.Before(func() {
			app = cutlass.New(fixture)
		})

		it.After(func() {
			app = DestroyApp(t, app)
		})

		it("displays a simple text homepage", func() {
			PushAppAndConfirm(t, app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})

		it("builds and runs the app", func() {
			root, err := cutlass.FindRoot()
			Expect(err).NotTo(HaveOccurred())

			bpFile := filepath.Join(root, settings.Buildpack.Version+"tmp")
			cmd := exec.Command("cp", settings.Buildpack.Path, bpFile)
			Expect(cmd.Run()).To(Succeed())
			defer os.Remove(bpFile)

			traffic, _, _, err := cutlass.InternetTraffic(fixture, bpFile, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(traffic).To(BeEmpty())
		})

	})
}
