// Package infrastructure はリポジトリの具象実装を提供します。
// Bun ORM を使い、永続化モデル（workoutModel）とドメインモデルを完全に分離します。
package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/uptrace/bun"
)

// ─── 永続化モデル（ドメインモデルと完全分離）──────────────────────────────────

type workoutModel struct {
	bun.BaseModel `bun:"table:workouts,alias:w"`

	ID        uuid.UUID `bun:"type:uuid,pk,default:gen_random_uuid()"`
	Name      string    `bun:",notnull"`
	Calories  int       `bun:",notnull"`
	Duration  int       `bun:",notnull"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// ─── ドメイン ↔ 永続化モデル 変換 ─────────────────────────────────────────────

func toModel(w *domain.Workout) *workoutModel {
	return &workoutModel{
		ID:        w.ID().UUID(),
		Name:      w.Name().String(),
		Calories:  w.Calories().Int(),
		Duration:  w.Duration().Minutes(),
		CreatedAt: w.CreatedAt(),
	}
}

func toDomain(m *workoutModel) (*domain.Workout, error) {
	wid, err := domain.WorkoutIDFromString(m.ID.String())
	if err != nil {
		return nil, fmt.Errorf("restore workout id: %w", err)
	}
	name, err := domain.NewWorkoutName(m.Name)
	if err != nil {
		return nil, fmt.Errorf("restore workout name: %w", err)
	}
	cal, err := domain.NewCalories(m.Calories)
	if err != nil {
		return nil, fmt.Errorf("restore workout calories: %w", err)
	}
	dur, err := domain.NewDuration(m.Duration)
	if err != nil {
		return nil, fmt.Errorf("restore workout duration: %w", err)
	}
	return domain.RestoreWorkout(wid, name, cal, dur, m.CreatedAt), nil
}

// ─── WorkoutRepositoryImpl ────────────────────────────────────────────────────

type WorkoutRepositoryImpl struct {
	db *bun.DB
}

func NewWorkoutRepository(db *bun.DB) *WorkoutRepositoryImpl {
	return &WorkoutRepositoryImpl{db: db}
}

// AutoMigrate はテーブルが存在しない場合に作成します。
func (r *WorkoutRepositoryImpl) AutoMigrate(ctx context.Context) error {
	_, err := r.db.NewCreateTable().
		Model((*workoutModel)(nil)).
		IfNotExists().
		Exec(ctx)
	return err
}

func (r *WorkoutRepositoryImpl) Save(ctx context.Context, workout *domain.Workout) error {
	m := toModel(workout)
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (r *WorkoutRepositoryImpl) FindByID(ctx context.Context, id domain.WorkoutID) (*domain.Workout, error) {
	m := new(workoutModel)
	err := r.db.NewSelect().Model(m).Where("id = ?", id.UUID()).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrWorkoutNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(m)
}

func (r *WorkoutRepositoryImpl) FindAll(ctx context.Context) ([]*domain.Workout, error) {
	var models []*workoutModel
	if err := r.db.NewSelect().Model(&models).Order("created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}
	workouts := make([]*domain.Workout, 0, len(models))
	for _, m := range models {
		w, err := toDomain(m)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, nil
}
