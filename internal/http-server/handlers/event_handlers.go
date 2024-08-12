package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/NikKazzzzzz/Calendar-main/internal/models"

	"github.com/NikKazzzzzz/Calendar-main/internal/lib/logger/sl"
	"github.com/NikKazzzzzz/Calendar-main/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	storage *postgres.Storage
	Logger  *slog.Logger
}

func NewEventHandler(storage *postgres.Storage, logger *slog.Logger) *EventHandler {
	return &EventHandler{storage: storage, Logger: logger}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.Logger.Error("failed to decode request body", sl.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		slog.String("start_time", event.StartTime.Format(time.RFC3339)),
		slog.String("end_time", event.EndTime.Format(time.RFC3339)))
	w.WriteHeader(http.StatusCreated)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.Logger.Error("failed to decode request body", sl.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isTaken, err := h.storage.IsTimeSlotTaken(event.StartTime, event.EndTime)
	if err != nil {
		h.Logger.Error("failed to check time slot", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isTaken {
		h.Logger.Warn("time slot is already taken",
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
		slog.Int("id", event.ID),
		slog.String("start_time", event.StartTime.Format(time.RFC3339)))
	w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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

	h.Logger.Info("event deleted successfully", slog.Int("id", id))
	w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
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
		h.Logger.Warn("event not found", slog.Int("id", id))
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	h.Logger.Info("event retrieved successfully", slog.Int("id", id))
	json.NewEncoder(w).Encode(event)
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
