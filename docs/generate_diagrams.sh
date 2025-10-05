#!/bin/bash

# Script to generate PNG images from Mermaid diagrams
# Make sure you have @mermaid-js/mermaid-cli installed:
# npm install -g @mermaid-js/mermaid-cli

echo "🎨 Generating architecture diagrams..."

# Check if mmdc is installed
if ! command -v mmdc &> /dev/null; then
    echo "❌ mermaid-cli not found. Please install it first:"
    echo "npm install -g @mermaid-js/mermaid-cli"
    exit 1
fi

# Navigate to docs directory
cd "$(dirname "$0")"

# Generate Order Flow Diagram
echo "📊 Generating Order Flow Diagram..."
mmdc -i order_flow_diagram.mmd -o order_flow_diagram.png -w 1200 -H 800 -b white

# Generate Microservices Overview
echo "🏗️ Generating Microservices Overview..."
mmdc -i microservices_overview.mmd -o microservices_overview.png -w 1400 -H 1000 -b white

# Generate Order Flow with Idempotency
echo "🔒 Generating Order Flow with Idempotency..."
mmdc -i order_flow_with_idempotency.mmd -o order_flow_with_idempotency.png -w 1400 -H 900 -b white

# Generate WebSocket Quotes Architecture
echo "📡 Generating WebSocket Quotes Architecture..."
mmdc -i websocket_quotes_architecture.mmd -o websocket_quotes_architecture.png -w 1400 -H 1000 -b white

# Generate Position Update Flow
echo "🔄 Generating Position Update Flow..."
mmdc -i position_update_flow.mmd -o position_update_flow.png -w 1600 -H 1200 -b white

# Generate Microservices Architecture
echo "🏛️ Generating Microservices Architecture..."
mmdc -i microservices_architecture.mmd -o microservices_architecture.png -w 1800 -H 1400 -b white

# Generate Microservices Event Flow
echo "📨 Generating Microservices Event Flow..."
mmdc -i microservices_event_flow.mmd -o microservices_event_flow.png -w 1600 -H 1200 -b white

echo "✅ Diagrams generated successfully!"
echo "📁 Files created:"
echo "   - order_flow_diagram.png"
echo "   - microservices_overview.png"
echo "   - order_flow_with_idempotency.png"
echo "   - websocket_quotes_architecture.png"
echo "   - position_update_flow.png"
echo "   - microservices_architecture.png"
echo "   - microservices_event_flow.png"
