package models

type ChatMessage struct {
	ID         string `json:"id"`
	ThreadID   string `json:"thread_id"`
	FromUserID string `json:"from_user_id"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	EditedAt   string `json:"edited_at,omitempty"`
}

type ChatMessageList struct {
	Messages   []*ChatMessage      `json:"messages"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type ChatSendRequest struct {
	Body string `json:"body"`
}

// ChatThread — расширенный тред со всеми метаданными для UI.
type ChatThread struct {
	ThreadID    string `json:"thread_id"`
	Kind        string `json:"kind,omitempty"`     // "application" | "task" | "quest"
	ResourceID  string `json:"resource_id,omitempty"`
	LastMessage string `json:"last_message,omitempty"`
	LastAt      string `json:"last_at,omitempty"`
	UnreadCount int32  `json:"unread_count,omitempty"`

	// Метаданные собеседника (заполняет Gateway).
	PeerID        string `json:"peer_id,omitempty"`
	PeerName      string `json:"peer_name,omitempty"`
	PeerRole      string `json:"peer_role,omitempty"`
	PeerCompany   string `json:"peer_company,omitempty"`
	PeerAvatarURL string `json:"peer_avatar_url,omitempty"`

	// Метаданные контекста (название вакансии / задачи / навыка квеста).
	ContextTitle string `json:"context_title,omitempty"`
}
