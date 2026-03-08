// Package command は書き込み系ユースケース（Command）を定義します。
// 状態を変更する操作はすべてここに集約します。
package command

import (
	"context"

	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/kouji/go_ddd_template/internal/workout/repository"
)

// ─── Command（入力） ─────────────────────────────────────────────────────────

// CreateWorkoutCommand はワークアウト作成に必要な入力データです。
type CreateWorkoutCommand struct {
	Name     string
	Calories int
	Duration int
}

// ─── Result（出力） ──────────────────────────────────────────────────────────

// WorkoutResult はコマンド実行後に返す結果型です。
type WorkoutResult struct {
	ID        string
	Name      string
	Calories  int
	Duration  int
	CreatedAt string
}

func toResult(w *domain.Workout) WorkoutResult {
	return WorkoutResult{
		ID:        w.ID().String(),
		Name:      w.Name().String(),
		Calories:  w.Calories().Int(),
		Duration:  w.Duration().Minutes(),
		CreatedAt: w.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}
}

// ─── WorkoutCommandService インターフェース ──────────────────────────────────

type WorkoutCommandService interface {
	CreateWorkout(ctx context.Context, cmd CreateWorkoutCommand) (WorkoutResult, error)
}

// ─── 実装 ────────────────────────────────────────────────────────────────────

type workoutCommandService struct {
	repo repository.WorkoutRepository
}

// NewWorkoutCommandService は WorkoutCommandService の実装を返します。
func NewWorkoutCommandService(repo repository.WorkoutRepository) WorkoutCommandService {
	return &workoutCommandService{repo: repo}
}

// CreateWorkout はワークアウトを新規作成して永続化します。
func (s *workoutCommandService) CreateWorkout(ctx context.Context, cmd CreateWorkoutCommand) (WorkoutResult, error) {
	workout, err := domain.NewWorkout(cmd.Name, cmd.Calories, cmd.Duration)
	if err != nil {
		// ドメインバリデーションエラーはそのまま上位に伝播
		return WorkoutResult{}, err
	}
	if err := s.repo.Save(ctx, workout); err != nil {
		return WorkoutResult{}, err
	}
	return toResult(workout), nil
}
