package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"avito-go/internal/api"
	"avito-go/internal/repository"
	"avito-go/internal/service"
)

type Handler struct {
	tripService *service.TripService
}

func NewHandler(tripService *service.TripService) *Handler {
	return &Handler{tripService: tripService}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var req api.CreateTripJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	trip, err := h.tripService.CreateTrip(r.Context(), service.CreateTripParams{
		UserID:         req.UserId,
		DriverID:       req.DriverId,
		StartLatitude:  req.StartPoint.Latitude,
		StartLongitude: req.StartPoint.Longitude,
		EndLatitude:    req.EndPoint.Latitude,
		EndLongitude:   req.EndPoint.Longitude,
		Price:          req.Price,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := mapTripToAPI(trip)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.tripService.GetTrip(r.Context(), uuid.UUID(tripId))
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			http.Error(w, "trip not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := mapTripToAPI(trip)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.tripService.FinishTrip(r.Context(), uuid.UUID(tripId))
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			http.Error(w, "trip not found or not active", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := mapTripToAPI(trip)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mapTripToAPI(t *repository.Trip) api.Trip {
	res := api.Trip{
		Id:       api.TripId(t.ID),
		UserId:   t.UserID,
		DriverId: t.DriverID,
		StartPoint: api.Coordinates{
			Latitude:  t.StartLatitude,
			Longitude: t.StartLongitude,
		},
		EndPoint: api.Coordinates{
			Latitude:  t.EndLatitude,
			Longitude: t.EndLongitude,
		},
		Price:     t.Price,
		Status:    api.TripStatus(t.Status),
		StartedAt: t.StartedAt,
	}
	if t.FinishedAt != nil {
		res.FinishedAt = t.FinishedAt
	}
	return res
}
