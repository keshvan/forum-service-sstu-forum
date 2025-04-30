package client

import (
	"context"
	"fmt"
	"time"

	userpb "github.com/keshvan/protos-forum/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	client userpb.UserServiceClient
	conn   *grpc.ClientConn
}

func New(address string) (*UserClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("client.UserClient - New - grpc.NewClient: %w", err)
	}

	c := userpb.NewUserServiceClient(conn)

	return &UserClient{
		client: c,
		conn:   conn,
	}, nil
}

func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *UserClient) GetUsernames(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	if len(userIDs) == 0 {
		return make(map[int64]string), nil
	}

	req := &userpb.GetUsernamesRequest{
		UserIds: userIDs,
	}

	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := c.client.GetUsernames(callCtx, req)
	if err != nil {
		return nil, fmt.Errorf("clients.user - GetUsernames - c.client.GetUsernames: %w", err)
	}

	return res.GetUsernames(), nil
}
