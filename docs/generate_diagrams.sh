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

echo "✅ Diagrams generated successfully!"
echo "📁 Files created:"
echo "   - order_flow_diagram.png"
echo "   - microservices_overview.png"
