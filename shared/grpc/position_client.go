package grpc

import (
	"context"
	"fmt"

	monolithpb "github.com/RodriguesYan/hub-proto-contracts/monolith"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type PositionClient struct {
	conn   *grpc.ClientConn
	client monolithpb.PositionServiceClient
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
	c.client = monolithpb.NewPositionServiceClient(conn)
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

func (c *PositionClient) GetPositions(token, userID string) (*monolithpb.GetPositionsResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.GetPositionsRequest{
		UserId: userID,
	}

	resp, err := c.client.GetPositions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.GetPositions failed: %w", err)
	}

	return resp, nil
}

func (c *PositionClient) GetAggregation(token, userID string) (*monolithpb.GetPositionAggregationResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	ctx, cancel := c.withAuth(token)
	defer cancel()

	req := &monolithpb.GetPositionAggregationRequest{
		UserId: userID,
	}

	resp, err := c.client.GetPositionAggregation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Position.GetPositionAggregation failed: %w", err)
	}

	return resp, nil
}

// Internal use for position updates
func (c *PositionClient) Create(token string, req *monolithpb.CreatePositionRequest) (*monolithpb.CreatePositionResponse, error) {
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
func (c *PositionClient) Update(token string, req *monolithpb.UpdatePositionRequest) (*monolithpb.UpdatePositionResponse, error) {
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
