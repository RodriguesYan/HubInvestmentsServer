#!/bin/bash

# Script to generate Go code from proto files
# This script generates separate proto files for each service

set -e

PROTO_DIR="proto"
OUTPUT_DIR="proto"

echo "Generating proto files..."

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate all proto files together to handle cross-references properly
echo "Generating all proto files together..."
protoc --go_out="$OUTPUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUTPUT_DIR" --go-grpc_opt=paths=source_relative \
    --proto_path="$PROTO_DIR" \
    "$PROTO_DIR/common.proto" \
    "$PROTO_DIR/auth_service.proto" \
    "$PROTO_DIR/order_service.proto" \
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
