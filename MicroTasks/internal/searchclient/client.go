// Package searchclient — best-effort клиент к Search-сервису для синхронной индексации микрозадач.
package searchclient

import (
	"context"
	"log"
	"time"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const indexTimeout = 5 * time.Second

type Client struct {
	conn *grpc.ClientConn
	cli  searchv1.SearchServiceClient
}

func New(addr string) *Client {
	if addr == "" {
		log.Printf("searchclient: SEARCH_GRPC_ADDR is empty, indexing disabled")
		return &Client{}
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("searchclient: dial %s failed: %v (indexing disabled)", addr, err)
		return &Client{}
	}
	return &Client{conn: conn, cli: searchv1.NewSearchServiceClient(conn)}
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) IndexTask(ctx context.Context, t *microtaskv1.MicroTask) {
	if c.cli == nil || t == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.IndexMicroTask(cctx, &searchv1.IndexMicroTaskRequest{Task: t}); err != nil {
		log.Printf("searchclient: index microtask %s failed: %v", t.GetId(), err)
	}
}

func (c *Client) DeleteTask(ctx context.Context, id string) {
	if c.cli == nil || id == "" {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.DeleteMicroTask(cctx, &searchv1.DeleteDocumentRequest{Id: id}); err != nil {
		log.Printf("searchclient: delete microtask %s failed: %v", id, err)
	}
}
