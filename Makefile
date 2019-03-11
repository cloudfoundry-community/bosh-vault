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
	@go build -o bin/bosh-vault

.PHONY: package
package:
	./local-dev/tasks/package

.PHONY: vault
vault: ## Runs a pre-configured local vault server for development purposes
	./local-dev/tasks/run-local-vault

.PHONY: unseal
unseal: ## Unseal's the local vault server
	./local-dev/tasks/unseal-local-vault

.PHONY: reset-vault
reset-vault: ## Resets the development Vault to default state
	./local-dev/tasks/reset-vault

.PHONY: run
run: bin/blite build ## Builds and runs the binary using local-dev settings (requires running bosh-lite setup)
	./local-dev/tasks/run-local-binary

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
	./local-dev/tasks/bootstrap-local-director

test-deploy-redis: ## Tries to deploy redis with a generated password on the bosh-lite director
	./local-dev/tasks/test-deploy-redis

test-deploy-nginx: ## Tries to deploy NGINX and host a page filled with all the types of secrets that can be generated
	./local-dev/tasks/test-deploy-nginx

local-vars: local-dev/vars/local-dev-vars.yml

local-dev/vars/local-dev-vars.yml:
	./local-dev/tasks/generate-local-dev-vars-file

local-certs: local-dev/certs/local-dev.crt

local-dev/certs/local-dev.crt:
	./local-dev/tasks/generate-local-dev-certs

destroy: reset-vault ## Burns down local dev environment
	-rm -r ./local-dev/certs/*
	-rm -r ./local-dev/vars/*
	-blite destroy
	-pkill bosh-vault
