// Package app はアプリケーション層です。
// ドメインオブジェクトを操作するユースケース（Command / Query）を定義します。
package app

import (
	"context"

	"github.com/kouji/go_ddd_template/internal/workout/domain"
)

// ─── Command / Query の入出力型 ───────────────────────────────────────────────

// CreateWorkoutCommand はワークアウト作成のコマンドです。
type CreateWorkoutCommand struct {
	Name     string
	Calories int
	Duration int
}

// WorkoutDTO はアプリケーション層の出力型（ドメインオブジェクトを外部に漏らさない）です。
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

// ─── UseCase インターフェース（依存逆転・テスト容易性のため） ─────────────────

type WorkoutCreator interface {
	CreateWorkout(ctx context.Context, cmd CreateWorkoutCommand) (WorkoutDTO, error)
}

type WorkoutReader interface {
	GetWorkouts(ctx context.Context) ([]WorkoutDTO, error)
	GetWorkoutByID(ctx context.Context, id string) (WorkoutDTO, error)
}

// WorkoutService はすべてのユースケースを束ねたインターフェースです。
type WorkoutService interface {
	WorkoutCreator
	WorkoutReader
}

// ─── Service 実装 ─────────────────────────────────────────────────────────────

type service struct {
	repo domain.Repository
}

// NewService は WorkoutService の実装を返します。
func NewService(repo domain.Repository) WorkoutService {
	return &service{repo: repo}
}

func (s *service) CreateWorkout(ctx context.Context, cmd CreateWorkoutCommand) (WorkoutDTO, error) {
	workout, err := domain.NewWorkout(cmd.Name, cmd.Calories, cmd.Duration)
	if err != nil {
		return WorkoutDTO{}, err
	}
	if err := s.repo.Save(ctx, workout); err != nil {
		return WorkoutDTO{}, err
	}
	return toDTO(workout), nil
}

func (s *service) GetWorkouts(ctx context.Context) ([]WorkoutDTO, error) {
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

func (s *service) GetWorkoutByID(ctx context.Context, id string) (WorkoutDTO, error) {
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
