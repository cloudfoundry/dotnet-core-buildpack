GO_CMD=go
GO_GET=$(GO_CMD) get
GOLINT_CMD=golint
GO_TEST=$(GO_CMD) test -v ./...
GO_VET=$(GO_CMD) vet ./...
GO_LINT=$(GOLINT_CMD) .
GO_VET_SETUP=$(GO_GET) code.google.com/p/go.tools/cmd/vet
GO_LINT_SETUP=$(GO_GET) github.com/golang/lint/golint
GO_MINIASSERT_SETUP=$(GO_GET) github.com/pilu/miniassert

all:
	$(GO_VET)
	$(GO_LINT)
	$(GO_TEST)

setup:
	$(GO_VET_SETUP)
	$(GO_LINT_SETUP)
	$(GO_MINIASSERT_SETUP)
