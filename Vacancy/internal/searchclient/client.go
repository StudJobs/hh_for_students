// Package searchclient — best-effort клиент к Search-сервису для синхронной индексации.
// Если адрес не задан или сервис недоступен, операции логируются и не возвращают ошибку
// в основной хендлер: индексация не должна валить запрос на сохранение вакансии.
// Холодная переиндексация через make reindex компенсирует временные пропуски.
package searchclient

import (
	"context"
	"log"
	"time"

	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
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

func (c *Client) IndexVacancy(ctx context.Context, v *vacancyv1.Vacancy) {
	if c.cli == nil || v == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.IndexVacancy(cctx, &searchv1.IndexVacancyRequest{Vacancy: v}); err != nil {
		log.Printf("searchclient: index vacancy %s failed: %v", v.GetId(), err)
	}
}

func (c *Client) DeleteVacancy(ctx context.Context, id string) {
	if c.cli == nil || id == "" {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, indexTimeout)
	defer cancel()
	if _, err := c.cli.DeleteVacancy(cctx, &searchv1.DeleteDocumentRequest{Id: id}); err != nil {
		log.Printf("searchclient: delete vacancy %s failed: %v", id, err)
	}
}
