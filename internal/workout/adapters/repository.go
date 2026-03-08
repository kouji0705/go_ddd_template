// Package adapters はドメインのリポジトリインターフェースの具象実装を提供します。
// ドメインモデルと永続化モデル（bunModel）を分離し、インフラ詳細がドメインに漏れないようにします。
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kouji/go_ddd_template/internal/workout/domain"
	"github.com/uptrace/bun"
)

// ─── 永続化モデル（ドメインモデルとは別に定義）────────────────────────────────

type workoutModel struct {
	bun.BaseModel `bun:"table:workouts,alias:w"`

	ID        uuid.UUID `bun:"type:uuid,pk,default:gen_random_uuid()"`
	Name      string    `bun:",notnull"`
	Calories  int       `bun:",notnull"`
	Duration  int       `bun:",notnull"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// ─── マッピング関数 ───────────────────────────────────────────────────────────

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
	id := domain.WorkoutID{}
	wid, err := domain.WorkoutIDFromString(m.ID.String())
	if err != nil {
		return nil, fmt.Errorf("restore workout id: %w", err)
	}
	id = wid

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
	return domain.RestoreWorkout(id, name, cal, dur, m.CreatedAt), nil
}

// ─── WorkoutRepository ────────────────────────────────────────────────────────

type WorkoutRepository struct {
	db *bun.DB
}

func NewWorkoutRepository(db *bun.DB) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

// AutoMigrate はテーブルが存在しない場合に作成します。
func (r *WorkoutRepository) AutoMigrate(ctx context.Context) error {
	_, err := r.db.NewCreateTable().
		Model((*workoutModel)(nil)).
		IfNotExists().
		Exec(ctx)
	return err
}

func (r *WorkoutRepository) Save(ctx context.Context, workout *domain.Workout) error {
	m := toModel(workout)
	_, err := r.db.NewInsert().Model(m).Exec(ctx)
	return err
}

func (r *WorkoutRepository) FindByID(ctx context.Context, id domain.WorkoutID) (*domain.Workout, error) {
	m := new(workoutModel)
	err := r.db.NewSelect().Model(m).Where("id = ?", id.UUID()).Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, domain.ErrWorkoutNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(m)
}

func (r *WorkoutRepository) FindAll(ctx context.Context) ([]*domain.Workout, error) {
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
