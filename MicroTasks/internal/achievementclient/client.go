// Package achievementclient — best-effort клиент к Achievements-сервису.
// F5: при approve микрозадачи автоматически создаём ачивку в портфолио студента.
// Падение клиента не блокирует approve (write-through best effort) — log и продолжаем.
package achievementclient

import (
	"context"
	"log"
	"time"

	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const callTimeout = 5 * time.Second

type Client struct {
	conn *grpc.ClientConn
	cli  achievementv1.AchievementServiceClient
}

func New(addr string) *Client {
	if addr == "" {
		log.Printf("achievementclient: ACHIEVEMENTS_GRPC_ADDR is empty, autopopulate disabled")
		return &Client{}
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("achievementclient: dial %s failed: %v (autopopulate disabled)", addr, err)
		return &Client{}
	}
	return &Client{conn: conn, cli: achievementv1.NewAchievementServiceClient(conn)}
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// CreateMicrotaskAchievement — best-effort: ошибки логируются и не пробрасываются вверх,
// чтобы approve микрозадачи не падал из-за недоступности Achievements-сервиса.
// Идемпотентность гарантирована на стороне Achievements (UNIQUE по s3_key).
func (c *Client) CreateMicrotaskAchievement(
	ctx context.Context,
	userUUID, microtaskID, microtaskTitle, solutionURL, reviewerUUID, reviewComment string,
) {
	if c.cli == nil || userUUID == "" || microtaskID == "" {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	req := &achievementv1.CreateMicrotaskAchievementRequest{
		UserUuid:       userUUID,
		MicrotaskId:    microtaskID,
		MicrotaskTitle: microtaskTitle,
		SolutionUrl:    solutionURL,
		ReviewerUuid:   reviewerUUID,
		ReviewComment:  reviewComment,
	}
	if _, err := c.cli.CreateMicrotaskAchievement(cctx, req); err != nil {
		log.Printf("achievementclient: create microtask achievement (user=%s, microtask=%s) failed: %v",
			userUUID, microtaskID, err)
	}
}
