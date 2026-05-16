package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChatRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *ChatRepository) Insert(ctx context.Context, threadID, fromUser, body string) (*chatv1.Message, error) {
	query, args, err := r.sb.
		Insert("chat_messages").
		Columns("thread_id", "from_user_id", "body").
		Values(threadID, fromUser, body).
		Suffix("RETURNING id, thread_id, from_user_id, body, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}
	var m chatv1.Message
	var createdAt time.Time
	if err := r.db.QueryRow(ctx, query, args...).Scan(&m.Id, &m.ThreadId, &m.FromUserId, &m.Body, &createdAt); err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}
	m.CreatedAt = createdAt.Format(time.RFC3339)
	return &m, nil
}

func (r *ChatRepository) ListByThread(ctx context.Context, threadID string, page, limit int32) (*chatv1.MessageList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	query, args, err := r.sb.
		Select("id", "thread_id", "from_user_id", "body", "created_at").
		From("chat_messages").
		Where(squirrel.Eq{"thread_id": threadID}).
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var msgs []*chatv1.Message
	for rows.Next() {
		var m chatv1.Message
		var createdAt time.Time
		if err := rows.Scan(&m.Id, &m.ThreadId, &m.FromUserId, &m.Body, &createdAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		m.CreatedAt = createdAt.Format(time.RFC3339)
		msgs = append(msgs, &m)
	}

	var total int32
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM chat_messages WHERE thread_id=$1`, threadID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}
	pages := int32(0)
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}
	return &chatv1.MessageList{
		Messages: msgs,
		Pagination: &commonv1.PaginationResponse{
			Total:       total,
			Pages:       pages,
			CurrentPage: page,
		},
	}, nil
}
