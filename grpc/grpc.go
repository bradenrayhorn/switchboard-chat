package grpc

import (
	"context"
	"github.com/bradenrayhorn/switchboard-protos/groups"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
)

type Client struct {
	groupService groups.GroupServiceClient
}

func NewClient() Client {
	client := Client{}
	// try to connect to core
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(viper.GetString("core_grpc_host"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect to core grpc: %s", err)
	}
	client.groupService = groups.NewGroupServiceClient(conn)
	return client
}

func (c *Client) GetGroups(userId string) ([]string, error) {
	resp, err := c.groupService.GetGroups(context.Background(), &groups.GetGroupsRequest{UserId: userId})
	if err != nil {
		return make([]string, 0), err
	}
	return resp.GetGroupIds(), nil
}
