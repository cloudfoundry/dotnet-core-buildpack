package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNode(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		app    *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "node_apps", "angular_dotnet"))
		app.Disk = "2G"
		app.Memory = "2G"
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("deploying an angular app", func() {
		it("displays a simple text homepage", func() {
			PushAppAndConfirm(t, app)
			Expect(app.GetBody("/")).To(ContainSubstring("<title>source_app</title>"))
		})
	})
}
