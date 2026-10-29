package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrDriverBusy    = errors.New("driver already has active trip")
	ErrInvalidStatus = errors.New("invalid status transition")
)

type Trip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TripRepository struct {
	mgr *Manager
}

func NewTripRepository(mgr *Manager) *TripRepository {
	return &TripRepository{mgr: mgr}
}

func (r *TripRepository) CreateTrip(ctx context.Context, trip Trip) error {
	db := r.mgr.GetDB(ctx)

	query, args, err := psql.Insert("trips").
		Columns(
			"id", "user_id", "driver_id",
			"start_latitude", "start_longitude", "end_latitude", "end_longitude",
			"price", "status", "started_at", "created_at", "updated_at",
		).
		Values(
			trip.ID, trip.UserID, trip.DriverID,
			trip.StartLatitude, trip.StartLongitude, trip.EndLatitude, trip.EndLongitude,
			trip.Price, trip.Status, trip.StartedAt, trip.CreatedAt, trip.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip sql: %w", err)
	}

	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute insert trip: %w", err)
	}

	return nil
}

func (r *TripRepository) AddStatusHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus string, reason string) error {
	db := r.mgr.GetDB(ctx)

	query, args, err := psql.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(tripID, fromStatus, toStatus, reason, time.Now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert history sql: %w", err)
	}

	_, err = db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute insert history: %w", err)
	}

	return nil
}

func (r *TripRepository) GetTripByID(ctx context.Context, id uuid.UUID) (*Trip, error) {
	db := r.mgr.GetDB(ctx)

	query, args, err := psql.Select(
		"id", "user_id", "driver_id",
		"start_latitude", "start_longitude", "end_latitude", "end_longitude",
		"price", "status", "started_at", "finished_at", "created_at", "updated_at",
	).
		From("trips").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select trip sql: %w", err)
	}

	row := db.QueryRow(ctx, query, args...)
	var t Trip
	err = row.Scan(
		&t.ID, &t.UserID, &t.DriverID,
		&t.StartLatitude, &t.StartLongitude, &t.EndLatitude, &t.EndLongitude,
		&t.Price, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("scan trip: %w", err)
	}

	return &t, nil
}

func (r *TripRepository) FinishTrip(ctx context.Context, id uuid.UUID, finishedAt time.Time) error {
	db := r.mgr.GetDB(ctx)

	query, args, err := psql.Update("trips").
		Set("status", "completed").
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(sq.Eq{"id": id, "status": "active"}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build finish trip sql: %w", err)
	}

	tag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("execute finish trip: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrTripNotFound
	}

	return nil
}
