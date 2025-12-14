# Build the provider
.PHONY: build
build:
	go build -v .

# Run tests
.PHONY: test
test:
	go test -v -cover ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Generate documentation using tfplugindocs
.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0 generate

# Check if documentation is up-to-date (fails if docs need regeneration)
.PHONY: docs-check
docs-check:
	@echo "Checking if documentation is up-to-date..."
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.24.0 generate
	@if [ -n "$$(git status --porcelain docs/)" ]; then \
		echo "ERROR: Documentation is out of date. Please run 'make docs' and commit the changes."; \
		git diff docs/; \
		exit 1; \
	fi
	@echo "Documentation is up-to-date."

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-vboxweb