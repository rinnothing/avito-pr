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

type PostTeamAddResponse struct {
	Team *gen.Team `json:"team,omitempty"`
}

func (s *serverImplementation) PostTeamAdd(c echo.Context) error {
	var body gen.PostTeamAddJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return badRequest(fmt.Sprintf("Invalid Request data in PostTeamAdd: %s", err.Error()))
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "PostTeamAdd"), zap.String("teamname", body.TeamName)))

	teamIds := make([]model.UserId, 0, len(body.Members))
	members := make([]model.User, 0, len(body.Members))
	for i := range body.Members {
		members = append(members, model.User{
			Id:       model.UserId(body.Members[i].UserId),
			Username: body.Members[i].Username,
			TeamName: model.TeamName(body.TeamName),
			IsActive: body.Members[i].IsActive,
		})
	}
	team := &model.Team{
		Name:    model.TeamName(body.TeamName),
		Members: teamIds,
	}
	newTeam, newUsers, err := s.team.CreateTeam(ctx, team, members)
	if err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return alreadyExists(c, err.Error())
		}
		return reportInternalError(ctx, err)
	}

	resp := &PostTeamAddResponse{
		Team: teamToResp(newTeam, newUsers),
	}
	return c.JSON(http.StatusCreated, resp)
}

type GetTeamGetResponse *gen.Team

func (s *serverImplementation) GetTeamGet(c echo.Context, params gen.GetTeamGetParams) error {
	if !isAdminOrUser(c) {
		return unauthorized(c, "Only admins and users are allowed to execute this method")
	}

	ctx := logger.NewContext(c.Request().Context(), s.l.With(zap.String("method", "GetTeamGet"), zap.String("teamname", params.TeamName)))

	teamName := model.TeamName(params.TeamName)
	team, users, err := s.team.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound(c, err.Error())
		}
		return reportInternalError(ctx, err)
	}

	resp := GetTeamGetResponse(teamToResp(team, users))
	return c.JSON(http.StatusOK, resp)
}

func teamToResp(team *model.Team, users []model.User) *gen.Team {
	members := make([]gen.TeamMember, 0, len(users))
	for i := range users {
		members = append(members, gen.TeamMember{
			IsActive: users[i].IsActive,
			UserId:   string(users[i].Id),
			Username: users[i].Username,
		})
	}

	return &gen.Team{
		TeamName: string(team.Name),
		Members:  members,
	}
}
