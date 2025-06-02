FORMAT_FILES := grep -L -R "^\/\/ Code generated .* DO NOT EDIT\.$$" --exclude-dir=.git --exclude-dir=vendor --include="*.go" .

GOLANGCI_LINT_VERSION ?= v2.1.6

install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

	go install \
		github.com/daixiang0/gci \
		gotest.tools/gotestsum \
		mvdan.cc/gofumpt

lint:
	golangci-lint run --fix

mod:
	go mod tidy

semgrep:
	semgrep scan --config .semgrep/rules.yaml --config=p/semgrep-go-correctness

test:
	gotestsum -- -vet=off -race ./...
