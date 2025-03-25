package integration_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testDynatrace(platform switchblade.Platform, fixtures, uri string) func(*testing.T, spec.G, spec.S) {
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

		context("deploying a Dotnet Core app with Dynatrace agent with configured network zone", func() {
			it("checks if networkzone setting was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
							"networkzone":   "testzone",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent.")))
				Expect(logs).To(ContainLines(ContainSubstring("Starting Dynatrace OneAgent installer")))
				Expect(logs).To(ContainLines(ContainSubstring("Copy dynatrace-env.sh")))
				Expect(logs).To(ContainLines(ContainSubstring("Setting DT_NETWORK_ZONE...")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent installed.")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent injection is set up.")))
			})
		})

		context("when deploying with Dynatrace agent with single credentials service", func() {
			it("checks if Dynatrace injection was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent.")))
				Expect(logs).To(ContainLines(ContainSubstring("Starting Dynatrace OneAgent installer")))
				Expect(logs).To(ContainLines(ContainSubstring("Copy dynatrace-env.sh")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent installed.")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent injection is set up.")))
			})
		})

		context("when deploying with Dynatrace agent with two credentials services", func() {
			it("checks if detection of second service with credentials works", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
						"other-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("More than one matching service found!")))
			})
		})

		context("when deploying with Dynatrace agent with failing agent download and ignoring errors", func() {
			it("checks if skipping download errors works", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        fmt.Sprintf("%s/no-such-endpoint", uri),
							"environmentid": "envid",
							"skiperrors":    "true",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Download returned with status 404")))
				Expect(logs).To(ContainLines(ContainSubstring("Error during installer download, skipping installation")))
			})
		})

		context("deploying a with Dynatrace agent with two dynatrace services", func() {
			it("check if service detection isn't disturbed by a service with tags", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
						"dynatrace-tags": {
							"tag:dttest": "dynatrace_test",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent.")))
				Expect(logs).To(ContainLines(ContainSubstring("Starting Dynatrace OneAgent installer")))
				Expect(logs).To(ContainLines(ContainSubstring("Copy dynatrace-env.sh")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent installed.")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent injection is set up.")))
			})
		})

		context("deploying with Dynatrace agent with single credentials service and without manifest.json", func() {
			it("checks if Dynatrace injection was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent.")))
				Expect(logs).To(ContainLines(ContainSubstring("Starting Dynatrace OneAgent installer")))
				Expect(logs).To(ContainLines(ContainSubstring("Copy dynatrace-env.sh")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent installed.")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent injection is set up.")))
			})
		})

		context("deploying Dynatrace agent with failing agent download and checking retry", func() {
			it("checks if retrying downloads works", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        fmt.Sprintf("%s/no-such-endpoint", uri),
							"environmentid": "envid",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Expect(logs).To(ContainLines(ContainSubstring("Error during installer download, retrying in 4s")))
				Expect(logs).To(ContainLines(ContainSubstring("Error during installer download, retrying in 5s")))
				Expect(logs).To(ContainLines(ContainSubstring("Error during installer download, retrying in 7s")))
				Expect(logs).To(ContainLines(ContainSubstring("Download returned with status 404")))
			})
		})

		context("deploying Dynatrace agent with single credentials service and a redis service", func() {
			it("checks if Dynatrace injection was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
						"some-redis": {
							"name": "redis",
							"credentials": map[string]interface{}{
								"db_type": "redis",
								"instance_administration_api": map[string]interface{}{
									"deployment_id": "12345asdf",
									"instance_id":   "12345asdf",
									"root":          "https://doesnotexi.st",
								},
							},
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace service credentials found. Setting up Dynatrace OneAgent.")))
				Expect(logs).To(ContainLines(ContainSubstring("Starting Dynatrace OneAgent installer")))
				Expect(logs).To(ContainLines(ContainSubstring("Copy dynatrace-env.sh")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent installed.")))
				Expect(logs).To(ContainLines(ContainSubstring("Dynatrace OneAgent injection is set up.")))
			})
		})

		context("deploying Dynatrace agent with single credentials service", func() {
			it("checks if agent config update via API was successful", func() {
				_, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"BP_DEBUG": "true",
					}).
					WithServices(map[string]switchblade.Service{
						"some-dynatrace": {
							"apitoken":      "secretpaastoken",
							"apiurl":        uri,
							"environmentid": "envid",
						},
					}).
					Execute(name, filepath.Join(fixtures, "source_apps", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Fetching updated OneAgent configuration from tenant...")))
				Expect(logs).To(ContainLines(ContainSubstring("Finished writing updated OneAgent config back to")))
			})
		})
	}
}
