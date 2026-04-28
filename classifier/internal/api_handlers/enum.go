package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"

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
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.TypeNodeID < 4 || req.TypeNodeID > 6 {
		response.WriteError(w, http.StatusBadRequest, "type_node_id must be 4,5 or 6")
		return
	}
	createReq := models.CreateEnumRequest{
		Name:        req.Name,
		Description: req.Description,
		TypeNodeID:  req.TypeNodeID,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	enum, err := h.Repo.CreateEnum(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.WriteJSON(w, http.StatusCreated, enum)
}

func (h *EnumHandler) GetEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	enum, err := h.Repo.GetEnum(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, enum)
}

func (h *EnumHandler) GetAllEnums(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	enums, err := h.Repo.GetAllEnums(ctx)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, enums)
}

func (h *EnumHandler) GetEnumsByTypeNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	typeNodeIDStr := vars["type_node_id"]
	typeNodeID, err := strconv.Atoi(typeNodeIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid type_node_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	enums, err := h.Repo.GetEnumsByTypeNode(ctx, typeNodeID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, enums)
}

func (h *EnumHandler) UpdateEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req UpdateEnumRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.TypeNodeID < 4 || req.TypeNodeID > 6 {
		response.WriteError(w, http.StatusBadRequest, "type_node_id must be 4,5 or 6")
		return
	}
	updateReq := models.UpdateEnumRequest{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		TypeNodeID:  req.TypeNodeID,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	enum, err := h.Repo.UpdateEnum(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, enum)
}

func (h *EnumHandler) DeleteEnum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.DeleteEnum(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *EnumHandler) CreateEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid enum_id")
		return
	}
	var req CreateEnumValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Value == "" {
		response.WriteError(w, http.StatusBadRequest, "value is required")
		return
	}
	createReq := models.CreateEnumValueRequest{
		EnumID:    enumID,
		Value:     req.Value,
		SortOrder: req.SortOrder,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	ev, err := h.Repo.CreateEnumValue(ctx, createReq)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, ev)
}

func (h *EnumHandler) GetEnumValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid enum_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	values, err := h.Repo.GetEnumValues(ctx, enumID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, values)
}

func (h *EnumHandler) GetEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid value_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	ev, err := h.Repo.GetEnumValue(ctx, valueID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum value not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, ev)
}

func (h *EnumHandler) UpdateEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid value_id")
		return
	}
	var req UpdateEnumValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Value == "" {
		response.WriteError(w, http.StatusBadRequest, "value is required")
		return
	}
	updateReq := models.UpdateEnumValueRequest{
		ID:        valueID,
		Value:     req.Value,
		SortOrder: req.SortOrder,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	ev, err := h.Repo.UpdateEnumValue(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum value not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, ev)
}

func (h *EnumHandler) DeleteEnumValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	valueIDStr := vars["value_id"]
	valueID, err := strconv.Atoi(valueIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid value_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.DeleteEnumValue(ctx, valueID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Enum value not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *EnumHandler) ReorderEnumValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enumIDStr := vars["enum_id"]
	enumID, err := strconv.Atoi(enumIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid enum_id")
		return
	}
	var req ReorderEnumValuesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.ValueIDs) == 0 {
		response.WriteError(w, http.StatusBadRequest, "value_ids cannot be empty")
		return
	}
	reorderReq := models.ReorderEnumValuesRequest{
		EnumID:   enumID,
		ValueIDs: req.ValueIDs,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.ReorderEnumValues(ctx, reorderReq)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
