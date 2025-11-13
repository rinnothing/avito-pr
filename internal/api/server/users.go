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

type GetUsersReviewResponse struct {
	PullRequests []gen.PullRequestShort `json:"pull_requests"`
	UserId       string                 `json:"user_id"`
}

func (s *serverImplementation) GetUsersGetReview(c echo.Context, params gen.GetUsersGetReviewParams) error {
	if !isAdminOrUser(c) {
		return unauthorized(c, "Only admins and users are allowed to execute this method")
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "GetUsersGetReview"), zap.String("user_id", params.UserId)))

	userId := model.UserId(params.UserId)
	newId, prs, err := s.pr.ListPR(ctx, userId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			prs = make([]model.PullRequest, 0)
		} else {
			return reportInternalError(ctx, err)
		}
	}

	pullRequests := make([]gen.PullRequestShort, 0, len(prs))
	for i := range prs {
		pullRequests = append(pullRequests, gen.PullRequestShort{
			AuthorId:        string(prs[i].AuthorId),
			PullRequestId:   string(prs[i].Id),
			PullRequestName: string(prs[i].Name),
			Status:          mapShortStatus(prs[i].Status),
		})
	}

	resp := &GetUsersReviewResponse{
		PullRequests: pullRequests,
		UserId:       string(newId),
	}
	return c.JSON(http.StatusOK, resp)
}

type SetActiveResponse struct {
	User *gen.User `json:"user,omitempty"`
}

func (s *serverImplementation) PostUsersSetIsActive(c echo.Context) error {
	if !isAdmin(c) {
		return unauthorized(c, "Only admins are allowed to execute this method")
	}

	var body gen.PostUsersSetIsActiveJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return badRequest(fmt.Sprintf("Invalid Request data in PostUsersSetIsActive: %s", err.Error()))
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "PostUsersSetIsActive"), zap.String("user_id", body.UserId)))

	isActive := body.IsActive
	userId := model.UserId(body.UserId)
	user, err := s.user.SetIsActive(ctx, userId, isActive)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound(c, err.Error())
		}
		return reportInternalError(ctx, err)
	}

	resp := &SetActiveResponse{
		User: &gen.User{
			IsActive: user.IsActive,
			TeamName: string(user.TeamName),
			UserId:   string(user.Id),
			Username: user.Username,
		},
	}
	return c.JSON(http.StatusOK, resp)
}
