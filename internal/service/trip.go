package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"avito-go/internal/repository"
)

type TripService struct {
	transactor repository.Transactor
	repo       *repository.TripRepository
}

func NewTripService(transactor repository.Transactor, repo *repository.TripRepository) *TripService {
	return &TripService{
		transactor: transactor,
		repo:       repo,
	}
}

type CreateTripParams struct {
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
}

func (s *TripService) CreateTrip(ctx context.Context, params CreateTripParams) (*repository.Trip, error) {
	tripID := uuid.New()
	now := time.Now().UTC()

	trip := repository.Trip{
		ID:             tripID,
		UserID:         params.UserID,
		DriverID:       params.DriverID,
		StartLatitude:  params.StartLatitude,
		StartLongitude: params.StartLongitude,
		EndLatitude:    params.EndLatitude,
		EndLongitude:   params.EndLongitude,
		Price:          params.Price,
		Status:         "active",
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateTrip(txCtx, trip); err != nil {
			return fmt.Errorf("create trip: %w", err)
		}

		if err := s.repo.AddStatusHistory(txCtx, trip.ID, nil, "active", "trip created"); err != nil {
			return fmt.Errorf("add status history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &trip, nil
}

func (s *TripService) GetTrip(ctx context.Context, tripID uuid.UUID) (*repository.Trip, error) {
	return s.repo.GetTripByID(ctx, tripID)
}

func (s *TripService) FinishTrip(ctx context.Context, tripID uuid.UUID) (*repository.Trip, error) {
	now := time.Now().UTC()
	fromStatus := "active"

	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.FinishTrip(txCtx, tripID, now); err != nil {
			return fmt.Errorf("finish trip: %w", err)
		}

		if err := s.repo.AddStatusHistory(txCtx, tripID, &fromStatus, "completed", "trip finished"); err != nil {
			return fmt.Errorf("add status history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.GetTripByID(ctx, tripID)
}
