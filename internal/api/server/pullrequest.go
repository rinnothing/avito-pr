package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

type PostPullRequestCreateResponse struct {
	Pr *gen.PullRequest `json:"pr,omitempty"`
}

func (s *serverImplementation) PostPullRequestCreate(c echo.Context) error {
	if !isAdmin(c) {
		return unauthorized(c, "Only admins are allowed to execute this method")
	}

	var body gen.PostPullRequestCreateJSONBody
	if err := c.Bind(&body); err != nil {
		return badRequest(fmt.Sprintf("Invalid Request data in PostPullRequestCreate: %s", err.Error()))
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "PostPullRequestCreate"), zap.String("pull_request_id", body.PullRequestId)))

	pr := &model.PullRequest{
		AuthorId: model.UserId(body.AuthorId),
		Id:       model.PRId(body.PullRequestId),
		Name:     body.PullRequestName,
	}
	newPr, err := s.pr.CreatePR(ctx, pr)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound(c, err.Error())
		} else if errors.Is(err, model.ErrAlreadyExists) {
			return conflict(c, err.Error(), gen.PREXISTS)
		}
		return reportInternalError(ctx, err)
	}

	resp := &PostPullRequestCreateResponse{
		Pr: prToResp(newPr),
	}
	return c.JSON(http.StatusCreated, resp)
}

type PostPulllRequestMergeResponse struct {
	Pr *gen.PullRequest `json:"pr,omitempty"`
}

func (s *serverImplementation) PostPullRequestMerge(c echo.Context) error {
	if !isAdmin(c) {
		return unauthorized(c, "Only admins are allowed to execute this method")
	}

	var body gen.PostPullRequestMergeJSONBody
	if err := c.Bind(&body); err != nil {
		return badRequest(fmt.Sprintf("Invalid Request data in PostPullRequestMerge: %s", err.Error()))
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "PostPullRequestMerge"), zap.String("pull_request_id", body.PullRequestId)))

	id := model.PRId(body.PullRequestId)
	pr, err := s.pr.MergePR(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound(c, err.Error())
		}
		return reportInternalError(ctx, err)
	}

	resp := &PostPulllRequestMergeResponse{
		Pr: prToResp(pr),
	}
	return c.JSON(http.StatusOK, resp)
}

type PostPullRequestReassignResponse struct {
	Pr         gen.PullRequest `json:"pr"`
	ReplacedBy string          `json:"replaced_by"`
}

func (s *serverImplementation) PostPullRequestReassign(c echo.Context) error {
	if !isAdmin(c) {
		return unauthorized(c, "Only admins are allowed to execute this method")
	}

	var body gen.PostPullRequestReassignJSONBody
	if err := c.Bind(&body); err != nil {
		return badRequest(fmt.Sprintf("Invalid Request data in PostPullRequestReassign: %s", err.Error()))
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "PostPullRequestReassign"), zap.String("pull_request_id", body.PullRequestId), zap.String("old_user_id", body.OldUserId)))

	prId := model.PRId(body.PullRequestId)
	oldUserId := model.UserId(body.OldUserId)
	pr, newUserId, err := s.pr.Reassign(ctx, prId, oldUserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound(c, err.Error())
		} else if errors.Is(err, model.ErrNoCandidates) {
			return conflict(c, err.Error(), gen.NOCANDIDATE)
		} else if errors.Is(err, model.ErrNotReviewer) {
			return conflict(c, err.Error(), gen.NOTASSIGNED)
		} else if errors.Is(err, model.ErrAlreadyMerged) {
			return conflict(c, err.Error(), gen.PRMERGED)
		}
		return reportInternalError(ctx, err)
	}

	resp := &PostPullRequestReassignResponse{
		ReplacedBy: string(newUserId),
		Pr:         *prToResp(pr),
	}
	return c.JSON(http.StatusOK, resp)
}

func prToResp(pr *model.PullRequest) *gen.PullRequest {
	reviewers := make([]string, 0, len(pr.Reviewers))
	for i := range len(pr.Reviewers) {
		reviewers = append(reviewers, string(pr.Reviewers[i]))
	}

	return &gen.PullRequest{
		AssignedReviewers: reviewers,
		AuthorId:          string(pr.AuthorId),
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
		PullRequestId:     string(pr.Id),
		PullRequestName:   pr.Name,
		Status:            mapStatus(pr.Status),
	}
}
