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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	class, err := h.NodeRepo.GetNode(ctx, req.ClassNodeID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Class node not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if class.IsTerminal != nil && *class.IsTerminal == false {
		http.Error(w, "Class is non-terminal, cannot add product", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if class.IsTerminal == nil {
		_ = h.NodeRepo.UpdateNodeIsTerminal(ctx, req.ClassNodeID, boolPtr(true))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	product, err := h.Repo.GetProduct(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) GetProductsByClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	classIDStr := vars["node_id"]
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid node_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	products, err := h.Repo.GetProductsByClass(ctx, classID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.Repo.UpdateProduct(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	product, err := h.Repo.GetProduct(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	classID := product.ClassNodeID

	err = h.Repo.DeleteProduct(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	products, err := h.Repo.GetProductsByClass(ctx, classID)
	if err != nil {
	} else if len(products) == 0 {
		_ = h.NodeRepo.UpdateNodeIsTerminal(ctx, classID, nil)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func boolPtr(b bool) *bool {
	return &b
}
