// Package ports はHTTPインターフェースアダプター層です。
// アプリケーション層のユースケースをHTTPハンドラーとして公開します。
package ports

import (
	"errors"
	"net/http"

	"github.com/kouji/go_ddd_template/internal/workout/app"
	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/labstack/echo/v4"
)

// ─── レスポンス型（外部APIスキーマをアプリ層DTOから分離）────────────────────

type workoutResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Calories  int    `json:"calories"`
	Duration  int    `json:"duration_minutes"`
	CreatedAt string `json:"created_at"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func dtoToResponse(dto app.WorkoutDTO) workoutResponse {
	return workoutResponse{
		ID:        dto.ID,
		Name:      dto.Name,
		Calories:  dto.Calories,
		Duration:  dto.Duration,
		CreatedAt: dto.CreatedAt,
	}
}

// ─── エラーマッピング ─────────────────────────────────────────────────────────

func httpError(err error) (int, errorResponse) {
	switch {
	case errors.Is(err, domain.ErrWorkoutNotFound):
		return http.StatusNotFound, errorResponse{Message: err.Error()}
	case errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrNonPositiveCalorie),
		errors.Is(err, domain.ErrNonPositiveDuration):
		return http.StatusUnprocessableEntity, errorResponse{Message: err.Error()}
	default:
		return http.StatusInternalServerError, errorResponse{Message: "internal server error"}
	}
}

// ─── HTTPHandler ──────────────────────────────────────────────────────────────

// HTTPHandler は WorkoutService に依存します（具象型ではなくインターフェース）。
type HTTPHandler struct {
	service app.WorkoutService
}

func NewHTTPHandler(service app.WorkoutService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// RegisterRoutes はEchoのルーターにエンドポイントを登録します。
func (h *HTTPHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/workouts")
	g.POST("", h.CreateWorkout)
	g.GET("", h.GetWorkouts)
	g.GET("/:id", h.GetWorkoutByID)
}

// POST /workouts
func (h *HTTPHandler) CreateWorkout(c echo.Context) error {
	type request struct {
		Name     string `json:"name"`
		Calories int    `json:"calories"`
		Duration int    `json:"duration"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Message: "invalid request body"})
	}

	dto, err := h.service.CreateWorkout(c.Request().Context(), app.CreateWorkoutCommand{
		Name:     req.Name,
		Calories: req.Calories,
		Duration: req.Duration,
	})
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}

	return c.JSON(http.StatusCreated, dtoToResponse(dto))
}

// GET /workouts
func (h *HTTPHandler) GetWorkouts(c echo.Context) error {
	dtos, err := h.service.GetWorkouts(c.Request().Context())
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}

	resp := make([]workoutResponse, 0, len(dtos))
	for _, dto := range dtos {
		resp = append(resp, dtoToResponse(dto))
	}
	return c.JSON(http.StatusOK, resp)
}

// GET /workouts/:id
func (h *HTTPHandler) GetWorkoutByID(c echo.Context) error {
	id := c.Param("id")
	dto, err := h.service.GetWorkoutByID(c.Request().Context(), id)
	if err != nil {
		code, body := httpError(err)
		return c.JSON(code, body)
	}
	return c.JSON(http.StatusOK, dtoToResponse(dto))
}
