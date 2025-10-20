#!/bin/bash

# Script to update proto imports to use hub-proto-contracts

set -e

echo "🔄 Updating proto imports to hub-proto-contracts..."

# List of files to update
files=(
    "./internal/balance/presentation/grpc/balance_grpc_handler.go"
    "./internal/market_data/presentation/grpc/market_data_grpc_handler.go"
    "./internal/order_mngmt_system/presentation/grpc/order_grpc_handler.go"
    "./internal/portfolio_summary/presentation/grpc/portfolio_grpc_handler.go"
    "./internal/position/presentation/grpc/position_grpc_handler.go"
    "./shared/grpc/auth_client.go"
    "./shared/grpc/auth_server.go"
    "./shared/grpc/grpc_integration_test.go"
    "./shared/grpc/order_client.go"
    "./shared/grpc/order_server.go"
    "./shared/grpc/position_client.go"
    "./shared/grpc/position_server.go"
    "./shared/grpc/server.go"
    "./shared/grpc/user_service_client.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "  Updating: $file"
        
        # Replace old import with new one
        sed -i '' 's|"HubInvestments/shared/grpc/proto"|authpb "github.com/RodriguesYan/hub-proto-contracts/auth"\n\tmonolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"|g' "$file"
        
        # Update proto. references to appropriate package
        # This is a simplified approach - may need manual adjustment
        sed -i '' 's/proto\.AuthService/authpb.AuthService/g' "$file"
        sed -i '' 's/proto\.UserService/authpb.UserService/g' "$file"
        sed -i '' 's/proto\.BalanceService/monolithpb.BalanceService/g' "$file"
        sed -i '' 's/proto\.MarketDataService/monolithpb.MarketDataService/g' "$file"
        sed -i '' 's/proto\.OrderService/monolithpb.OrderService/g' "$file"
        sed -i '' 's/proto\.PortfolioService/monolithpb.PortfolioService/g' "$file"
        sed -i '' 's/proto\.PositionService/monolithpb.PositionService/g' "$file"
        
        # Update message types
        sed -i '' 's/\*proto\./\*authpb./g' "$file"
        sed -i '' 's/proto\./monolithpb./g' "$file"
    fi
done

echo "✅ Import updates complete"
echo ""
echo "⚠️  Note: Some files may need manual adjustment for correct package references"

