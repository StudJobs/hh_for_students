package handlers

import (
	"context"
	"log"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	"github.com/studjobs/hh_for_students/search/internal/indexer"
	"github.com/studjobs/hh_for_students/search/internal/reindexer"
	"github.com/studjobs/hh_for_students/search/internal/searcher"
)

type Handler struct {
	searchv1.UnimplementedSearchServiceServer

	searcher  *searcher.Searcher
	indexer   *indexer.Indexer
	reindexer *reindexer.Reindexer
}

func New(s *searcher.Searcher, i *indexer.Indexer, r *reindexer.Reindexer) *Handler {
	return &Handler{searcher: s, indexer: i, reindexer: r}
}

func (h *Handler) SearchProfiles(ctx context.Context, req *searchv1.SearchProfilesRequest) (*usersv1.ProfileList, error) {
	log.Printf("Handler: SearchProfiles q=%q skills=%v cat=%q", req.GetQuery(), req.GetSkillSlugs(), req.GetProfessionCategory())
	return h.searcher.SearchProfiles(ctx, req)
}

func (h *Handler) SearchVacancies(ctx context.Context, req *searchv1.SearchVacanciesRequest) (*vacancyv1.VacancyList, error) {
	log.Printf("Handler: SearchVacancies q=%q skills=%v salary>=%d exp<=%d company=%q",
		req.GetQuery(), req.GetSkillSlugs(), req.GetSalaryMin(), req.GetExperienceMax(), req.GetCompanyId())
	return h.searcher.SearchVacancies(ctx, req)
}

func (h *Handler) IndexProfile(ctx context.Context, req *searchv1.IndexProfileRequest) (*commonv1.Empty, error) {
	if err := h.indexer.IndexProfile(ctx, req.GetProfile()); err != nil {
		log.Printf("Handler: IndexProfile id=%s error: %v", req.GetProfile().GetId(), err)
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

func (h *Handler) IndexVacancy(ctx context.Context, req *searchv1.IndexVacancyRequest) (*commonv1.Empty, error) {
	if err := h.indexer.IndexVacancy(ctx, req.GetVacancy()); err != nil {
		log.Printf("Handler: IndexVacancy id=%s error: %v", req.GetVacancy().GetId(), err)
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

func (h *Handler) DeleteProfile(ctx context.Context, req *searchv1.DeleteDocumentRequest) (*commonv1.Empty, error) {
	if err := h.indexer.DeleteProfile(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

func (h *Handler) DeleteVacancy(ctx context.Context, req *searchv1.DeleteDocumentRequest) (*commonv1.Empty, error) {
	if err := h.indexer.DeleteVacancy(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

func (h *Handler) Reindex(ctx context.Context, req *searchv1.ReindexRequest) (*searchv1.ReindexResponse, error) {
	log.Printf("Handler: Reindex recreate=%v", req.GetRecreateIndices())
	profiles, vacancies, err := h.reindexer.Run(ctx, req.GetRecreateIndices())
	if err != nil {
		return nil, err
	}
	return &searchv1.ReindexResponse{
		IndexedProfiles:  profiles,
		IndexedVacancies: vacancies,
	}, nil
}
