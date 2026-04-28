package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	Repo     repository.ProductRepository
	NodeRepo repository.NodeRepository
}

type CreateProductRequest struct {
	Name           string   `json:"name"`
	ClassNodeID    int      `json:"class_node_id"`
	UnitType       *string  `json:"unit_type,omitempty"`
	WeightPerMeter *float64 `json:"weight_per_meter,omitempty"`
	PieceLength    *float64 `json:"piece_length,omitempty"`
	DefaultUnitID  *int     `json:"default_unit_id,omitempty"`
}

type UpdateProductRequest struct {
	Name           string   `json:"name"`
	ClassNodeID    int      `json:"class_node_id"`
	UnitType       *string  `json:"unit_type,omitempty"`
	WeightPerMeter *float64 `json:"weight_per_meter,omitempty"`
	PieceLength    *float64 `json:"piece_length,omitempty"`
	DefaultUnitID  *int     `json:"default_unit_id,omitempty"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	class, err := h.NodeRepo.GetNode(ctx, req.ClassNodeID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Class node not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if class.IsTerminal != nil && *class.IsTerminal == false {
		response.WriteError(w, http.StatusBadRequest, "Class is non-terminal, cannot add product")
		return
	}

	createReq := models.CreateProductRequest{
		Name:           req.Name,
		ClassNodeID:    req.ClassNodeID,
		UnitType:       req.UnitType,
		WeightPerMeter: req.WeightPerMeter,
		PieceLength:    req.PieceLength,
		DefaultUnitID:  req.DefaultUnitID,
	}
	product, err := h.Repo.CreateProduct(ctx, createReq)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if class.IsTerminal == nil {
		_ = h.NodeRepo.UpdateNodeIsTerminal(ctx, req.ClassNodeID, boolPtr(true))
	}

	response.WriteJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	product, err := h.Repo.GetProduct(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Product not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) GetProductsByClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	classIDStr := vars["node_id"]
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid node_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	products, err := h.Repo.GetProductsByClass(ctx, classID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req UpdateProductRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	updateReq := models.UpdateProductRequest{
		ID:             id,
		Name:           req.Name,
		ClassNodeID:    req.ClassNodeID,
		UnitType:       req.UnitType,
		WeightPerMeter: req.WeightPerMeter,
		PieceLength:    req.PieceLength,
		DefaultUnitID:  req.DefaultUnitID,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.UpdateProduct(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Product not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()

	product, err := h.Repo.GetProduct(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Product not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	classID := product.ClassNodeID

	err = h.Repo.DeleteProduct(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	products, err := h.Repo.GetProductsByClass(ctx, classID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
	} else if len(products) == 0 {
		_ = h.NodeRepo.UpdateNodeIsTerminal(ctx, classID, nil)
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func boolPtr(b bool) *bool {
	return &b
}
