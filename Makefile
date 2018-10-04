# Help Helper matches comments at the start of the task block so make help gives users information about each task
.PHONY: help
help: ## Displays information about available make tasks
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

init: bin/blite ## Installs developer dependencies

build: fmt ## Builds the binary
	@go build -o bin/vault-cfcs

run: build ## Run the binary (also builds)
	./bin/vault-cfcs

.PHONY: fmt
fmt: ## Runs gofmt on the entire project
	@go fmt ./...

bin/blite:
	@curl -o bin/blite https://raw.githubusercontent.com/Zipcar/blite/master/blite
