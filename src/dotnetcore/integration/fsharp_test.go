package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFsharp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		app    *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "fsharp"))
		app.Memory = "2G"
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("deploying an fsharp app", func() {
		it("displays a simple text homepage", func() {
			PushAppAndConfirm(t, app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World from F#!"))
		})
	})
}
