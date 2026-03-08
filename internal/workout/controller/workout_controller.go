// Package controller はHTTPリクエストを受け取り、
// CommandService / QueryService に処理を委譲するController層です。
package controller

import (
	"errors"
	"net/http"

	"github.com/kouji/go_ddd_template/internal/workout/command"
	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/kouji/go_ddd_template/internal/workout/query"
	"github.com/labstack/echo/v4"
)

// ─── HTTP レスポンス型 ────────────────────────────────────────────────────────

type workoutResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Calories  int    `json:"calories"`
	DurationMin int  `json:"duration_minutes"`
	CreatedAt string `json:"created_at"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func fromCommandResult(r command.WorkoutResult) workoutResponse {
	return workoutResponse{
		ID:          r.ID,
		Name:        r.Name,
		Calories:    r.Calories,
		DurationMin: r.Duration,
		CreatedAt:   r.CreatedAt,
	}
}

func fromQueryDTO(d query.WorkoutDTO) workoutResponse {
	return workoutResponse{
		ID:          d.ID,
		Name:        d.Name,
		Calories:    d.Calories,
		DurationMin: d.Duration,
		CreatedAt:   d.CreatedAt,
	}
}

// ─── ドメインエラー → HTTP ステータスのマッピング ─────────────────────────────

func httpError(err error) (int, errorResponse) {
	switch {
	case errors.Is(err, domain.ErrWorkoutNotFound):
		return http.StatusNotFound, errorResponse{Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidWorkoutID),
		errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrNonPositiveCalorie),
		errors.Is(err, domain.ErrNonPositiveDuration):
		return http.StatusUnprocessableEntity, errorResponse{Message: err.Error()}
	default:
		return http.StatusInternalServerError, errorResponse{Message: "internal server error"}
	}
}

// ─── WorkoutController ────────────────────────────────────────────────────────

// WorkoutController は CommandService と QueryService の両方を持ちます。
type WorkoutController struct {
	commandSvc command.WorkoutCommandService
	querySvc   query.WorkoutQueryService
}

func NewWorkoutController(
	commandSvc command.WorkoutCommandService,
	querySvc query.WorkoutQueryService,
) *WorkoutController {
	return &WorkoutController{
		commandSvc: commandSvc,
		querySvc:   querySvc,
	}
}

// RegisterRoutes はEchoのルーターにエンドポイントを登録します。
func (ctrl *WorkoutController) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/workouts")
	g.POST("", ctrl.CreateWorkout)
	g.GET("", ctrl.GetWorkouts)
	g.GET("/:id", ctrl.GetWorkoutByID)
}

// POST /workouts
func (ctrl *WorkoutController) CreateWorkout(c echo.Context) error {
	type request struct {
		Name     string `json:"name"`
		Calories int    `json:"calories"`
		Duration int    `json:"duration"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Message: "invalid request body"})
	}

	result, err := ctrl.commandSvc.CreateWorkout(c.Request().Context(), command.CreateWorkoutCommand{
		Name:     req.Name,
		Calories: req.Calories,
		Duration: req.Duration,
	})
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}
	return c.JSON(http.StatusCreated, fromCommandResult(result))
}

// GET /workouts
func (ctrl *WorkoutController) GetWorkouts(c echo.Context) error {
	dtos, err := ctrl.querySvc.GetAll(c.Request().Context())
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}
	resp := make([]workoutResponse, 0, len(dtos))
	for _, d := range dtos {
		resp = append(resp, fromQueryDTO(d))
	}
	return c.JSON(http.StatusOK, resp)
}

// GET /workouts/:id
func (ctrl *WorkoutController) GetWorkoutByID(c echo.Context) error {
	id := c.Param("id")
	dto, err := ctrl.querySvc.GetByID(c.Request().Context(), id)
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}
	return c.JSON(http.StatusOK, fromQueryDTO(dto))
}
