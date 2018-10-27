# Help Helper matches comments at the start of the task block so make help gives users information about each task
.PHONY: help
help: ## Displays information about available make tasks
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init: ## Installs developer dependencies
	@go get github.com/onsi/ginkgo/ginkgo
	@go get github.com/onsi/gomega/...

.PHONY: build
build: fmt ## Builds the binary
	@go build -o bin/vault-cfcs

.PHONY: run
run: build ## Builds and runs the binary using local-dev settings (requires running bosh-lite setup)
	./tasks/run-local-binary

.PHONY: fmt
fmt: ## Runs gofmt on the entire project
	@go fmt ./...

.PHONY: test
test: init fmt ## Runs all test suites with ginkgo
	@ginkgo -v -p --randomizeAllSpecs --randomizeSuites --succinct * */* */*/*

bin/blite:
	@mkdir -p bin
	@curl -o bin/blite https://raw.githubusercontent.com/Zipcar/blite/master/blite
	@chmod +x bin/blite

.PHONY: bosh-lite
bosh-lite: bin/blite local-certs local-vars ## Spin up a local bosh director with UAA that is ready to communicate with the local binary
	./tasks/bootstrap-local-director

test-deploy-redis: ## Tries to deploy redis with a generated password on the bosh-lite director
	./tasks/test-deploy-redis

local-vars: local-dev/vars/local-dev-vars.yml

local-dev/vars/local-dev-vars.yml:
	./tasks/generate-local-dev-vars-file

local-certs: local-dev/certs/local-dev.crt

local-dev/certs/local-dev.crt:
	./tasks/generate-local-dev-certs

destroy: ## Burns down local dev environment
	-rm -r ./local-dev/certs/*
	-rm -r ./local-dev/vars/*
	-blite destroy
	-pkill vault-cfcs
