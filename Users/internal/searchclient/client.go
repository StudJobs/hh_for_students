// Package searchclient — best-effort клиент к Search-сервису для синхронной индексации.
// Если адрес не задан или сервис недоступен, операции логируются и не возвращают ошибку
// в основной хендлер: индексация не должна валить запрос на сохранение профиля.
// Холодная переиндексация через make reindex компенсирует временные пропуски.
package searchclient

import (
	"context"
	"log"
	"time"

	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const indexTimeout = 5 * time.Second

type Client struct {
	conn *grpc.ClientConn
	cli  searchv1.SearchServiceClient
}

// New создаёт клиент. Если addr пустой — возвращает нерабочий клиент (no-op).
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

func (c *Client) IndexProfile(ctx context.Context, p *usersv1.Profile) {
	if c.cli == nil || p == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.IndexProfile(cctx, &searchv1.IndexProfileRequest{Profile: p}); err != nil {
		log.Printf("searchclient: index profile %s failed: %v", p.GetId(), err)
	}
}

func (c *Client) DeleteProfile(ctx context.Context, id string) {
	if c.cli == nil || id == "" {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.DeleteProfile(cctx, &searchv1.DeleteDocumentRequest{Id: id}); err != nil {
		log.Printf("searchclient: delete profile %s failed: %v", id, err)
	}
}
