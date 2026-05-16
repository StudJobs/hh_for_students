// Package usersclient — best-effort клиент к Users-сервису для добавления verified-навыков.
package usersclient

import (
	"context"
	"log"
	"time"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const callTimeout = 5 * time.Second

type Client struct {
	conn *grpc.ClientConn
	cli  usersv1.UsersServiceClient
}

func New(addr string) *Client {
	if addr == "" {
		log.Printf("usersclient (achievements): addr empty, AddVerifiedSkills disabled")
		return &Client{}
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("usersclient (achievements): dial %s failed: %v", addr, err)
		return &Client{}
	}
	return &Client{conn: conn, cli: usersv1.NewUsersServiceClient(conn)}
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) AddVerifiedSkills(ctx context.Context, userID string, slugs []string) {
	if c.cli == nil || userID == "" || len(slugs) == 0 {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	if _, err := c.cli.AddVerifiedSkills(cctx, &usersv1.AddVerifiedSkillsRequest{
		UserId:     userID,
		SkillSlugs: slugs,
	}); err != nil {
		log.Printf("usersclient (achievements): AddVerifiedSkills user=%s skills=%v failed: %v", userID, slugs, err)
	}
}
