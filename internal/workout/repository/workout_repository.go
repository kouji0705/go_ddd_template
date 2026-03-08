// Package repository はワークアウトの永続化インターフェースを定義します。
// domain パッケージに依存しますが、実装（infrastructure）には依存しません。
package repository

import (
	"context"

	"github.com/kouji/go_ddd_template/internal/workout/domain"
)

// WorkoutRepository はワークアウトの永続化操作を定義するインターフェースです。
// infrastructure パッケージがこのインターフェースを実装します。
type WorkoutRepository interface {
	Save(ctx context.Context, workout *domain.Workout) error
	FindByID(ctx context.Context, id domain.WorkoutID) (*domain.Workout, error)
	FindAll(ctx context.Context) ([]*domain.Workout, error)
}
