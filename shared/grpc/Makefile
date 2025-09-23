# Makefile for gRPC proto generation

.PHONY: proto clean help

# Generate all proto files
proto:
	@echo "Generating proto files..."
	@./generate_proto.sh

# Clean generated proto files
clean:
	@echo "Cleaning generated proto files..."
	@rm -f proto/*.pb.go
	@rm -f proto/*_grpc.pb.go
	@echo "Cleaned generated files"

# Show help
help:
	@echo "Available targets:"
	@echo "  proto  - Generate Go code from proto files"
	@echo "  clean  - Remove generated proto files"
	@echo "  help   - Show this help message"
