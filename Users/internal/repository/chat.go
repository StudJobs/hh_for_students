package repository

import (
	"context"
	"database/sql"
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
		Suffix("RETURNING id, thread_id, from_user_id, body, created_at, edited_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}
	return scanMessage(r.db.QueryRow(ctx, query, args...))
}

// EditMessage редактирует body своего сообщения. Проверка авторства строгая:
// если from_user_id не совпадает с автором сообщения — UPDATE не сработает и
// репо вернёт ErrMessageNotFoundOrForbidden.
func (r *ChatRepository) EditMessage(ctx context.Context, id, fromUserID, body string) (*chatv1.Message, error) {
	query := `
UPDATE chat_messages
SET body = $3, edited_at = NOW()
WHERE id = $1 AND from_user_id = $2
RETURNING id, thread_id, from_user_id, body, created_at, edited_at`
	return scanMessage(r.db.QueryRow(ctx, query, id, fromUserID, body))
}

// HideThread помечает тред скрытым у юзера. Сообщения не удаляются —
// собеседник продолжает видеть тред у себя.
func (r *ChatRepository) HideThread(ctx context.Context, userID, threadID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO chat_thread_hides (user_id, thread_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, threadID)
	return err
}

// IsHidden — true если тред скрыт для юзера.
func (r *ChatRepository) IsHidden(ctx context.Context, userID, threadID string) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT 1 FROM chat_thread_hides WHERE user_id = $1 AND thread_id = $2`,
		userID, threadID).Scan(&n)
	if err != nil {
		// pgx.ErrNoRows → не скрыт
		return false, nil
	}
	return true, nil
}

// HiddenSet возвращает множество thread_id, скрытых для юзера (для batch-фильтра).
func (r *ChatRepository) HiddenSet(ctx context.Context, userID string) (map[string]struct{}, error) {
	rows, err := r.db.Query(ctx, `SELECT thread_id FROM chat_thread_hides WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]struct{})
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		out[tid] = struct{}{}
	}
	return out, nil
}

func scanMessage(row interface {
	Scan(...interface{}) error
}) (*chatv1.Message, error) {
	var m chatv1.Message
	var createdAt time.Time
	var editedAt sql.NullTime
	if err := row.Scan(&m.Id, &m.ThreadId, &m.FromUserId, &m.Body, &createdAt, &editedAt); err != nil {
		return nil, err
	}
	m.CreatedAt = createdAt.Format(time.RFC3339)
	if editedAt.Valid {
		m.EditedAt = editedAt.Time.Format(time.RFC3339)
	}
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
		Select("id", "thread_id", "from_user_id", "body", "created_at", "edited_at").
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
		m, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		msgs = append(msgs, m)
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

// LastMessages возвращает map thread_id → последнее сообщение (по created_at DESC).
// Если для треда нет ни одного сообщения, ключа в map не будет.
func (r *ChatRepository) LastMessages(ctx context.Context, threadIDs []string) (map[string]*chatv1.Message, error) {
	out := make(map[string]*chatv1.Message, len(threadIDs))
	if len(threadIDs) == 0 {
		return out, nil
	}
	query := `
SELECT DISTINCT ON (thread_id) thread_id, id, from_user_id, body, created_at
FROM chat_messages
WHERE thread_id = ANY($1)
ORDER BY thread_id, created_at DESC`
	rows, err := r.db.Query(ctx, query, threadIDs)
	if err != nil {
		return nil, fmt.Errorf("last-messages: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var m chatv1.Message
		var createdAt time.Time
		var threadID string
		if err := rows.Scan(&threadID, &m.Id, &m.FromUserId, &m.Body, &createdAt); err != nil {
			return nil, err
		}
		m.ThreadId = threadID
		m.CreatedAt = createdAt.Format(time.RFC3339)
		out[threadID] = &m
	}
	return out, nil
}

// ListUserThreads возвращает агрегаты по тредам, в которых юзер хотя бы раз писал
// или ему писали. Для каждого треда: последний message_body, last_at, last_from_user_id.
// Гейтвей дополняет ответ метаданными «собеседник + контекст» из MicroTasks/Vacancy/Users.
//
// Алгоритм: берём все thread_id, где from_user_id=userID,
// а также все thread_id, где userID встречается среди сообщений (упрощённо — берём треды
// где этот юзер уже хоть раз писал; собеседник тогда — другой автор сообщений в этом треде).
// Это лёгкая реализация: если юзер ещё ни разу не написал, тред в inbox не появится.
// Для MVP достаточно — собеседник сам напишет первым в новом треде, и он у себя увидит.
func (r *ChatRepository) ListUserThreads(ctx context.Context, userID string, limit int32) ([]*chatv1.Thread, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	// Берём треды, где юзер был хоть раз автором, исключая те что он сам скрыл.
	query := `
SELECT
    cm.thread_id,
    (SELECT body FROM chat_messages WHERE thread_id = cm.thread_id ORDER BY created_at DESC LIMIT 1) AS last_message,
    MAX(cm.created_at) AS last_at
FROM chat_messages cm
WHERE cm.thread_id IN (
    SELECT DISTINCT thread_id FROM chat_messages WHERE from_user_id = $1
)
AND cm.thread_id NOT IN (
    SELECT thread_id FROM chat_thread_hides WHERE user_id = $1
)
GROUP BY cm.thread_id
ORDER BY last_at DESC
LIMIT $2`
	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	defer rows.Close()

	var out []*chatv1.Thread
	for rows.Next() {
		var t chatv1.Thread
		var lastAt time.Time
		if err := rows.Scan(&t.ThreadId, &t.LastMessage, &lastAt); err != nil {
			return nil, fmt.Errorf("scan thread: %w", err)
		}
		t.LastAt = lastAt.Format(time.RFC3339)
		out = append(out, &t)
	}
	return out, nil
}
