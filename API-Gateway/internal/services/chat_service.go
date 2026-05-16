package services

import (
	"context"

	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type chatService struct {
	client chatv1.ChatServiceClient
}

func NewChatService(client chatv1.ChatServiceClient) ChatService {
	return &chatService{client: client}
}

func (s *chatService) SendMessage(ctx context.Context, threadID, fromUser, body string) (*models.ChatMessage, error) {
	resp, err := s.client.SendMessage(ctx, &chatv1.SendMessageRequest{ThreadId: threadID, FromUserId: fromUser, Body: body})
	if err != nil {
		return nil, err
	}
	return chatMessageFromProto(resp), nil
}

func (s *chatService) ListMessages(ctx context.Context, threadID string, page, limit int32) (*models.ChatMessageList, error) {
	resp, err := s.client.ListMessages(ctx, &chatv1.ListMessagesRequest{
		ThreadId:   threadID,
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	out := &models.ChatMessageList{}
	for _, m := range resp.GetMessages() {
		out.Messages = append(out.Messages, chatMessageFromProto(m))
	}
	if p := resp.GetPagination(); p != nil {
		out.Pagination = &models.PaginationResponse{Total: p.GetTotal(), Pages: p.GetPages(), CurrentPage: p.GetCurrentPage()}
	}
	return out, nil
}

func (s *chatService) ListUserThreads(ctx context.Context, userID string, limit int32) ([]*models.ChatThread, error) {
	resp, err := s.client.ListUserThreads(ctx, &chatv1.ListUserThreadsRequest{
		UserId:     userID,
		Pagination: &commonv1.Pagination{Page: 1, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.ChatThread, 0, len(resp.GetThreads()))
	for _, t := range resp.GetThreads() {
		out = append(out, &models.ChatThread{
			ThreadID:    t.GetThreadId(),
			LastMessage: t.GetLastMessage(),
			LastAt:      t.GetLastAt(),
			UnreadCount: t.GetUnreadCount(),
		})
	}
	return out, nil
}

func chatMessageFromProto(p *chatv1.Message) *models.ChatMessage {
	if p == nil {
		return nil
	}
	return &models.ChatMessage{
		ID:         p.GetId(),
		ThreadID:   p.GetThreadId(),
		FromUserID: p.GetFromUserId(),
		Body:       p.GetBody(),
		CreatedAt:  p.GetCreatedAt(),
	}
}
