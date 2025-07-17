##@ General

.PHONY: help
help: ## ❓ Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

COMPOSE_PROFILES ?= bff-service

.PHONY: install
install: ## 🔨 Install go tools
	@(cd .devcontainer && go list -e -f '{{range .Imports}}{{.}} {{end}}' tools.go | CGO_ENABLED=0 xargs go install -mod=readonly)

.PHONY: generate
generate: ## 🪄  Generate docs, mocks, stubs, ...
	@go generate ./...

.PHONY: lint
lint: ## 🚨 Format, detect lint issues, fix and report
	@gofumpt -w .
	@golangci-lint run --fix ./...

.PHONY: test
test: ## ✅ Run unit tests
	@go test -race ./... 2>&1 -coverprofile cover-raw.out
	@echo "mode: set" > cover.out
	@grep -E "/internal/" cover-raw.out | grep -v "/mocks/" | grep -v "_mock.go" >> cover.out
	@go test -json ./... > report.json
	@go tool cover -func cover.out | grep "total:"

.PHONY: coverage
coverage: ## 📊 Show detailed coverage report with percentages
	@echo "Generating coverage report..."
	@go test -coverprofile=cover-raw.out ./...
	@echo "mode: set" > cover.out
	@grep -E "/internal/" cover-raw.out | grep -v "/mocks/" | grep -v "_mock.go" >> cover.out
	
	@echo "\n======================= COVERAGE SUMMARY ========================"
	@echo "Total coverage (internal packages, excluding mocks):"
	@go tool cover -func=cover.out | grep "total:" | awk '{printf "  %s\n", $$3}'
	
	@echo "\n================== FILES NEEDING COVERAGE ======================="
	@go tool cover -func=cover.out | grep -v "100.0%" | grep -v "total:" | grep -v "/mocks/" | grep -v "_mock.go" \
		| awk '{ print $$1, $$NF }' | sort -k2 -n | head -15 | awk '{ printf "  %-70s %s\n", $$1, $$2 }'

.PHONY: coveragehtml
coveragehtml: ## 📊 Generate enhanced coverage report with line highlighting
	@echo "Generating HTML coverage report..."
	@mkdir -p coverage-report
	
	@echo "Running tests with coverage..."
	@go test -coverprofile=cover-raw.out ./...
	@echo "mode: set" > cover.out
	@grep -E "/internal/" cover-raw.out | grep -v "/mocks/" | grep -v "_mock.go" >> cover.out
	
	@echo "Creating HTML report with line highlighting..."
	@go tool cover -html=cover.out -o coverage-report/coverage.html
	
	@open coverage-report/coverage.html

.PHONY: gensql
gensql: ## 🪄  Generate sql files for schema migration, i.e: make gensql NAME=create_users_table
	@migrate create -ext sql -dir cmd/migration/migrations $(NAME)

##@ Deployment

.PHONY: build
build: ## 🐳 Build container image
	@docker build --ssh default=$$SSH_AUTH_SOCK .

.PHONY: up
up: ## 🐳 Build and deploy containers using Docker Compose
	@COMPOSE_PROFILE=$(COMPOSE_PROFILE) docker compose up --build --remove-orphans --force-recreate

.PHONY: down
down: ## 🐳 Destroy containers without volumes using Docker Compose
	@COMPOSE_PROFILE=$(COMPOSE_PROFILE) docker compose down --rmi local

.PHONY: downv
downv: ## 🐳 Destroy containers with volumes using Docker Compose
	@COMPOSE_PROFILE=$(COMPOSE_PROFILE) docker compose down -v --rmi local

.PHONY: watch
watch: ## 🐳 Develop with watch mode using Docker Compose
	@COMPOSE_PROFILE=$(COMPOSE_PROFILE) docker compose watch

.PHONY: ps
ps: ## 🐳 List all compose services
	@COMPOSE_PROFILE=$(COMPOSE_PROFILE) docker compose ps -a

.PHONY: deps
deps: ## 📦 Update go modules
	@go mod tidy && go mod vendor
	
	@# Parse the cover.out file to extract coverage data
	@echo "Generating coverage data..."
	@echo "[" > coverage-report/coverage-data.json
	@go tool cover -func=cover.out | grep -v "total:" | while read -r line; do \
		file=$$(echo "$$line" | awk '{print $$1}'); \
		statements=$$(echo "$$line" | grep -oP "\d+(?= statements)"); \
		covered=$$(echo "$$line" | grep -oP "\d+(?= statements covered)"); \
		coverage=$$(echo "$$line" | grep -oP "\d+\.\d+(?=%)"); \
		if [ ! -z "$$file" ] && [ ! -z "$$statements" ] && [ ! -z "$$covered" ] && [ ! -z "$$coverage" ]; then \
			echo "{ \"filename\": \"$$file\", \"statements\": $$statements, \"covered\": $$covered, \"coverage\": $$coverage }," >> coverage-report/coverage-data.json; \
		fi; \
	done
	@sed -i.bak '$$s/,$//' coverage-report/coverage-data.json
	@echo "]" >> coverage-report/coverage-data.json
	@rm -f coverage-report/coverage-data.json.bak
	
	@echo "Coverage reports generated at coverage-report/"
	@echo "- index.html: PHPUnit-like coverage summary"
	@echo "- coverage-details.html: Detailed line-by-line coverage"
	@open coverage-report/index.html