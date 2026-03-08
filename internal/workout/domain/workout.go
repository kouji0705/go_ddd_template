package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── Value Object: WorkoutID ──────────────────────────────────────────────────

type WorkoutID struct {
	value uuid.UUID
}

func NewWorkoutID() WorkoutID {
	return WorkoutID{value: uuid.New()}
}

func WorkoutIDFromString(s string) (WorkoutID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return WorkoutID{}, ErrInvalidWorkoutID
	}
	return WorkoutID{value: id}, nil
}

func (w WorkoutID) String() string  { return w.value.String() }
func (w WorkoutID) UUID() uuid.UUID { return w.value }

// ─── Value Object: WorkoutName ────────────────────────────────────────────────

type WorkoutName struct {
	value string
}

func NewWorkoutName(s string) (WorkoutName, error) {
	if s == "" {
		return WorkoutName{}, ErrEmptyName
	}
	return WorkoutName{value: s}, nil
}

func (n WorkoutName) String() string { return n.value }

// ─── Value Object: Calories ───────────────────────────────────────────────────

type Calories struct {
	value int
}

func NewCalories(v int) (Calories, error) {
	if v <= 0 {
		return Calories{}, ErrNonPositiveCalorie
	}
	return Calories{value: v}, nil
}

func (c Calories) Int() int { return c.value }

// ─── Value Object: Duration ───────────────────────────────────────────────────

type Duration struct {
	value int // minutes
}

func NewDuration(v int) (Duration, error) {
	if v <= 0 {
		return Duration{}, ErrNonPositiveDuration
	}
	return Duration{value: v}, nil
}

func (d Duration) Minutes() int { return d.value }

// ─── Aggregate Root: Workout ──────────────────────────────────────────────────

// Workout はワークアウトのアグリゲートルートです。
// フィールドは非公開にして、値の不変条件をドメイン層で保証します。
type Workout struct {
	id        WorkoutID
	name      WorkoutName
	calories  Calories
	duration  Duration
	createdAt time.Time
}

// NewWorkout はワークアウトを新規作成するファクトリ関数です。
func NewWorkout(name string, calories, duration int) (*Workout, error) {
	n, err := NewWorkoutName(name)
	if err != nil {
		return nil, err
	}
	c, err := NewCalories(calories)
	if err != nil {
		return nil, err
	}
	d, err := NewDuration(duration)
	if err != nil {
		return nil, err
	}
	return &Workout{
		id:        NewWorkoutID(),
		name:      n,
		calories:  c,
		duration:  d,
		createdAt: time.Now().UTC(),
	}, nil
}

// RestoreWorkout はDBなど永続化層からワークアウトを復元するためのファクトリ関数です。
func RestoreWorkout(id WorkoutID, name WorkoutName, calories Calories, duration Duration, createdAt time.Time) *Workout {
	return &Workout{
		id:        id,
		name:      name,
		calories:  calories,
		duration:  duration,
		createdAt: createdAt,
	}
}

func (w *Workout) ID() WorkoutID        { return w.id }
func (w *Workout) Name() WorkoutName    { return w.name }
func (w *Workout) Calories() Calories   { return w.calories }
func (w *Workout) Duration() Duration   { return w.duration }
func (w *Workout) CreatedAt() time.Time { return w.createdAt }
