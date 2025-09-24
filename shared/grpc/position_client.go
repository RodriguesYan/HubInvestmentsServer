package grpc

import (
	"context"
	"fmt"

	"HubInvestments/shared/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type PositionClient struct {
	conn   *grpc.ClientConn
	client proto.PositionServiceClient
	config *ClientConfig
}

func NewPositionClient(config *ClientConfig) *PositionClient {
	return &PositionClient{
		config: config,
	}
}

func (c *PositionClient) connect() error {
	if c.conn != nil {
		return nil
	}

	conn, err := dialGRPC(c.config)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = proto.NewPositionServiceClient(conn)
	return nil
}

func (c *PositionClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *PositionClient) withAuth(token string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return ctx, cancel
}

func (c *PositionClient) GetPositions(token, userID string) (*proto.GetPositionsResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &proto.GetPositionsRequest{
		UserId: userID,
	}

	resp, err := c.client.GetPositions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.GetPositions failed: %w", err)
	}

	return resp, nil
}

func (c *PositionClient) GetAggregation(token, userID string) (*proto.GetPositionAggregationResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &proto.GetPositionAggregationRequest{
		UserId: userID,
	}

	resp, err := c.client.GetPositionAggregation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.GetPositionAggregation failed: %w", err)
	}

	return resp, nil
}

// Internal use for position updates
func (c *PositionClient) Create(token string, req *proto.CreatePositionRequest) (*proto.CreatePositionResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	resp, err := c.client.CreatePosition(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.CreatePosition failed: %w", err)
	}

	return resp, nil
}

// Internal use for position updates
func (c *PositionClient) Update(token string, req *proto.UpdatePositionRequest) (*proto.UpdatePositionResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	resp, err := c.client.UpdatePosition(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.UpdatePosition failed: %w", err)
	}

	return resp, nil
}
