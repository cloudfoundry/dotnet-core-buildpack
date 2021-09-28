package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testMultipleProjects(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		app        *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "source_apps", "multiple_projects_msbuild"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("compiles both apps", func() {
		PushAppAndConfirm(t, app)
		Expect(app.GetBody("/")).To(ContainSubstring("Hello, I'm a string!"))
		Eventually(app.Stdout.String, 10*time.Second).Should(ContainSubstring("Hello from a secondary project!"))
	})

	context("Deploying a self-contained solution with multiple projects", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "self_contained_apps", "self_contained_solution_2.2"))
		})
		it("can run the app", func() {
			PushAppAndConfirm(t, app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
}
