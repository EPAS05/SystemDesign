package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"net/http"
	"strconv"

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

type SetUnitRequest struct {
	UnitID *int `json:"unit_id,omitempty"`
}

type SetDefaultUnitRequest struct {
	UnitID *int `json:"unit_id,omitempty"`
}

func (h *UnitHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var req CreateUnitRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Multiplier <= 0 {
		response.WriteError(w, http.StatusBadRequest, "multiplier must be positive")
		return
	}
	createReq := models.CreateUnitRequest{
		Name:       req.Name,
		Multiplier: req.Multiplier,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	unit, err := h.Repo.CreateUnit(ctx, createReq)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, unit)
}

func (h *UnitHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	unit, err := h.Repo.GetUnit(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Unit not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, unit)
}

func (h *UnitHandler) GetAllUnits(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	units, err := h.Repo.GetAllUnits(ctx)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, units)
}

func (h *UnitHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req UpdateUnitRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Multiplier <= 0 {
		response.WriteError(w, http.StatusBadRequest, "multiplier must be positive")
		return
	}
	updateReq := models.UpdateUnitRequest{
		ID:         id,
		Name:       req.Name,
		Multiplier: req.Multiplier,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	unit, err := h.Repo.UpdateUnit(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Unit not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, unit)
}

func (h *UnitHandler) SetUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req SetUnitRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	updatedNode, err := h.Repo.SetUnit(ctx, models.SetUnitRequest{NodeId: id, UnitID: req.UnitID})
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, updatedNode)
}

func (h *UnitHandler) SetDefaultUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req SetDefaultUnitRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	updatedProduct, err := h.Repo.SetDefaultUnit(ctx, models.SetDefaultUnitRequest{ProductID: id, UnitID: req.UnitID})
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Product not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, updatedProduct)
}

func (h *UnitHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.DeleteUnit(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Unit not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
