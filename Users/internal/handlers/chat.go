package handlers

import (
	"context"
	"log"

	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/studjobs/hh_for_students/users/internal/repository"
)

type ChatHandler struct {
	chatv1.UnimplementedChatServiceServer
	repo *repository.Repository
}

func NewChatHandler(repo *repository.Repository) *ChatHandler {
	return &ChatHandler{repo: repo}
}

func (h *ChatHandler) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.Message, error) {
	if req.GetThreadId() == "" || req.GetFromUserId() == "" || req.GetBody() == "" {
		return nil, status.Error(codes.InvalidArgument, "thread_id, from_user_id, body required")
	}
	m, err := h.repo.Chat.Insert(ctx, req.GetThreadId(), req.GetFromUserId(), req.GetBody())
	if err != nil {
		log.Printf("ChatHandler: SendMessage failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to send message")
	}
	return m, nil
}

func (h *ChatHandler) ListMessages(ctx context.Context, req *chatv1.ListMessagesRequest) (*chatv1.MessageList, error) {
	if req.GetThreadId() == "" {
		return nil, status.Error(codes.InvalidArgument, "thread_id required")
	}
	var page, limit int32 = 1, 50
	if p := req.GetPagination(); p != nil {
		if p.GetPage() > 0 {
			page = p.GetPage()
		}
		if p.GetLimit() > 0 {
			limit = p.GetLimit()
		}
	}
	list, err := h.repo.Chat.ListByThread(ctx, req.GetThreadId(), page, limit)
	if err != nil {
		log.Printf("ChatHandler: ListMessages thread=%s failed: %v", req.GetThreadId(), err)
		return nil, status.Error(codes.Internal, "failed to list messages")
	}
	return list, nil
}

func (h *ChatHandler) ListUserThreads(ctx context.Context, req *chatv1.ListUserThreadsRequest) (*chatv1.ThreadList, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	limit := int32(100)
	if p := req.GetPagination(); p != nil && p.GetLimit() > 0 {
		limit = p.GetLimit()
	}
	threads, err := h.repo.Chat.ListUserThreads(ctx, req.GetUserId(), limit)
	if err != nil {
		log.Printf("ChatHandler: ListUserThreads user=%s failed: %v", req.GetUserId(), err)
		return nil, status.Error(codes.Internal, "failed to list threads")
	}
	return &chatv1.ThreadList{Threads: threads, Pagination: &commonv1.PaginationResponse{Total: int32(len(threads))}}, nil
}
