package models

type ChatMessage struct {
	ID         string `json:"id"`
	ThreadID   string `json:"thread_id"`
	FromUserID string `json:"from_user_id"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
}

type ChatMessageList struct {
	Messages   []*ChatMessage      `json:"messages"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type ChatSendRequest struct {
	Body string `json:"body"`
}
