package grpc

import (
	"context"
	"fmt"

	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type OrderClient struct {
	conn   *grpc.ClientConn
	client monolithpb.OrderServiceClient
	config *ClientConfig
}

func NewOrderClient(config *ClientConfig) *OrderClient {
	return &OrderClient{
		config: config,
	}
}

func (c *OrderClient) connect() error {
	if c.conn != nil {
		return nil
	}

	conn, err := dialGRPC(c.config)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = monolithpb.NewOrderServiceClient(conn)
	return nil
}

func (c *OrderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *OrderClient) withAuth(token string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return ctx, cancel
}

func (c *OrderClient) Submit(token, userID, symbol, orderType, orderSide string, quantity float64, price *float64) (*monolithpb.SubmitOrderResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.SubmitOrderRequest{
		UserId:    userID,
		Symbol:    symbol,
		OrderType: orderType,
		OrderSide: orderSide,
		Quantity:  quantity,
	}

	if price != nil {
		req.Price = price
	}

	resp, err := c.client.SubmitOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Order.SubmitOrder failed: %w", err)
	}

	return resp, nil
}

func (c *OrderClient) GetStatus(token, orderID, userID string) (*monolithpb.GetOrderStatusResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.GetOrderStatusRequest{
		OrderId: orderID,
		UserId:  userID,
	}

	resp, err := c.client.GetOrderStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Order.GetOrderStatus failed: %w", err)
	}

	return resp, nil
}

func (c *OrderClient) GetDetails(token, orderID, userID string) (*monolithpb.GetOrderDetailsResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.GetOrderDetailsRequest{
		OrderId: orderID,
		UserId:  userID,
	}

	resp, err := c.client.GetOrderDetails(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Order.GetOrderDetails failed: %w", err)
	}

	return resp, nil
}

func (c *OrderClient) Cancel(token, orderID, userID string) (*monolithpb.CancelOrderResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	}

	resp, err := c.client.CancelOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Order.CancelOrder failed: %w", err)
	}

	return resp, nil
}
