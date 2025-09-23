#!/bin/bash

# Script to generate Go code from proto files
# This script generates separate proto files for each service

set -e

PROTO_DIR="shared/grpc/proto"
OUTPUT_DIR="shared/grpc/proto"

echo "Generating proto files..."

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate common types
echo "Generating common.proto..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/common.proto"

# Generate auth service
echo "Generating auth_service.proto..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUTPUT_DIR" --go-grpc_opt=paths=source_relative \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/auth_service.proto"

# Generate order service
echo "Generating order_service.proto..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUTPUT_DIR" --go-grpc_opt=paths=source_relative \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/order_service.proto"

# Generate position service
echo "Generating position_service.proto..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUTPUT_DIR" --go-grpc_opt=paths=source_relative \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/position_service.proto"

echo "Proto generation completed!"
echo "Generated files:"
echo "  - $OUTPUT_DIR/common.pb.go"
echo "  - $OUTPUT_DIR/auth_service.pb.go"
echo "  - $OUTPUT_DIR/auth_service_grpc.pb.go"
echo "  - $OUTPUT_DIR/order_service.pb.go"
echo "  - $OUTPUT_DIR/order_service_grpc.pb.go"
echo "  - $OUTPUT_DIR/position_service.pb.go"
echo "  - $OUTPUT_DIR/position_service_grpc.pb.go"
