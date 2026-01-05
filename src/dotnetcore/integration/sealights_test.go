package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSealights(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("deploying simple console app with binded Sealights service", func() {
			it("checks if Sealights installation was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"sealights-test-service": {
							"token":          "sometoken",
							"buildSessionId": "somesession",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainSubstring("Sealights. Service is enabled"))
				Expect(logs).To(ContainSubstring("Sealights. Agent is installed"))
				Expect(logs).To(ContainSubstring("Sealights. Service is set up"))
			})
		})
	}
}
