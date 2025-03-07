.PHONY: run lint

run:
	DEBUG=1 BLOCKS_PARSING_DEPTH=100 RPC_LIMIT=420 API_HOST='localhost:8080' go run .

# Run linter
lint:
	@which golangci-lint > /dev/null; if [ $$? -eq 0 ]; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi 
