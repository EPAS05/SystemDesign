package api_handlers

import (
	"classifier/internal/models"
	"classifier/internal/repository"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type UnitHandler struct {
	Repo repository.UnitRepository
}

type CreateUnitRequest struct {
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

type UpdateUnitRequest struct {
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
}

func (h *UnitHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var req CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Multiplier <= 0 {
		http.Error(w, "multiplier must be positive", http.StatusBadRequest)
		return
	}
	createReq := models.CreateUnitRequest{
		Name:       req.Name,
		Multiplier: req.Multiplier,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	unit, err := h.Repo.CreateUnit(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(unit)
}

func (h *UnitHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	unit, err := h.Repo.GetUnit(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Unit not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *UnitHandler) GetAllUnits(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	units, err := h.Repo.GetAllUnits(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *UnitHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.Multiplier <= 0 {
		http.Error(w, "multiplier must be positive", http.StatusBadRequest)
		return
	}
	updateReq := models.UpdateUnitRequest{
		ID:         id,
		Name:       req.Name,
		Multiplier: req.Multiplier,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.UpdateUnit(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Unit not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *UnitHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.DeleteUnit(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Unit not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
