package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/sl"
	"github.com/NikKazzzzzz/Calendar-main/internal/models"
	"github.com/NikKazzzzzz/Calendar-main/internal/storage/mongodb"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventHandler struct {
	storage *mongodb.Storage
	Logger  *slog.Logger
}

func NewEventHandler(storage *mongodb.Storage, logger *slog.Logger) *EventHandler {
	return &EventHandler{storage: storage, Logger: logger}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.Logger.Error("failed to decode request body", sl.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validateEventsTime(event.StartTime, event.EndTime, h.Logger, w) {
		return
	}

	isTaken, err := h.storage.IsTimeSlotTaken(event.StartTime, event.EndTime)
	if err != nil {
		h.Logger.Error("failed to check time slot", sl.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if isTaken {
		h.Logger.Error("time slot is already taken",
			slog.String("start_time", event.StartTime.Format(time.RFC3339)),
			slog.String("end_time", event.EndTime.Format(time.RFC3339)))
		http.Error(w, "time slot is already taken", http.StatusConflict)
		return
	}

	if err := h.storage.AddEvent(&event); err != nil {
		h.Logger.Error("failed to add event", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Info("event created successfully",
		slog.String("title", event.Title),
		slog.String("end_time", event.EndTime.Format(time.RFC3339)),
		slog.String("start_time", event.StartTime.Format(time.RFC3339)),
	)
	w.WriteHeader(http.StatusCreated)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		h.Logger.Error("invalid event ID", sl.Err(err))
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.Logger.Error("failed to decode request body", sl.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event.ID = id

	if !validateEventsTime(event.StartTime, event.EndTime, h.Logger, w) {
		return
	}

	isTaken, err := h.storage.IsTimeSlotTaken(event.StartTime, event.EndTime)
	if err != nil {
		h.Logger.Error("failed to check time slot", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isTaken {
		h.Logger.Error("time slot is already taken",
			slog.String("start_time", event.StartTime.Format(time.RFC3339)),
			slog.String("end_time", event.EndTime.Format(time.RFC3339)))
		http.Error(w, "time slot is already taken", http.StatusConflict)
		return
	}

	if err := h.storage.UpdateEvent(&event); err != nil {
		h.Logger.Error("failed to update event", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Info("event updated successfully",
		slog.String("id", id.Hex()),
		slog.String("start_time", event.StartTime.Format(time.RFC3339)))
	w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		h.Logger.Error("invalid event ID", sl.Err(err))
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteEvent(id); err != nil {
		h.Logger.Error("failed to delete event", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Info("event deleted successfully", slog.String("id", id.Hex()))
	w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		h.Logger.Error("invalid event ID", sl.Err(err))
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.storage.GetEventByID(id)
	if err != nil {
		h.Logger.Error("failed to get event by ID", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if event == nil {
		h.Logger.Warn("event not found", slog.String("id", id.Hex()))
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	h.Logger.Info("event retrieved successfully", slog.String("id", id.Hex()))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.Logger.Error("failed to encode response", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.storage.GetAllEvent()
	if err != nil {
		h.Logger.Error("failed to get all events", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Info("all events retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.Logger.Error("failed to encode events", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func validateEventsTime(startTime, endTime time.Time, logger *slog.Logger, w http.ResponseWriter) bool {
	// Проверяем, если событие начинается и заканчивается в один день
	if startTime.Year() == endTime.Year() &&
		startTime.Month() == endTime.Month() &&
		startTime.Day() == endTime.Day() {

		// Добавляем проверку, чтобы убедиться, что start_time меньше end_time
		if startTime.After(endTime) || startTime.Equal(endTime) {
			logger.Error("start_time must be before end_time",
				slog.String("start_time", startTime.Format(time.RFC3339)),
				slog.String("end_time", endTime.Format(time.RFC3339)))
			http.Error(w, "start_time must be before end_time", http.StatusBadRequest)
			return false
		}
	}
	return true
}
