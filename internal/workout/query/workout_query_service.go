// Package query は読み取り系ユースケース（Query）を定義します。
// 状態を変更しない参照操作はすべてここに集約します。
package query

import (
	"context"

	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/kouji/go_ddd_template/internal/workout/repository"
)

// ─── Query DTO（出力） ───────────────────────────────────────────────────────

// WorkoutDTO はクエリ結果として返すデータ転送オブジェクトです。
type WorkoutDTO struct {
	ID        string
	Name      string
	Calories  int
	Duration  int
	CreatedAt string
}

func toDTO(w *domain.Workout) WorkoutDTO {
	return WorkoutDTO{
		ID:        w.ID().String(),
		Name:      w.Name().String(),
		Calories:  w.Calories().Int(),
		Duration:  w.Duration().Minutes(),
		CreatedAt: w.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}
}

// ─── WorkoutQueryService インターフェース ────────────────────────────────────

type WorkoutQueryService interface {
	GetAll(ctx context.Context) ([]WorkoutDTO, error)
	GetByID(ctx context.Context, id string) (WorkoutDTO, error)
}

// ─── 実装 ────────────────────────────────────────────────────────────────────

type workoutQueryService struct {
	repo repository.WorkoutRepository
}

// NewWorkoutQueryService は WorkoutQueryService の実装を返します。
func NewWorkoutQueryService(repo repository.WorkoutRepository) WorkoutQueryService {
	return &workoutQueryService{repo: repo}
}

// GetAll は全ワークアウトを返します。
func (s *workoutQueryService) GetAll(ctx context.Context) ([]WorkoutDTO, error) {
	workouts, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]WorkoutDTO, 0, len(workouts))
	for _, w := range workouts {
		dtos = append(dtos, toDTO(w))
	}
	return dtos, nil
}

// GetByID は指定した ID のワークアウトを返します。
func (s *workoutQueryService) GetByID(ctx context.Context, id string) (WorkoutDTO, error) {
	wid, err := domain.WorkoutIDFromString(id)
	if err != nil {
		return WorkoutDTO{}, err
	}
	workout, err := s.repo.FindByID(ctx, wid)
	if err != nil {
		return WorkoutDTO{}, err
	}
	return toDTO(workout), nil
}
