package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/rinnothing/avito-pr/api/gen"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/logger"
)

func (s *serverImplementation) GetUsersGetReview(ctx echo.Context, params gen.GetUsersGetReviewParams) error {
	panic("not implemented")
}

func (s *serverImplementation) PostUsersSetIsActive(ectx echo.Context) error {
	ctx := logger.NewContext(ectx.Request().Context(), s.l.With(zap.String("method", "PostUsersSetIsActive")))

	var body gen.PostUsersSetIsActiveJSONRequestBody
	if err := ectx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("Invalid Request data in PostUsersSetIsActive: %s", err.Error()))
	}

	isActive := body.IsActive
	userId := model.UserId(body.UserId)
	user, err := s.user.SetIsActive(ctx, userId, isActive)
	if err != nil {

	}

	panic("not implemented")
}
