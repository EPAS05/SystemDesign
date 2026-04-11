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

type EnumHandler struct {
	Repo repository.EnumRepository
}

type CreateEnumRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	TypeNodeID  int     `json:"type_node_id"` // 4,5,6
}

type UpdateEnumRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	TypeNodeID  int     `json:"type_node_id"`
}

type CreateEnumValueRequest struct {
	Value     string `json:"value"`
	SortOrder *int   `json:"sort_order,omitempty"`
}

type UpdateEnumValueRequest struct {
	Value     string `json:"value"`
	SortOrder *int   `json:"sort_order,omitempty"`
}

type ReorderEnumValuesRequest struct {
	ValueIDs []int `json:"value_ids"`
}

func (h *EnumHandler) CreateEnum(w http.ResponseWriter, r *http.Request) {
	var req CreateEnumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.TypeNodeID < 4 || req.TypeNodeID > 6 {
		http.Error(w, "type_node_id must be 4,5 or 6", http.StatusBadRequest)
		return
	}
	createReq := models.CreateEnumRequest{
		Name:        req.Name,
		Description: req.Description,
		TypeNodeID:  req.TypeNodeID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	enum, err := h.Repo.CreateEnum(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(enum)
}

func (h *EnumHandler) GetEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	enum, err := h.Repo.GetEnum(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enum)
}

func (h *EnumHandler) GetAllEnums(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	enums, err := h.Repo.GetAllEnums(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enums)
}

func (h *EnumHandler) GetEnumsByTypeNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	typeNodeIDStr := vars["type_node_id"]
	typeNodeID, err := strconv.Atoi(typeNodeIDStr)
	if err != nil {
		http.Error(w, "Invalid type_node_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	enums, err := h.Repo.GetEnumsByTypeNode(ctx, typeNodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enums)
}

func (h *EnumHandler) UpdateEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateEnumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.TypeNodeID < 4 || req.TypeNodeID > 6 {
		http.Error(w, "type_node_id must be 4,5 or 6", http.StatusBadRequest)
		return
	}
	updateReq := models.UpdateEnumRequest{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		TypeNodeID:  req.TypeNodeID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.UpdateEnum(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *EnumHandler) DeleteEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.DeleteEnum(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *EnumHandler) CreateEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		http.Error(w, "Invalid enum_id", http.StatusBadRequest)
		return
	}
	var req CreateEnumValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Value == "" {
		http.Error(w, "value is required", http.StatusBadRequest)
		return
	}
	createReq := models.CreateEnumValueRequest{
		EnumID:    enumID,
		Value:     req.Value,
		SortOrder: req.SortOrder,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ev, err := h.Repo.CreateEnumValue(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ev)
}

func (h *EnumHandler) GetEnumValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		http.Error(w, "Invalid enum_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	values, err := h.Repo.GetEnumValues(ctx, enumID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(values)
}

func (h *EnumHandler) GetEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		http.Error(w, "Invalid value_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ev, err := h.Repo.GetEnumValue(ctx, valueID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum value not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ev)
}

func (h *EnumHandler) UpdateEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		http.Error(w, "Invalid value_id", http.StatusBadRequest)
		return
	}
	var req UpdateEnumValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Value == "" {
		http.Error(w, "value is required", http.StatusBadRequest)
		return
	}
	updateReq := models.UpdateEnumValueRequest{
		ID:        valueID,
		Value:     req.Value,
		SortOrder: req.SortOrder,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.UpdateEnumValue(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum value not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *EnumHandler) DeleteEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		http.Error(w, "Invalid value_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.DeleteEnumValue(ctx, valueID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Enum value not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *EnumHandler) ReorderEnumValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		http.Error(w, "Invalid enum_id", http.StatusBadRequest)
		return
	}
	var req ReorderEnumValuesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.ValueIDs) == 0 {
		http.Error(w, "value_ids cannot be empty", http.StatusBadRequest)
		return
	}
	reorderReq := models.ReorderEnumValuesRequest{
		EnumID:   enumID,
		ValueIDs: req.ValueIDs,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.ReorderEnumValues(ctx, reorderReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
