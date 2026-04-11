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

type ParameterHandler struct {
	ParamRepo   repository.ParameterRepository
	NodeRepo    repository.NodeRepository
	EnumRepo    repository.EnumRepository
	UnitRepo    repository.UnitRepository
	ProductRepo repository.ProductRepository
}

type CreateParamDefRequest struct {
	Name          string                      `json:"name"`
	Description   *string                     `json:"description,omitempty"`
	ParameterType string                      `json:"parameter_type"`
	UnitID        *int                        `json:"unit_id,omitempty"`
	EnumID        *int                        `json:"enum_id,omitempty"`
	IsRequired    bool                        `json:"is_required"`
	SortOrder     *int                        `json:"sort_order,omitempty"`
	Constraints   *models.ParameterConstraint `json:"constraints,omitempty"`
}

type UpdateParamDefRequest struct {
	Name        string                      `json:"name"`
	Description *string                     `json:"description,omitempty"`
	UnitID      *int                        `json:"unit_id,omitempty"`
	EnumID      *int                        `json:"enum_id,omitempty"`
	IsRequired  *bool                       `json:"is_required,omitempty"`
	SortOrder   *int                        `json:"sort_order,omitempty"`
	Constraints *models.ParameterConstraint `json:"constraints,omitempty"`
}

type SetParameterValueRequest struct {
	ProductID    int      `json:"product_id,omitempty"`
	ParamDefID   int      `json:"param_def_id"`
	ValueNumeric *float64 `json:"value_numeric,omitempty"`
	ValueEnumID  *int     `json:"value_enum_id,omitempty"`
}

type UpdateParameterValueRequest struct {
	ValueNumeric *float64 `json:"value_numeric,omitempty"`
	ValueEnumID  *int     `json:"value_enum_id,omitempty"`
}

func (h *ParameterHandler) CreateParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		http.Error(w, "Invalid node_id", http.StatusBadRequest)
		return
	}

	var req CreateParamDefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.ParameterType != "number" && req.ParameterType != "enum" {
		http.Error(w, "parameter_type must be 'number' or 'enum'", http.StatusBadRequest)
		return
	}
	if req.ParameterType == "enum" && req.EnumID == nil {
		http.Error(w, "enum_id is required for enum parameter", http.StatusBadRequest)
		return
	}
	if req.ParameterType == "enum" && req.UnitID != nil {
		http.Error(w, "unit_id is not allowed for enum parameter", http.StatusBadRequest)
		return
	}
	if req.ParameterType == "number" && req.EnumID != nil {
		http.Error(w, "enum_id is not allowed for number parameter", http.StatusBadRequest)
		return
	}
	if req.Constraints != nil && req.Constraints.MinValue != nil && req.Constraints.MaxValue != nil {
		if *req.Constraints.MinValue > *req.Constraints.MaxValue {
			http.Error(w, "min_value cannot be greater than max_value", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = h.NodeRepo.GetNode(ctx, nodeID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Node not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if req.UnitID != nil {
		_, err = h.UnitRepo.GetUnit(ctx, *req.UnitID)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Unit not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	if req.EnumID != nil {
		_, err = h.EnumRepo.GetEnum(ctx, *req.EnumID)
		if err != nil {
			if err == repository.ErrNotFound {
				http.Error(w, "Enum not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	createReq := models.CreateParameterDefinitionRequest{
		ClassNodeID:   nodeID,
		Name:          req.Name,
		Description:   req.Description,
		ParameterType: req.ParameterType,
		UnitID:        req.UnitID,
		EnumID:        req.EnumID,
		IsRequired:    req.IsRequired,
		SortOrder:     req.SortOrder,
		Constraints:   req.Constraints,
	}
	param, err := h.ParamRepo.CreateParameterDefinition(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(param)
}

func (h *ParameterHandler) GetParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	param, err := h.ParamRepo.GetParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter definition not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(param)
}

func (h *ParameterHandler) GetParameterDefinitionsForClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		http.Error(w, "Invalid node_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	params, err := h.ParamRepo.GetParameterDefinitionsForClass(ctx, nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(params)
}

func (h *ParameterHandler) GetParameterConstraints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	constraints, err := h.ParamRepo.GetParameterConstraints(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter constraints not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(constraints)
}

func (h *ParameterHandler) UpdateParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateParamDefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	param, err := h.ParamRepo.GetParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter definition not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	name := param.Name
	if req.Name != "" {
		name = req.Name
	}

	description := param.Description
	if req.Description != nil {
		description = req.Description
	}

	unitID := param.UnitID
	if req.UnitID != nil {
		unitID = req.UnitID
	}

	enumID := param.EnumID
	if req.EnumID != nil {
		enumID = req.EnumID
	}

	isRequired := param.IsRequired
	if req.IsRequired != nil {
		isRequired = *req.IsRequired
	}

	sortOrder := param.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	constraints := req.Constraints
	if constraints == nil {
		constraints, _ = h.ParamRepo.GetParameterConstraints(ctx, id)
	}
	if constraints != nil && constraints.MinValue != nil && constraints.MaxValue != nil {
		if *constraints.MinValue > *constraints.MaxValue {
			http.Error(w, "min_value cannot be greater than max_value", http.StatusBadRequest)
			return
		}
	}

	if param.ParameterType == "number" {
		if req.EnumID != nil {
			http.Error(w, "enum_id is not allowed for number parameter", http.StatusBadRequest)
			return
		}
		if req.UnitID != nil {
			_, err = h.UnitRepo.GetUnit(ctx, *req.UnitID)
			if err != nil {
				if err == repository.ErrNotFound {
					http.Error(w, "Unit not found", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
	}

	if param.ParameterType == "enum" {
		if req.UnitID != nil {
			http.Error(w, "unit_id is not allowed for enum parameter", http.StatusBadRequest)
			return
		}
		if req.EnumID != nil {
			_, err = h.EnumRepo.GetEnum(ctx, *req.EnumID)
			if err != nil {
				if err == repository.ErrNotFound {
					http.Error(w, "Enum not found", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
	}

	updateReq := models.UpdateParameterDefinitionRequest{
		ID:          id,
		Name:        name,
		Description: description,
		UnitID:      unitID,
		EnumID:      enumID,
		IsRequired:  isRequired,
		SortOrder:   &sortOrder,
		Constraints: constraints,
	}

	err = h.ParamRepo.UpdateParameterDefinition(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter definition not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ParameterHandler) DeleteParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.ParamRepo.DeleteParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter definition not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ParameterHandler) SetParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product_id", http.StatusBadRequest)
		return
	}

	var req SetParameterValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ProductID != 0 && req.ProductID != productID {
		http.Error(w, "product_id mismatch", http.StatusBadRequest)
		return
	}
	if req.ParamDefID == 0 {
		http.Error(w, "param_def_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	product, err := h.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	params, err := h.ParamRepo.GetParameterDefinitionsForClass(ctx, product.ClassNodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var target *models.ParameterDefinition
	for _, p := range params {
		if p.ID == req.ParamDefID {
			target = p
			break
		}
	}
	if target == nil {
		http.Error(w, "Parameter does not belong to this product class", http.StatusBadRequest)
		return
	}

	if target.ParameterType == "number" {
		if req.ValueNumeric == nil {
			http.Error(w, "value_numeric is required for number parameter", http.StatusBadRequest)
			return
		}
		if req.ValueEnumID != nil {
			http.Error(w, "value_enum_id is not allowed for number parameter", http.StatusBadRequest)
			return
		}
		constraints, err := h.ParamRepo.GetParameterConstraints(ctx, target.ID)
		if err == nil && constraints != nil {
			if constraints.MinValue != nil && *req.ValueNumeric < *constraints.MinValue {
				http.Error(w, "value_numeric is below minimum", http.StatusBadRequest)
				return
			}
			if constraints.MaxValue != nil && *req.ValueNumeric > *constraints.MaxValue {
				http.Error(w, "value_numeric is above maximum", http.StatusBadRequest)
				return
			}
		}
	}

	if target.ParameterType == "enum" {
		if req.ValueEnumID == nil {
			http.Error(w, "value_enum_id is required for enum parameter", http.StatusBadRequest)
			return
		}
		if req.ValueNumeric != nil {
			http.Error(w, "value_numeric is not allowed for enum parameter", http.StatusBadRequest)
			return
		}
		if target.EnumID == nil {
			http.Error(w, "Enum is not attached to this parameter", http.StatusBadRequest)
			return
		}
		values, err := h.EnumRepo.GetEnumValues(ctx, *target.EnumID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ok := false
		for _, v := range values {
			if v.ID == *req.ValueEnumID {
				ok = true
				break
			}
		}
		if !ok {
			http.Error(w, "Selected enum value does not belong to this enum", http.StatusBadRequest)
			return
		}
	}

	createReq := models.CreateParameterValueRequest{
		ProductID:    productID,
		ParamDefID:   req.ParamDefID,
		ValueNumeric: req.ValueNumeric,
		ValueEnumID:  req.ValueEnumID,
	}
	pv, err := h.ParamRepo.SetParameterValue(ctx, createReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pv)
}

func (h *ParameterHandler) GetParameterValuesForProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product_id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	values, err := h.ParamRepo.GetParameterValuesForProduct(ctx, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(values)
}

func (h *ParameterHandler) UpdateParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateParameterValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ValueNumeric == nil && req.ValueEnumID == nil {
		http.Error(w, "value is required", http.StatusBadRequest)
		return
	}
	if req.ValueNumeric != nil && req.ValueEnumID != nil {
		http.Error(w, "only one value can be set", http.StatusBadRequest)
		return
	}
	updateReq := models.UpdateParameterValueRequest{
		ID:           id,
		ValueNumeric: req.ValueNumeric,
		ValueEnumID:  req.ValueEnumID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.ParamRepo.UpdateParameterValue(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter value not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ParameterHandler) DeleteParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = h.ParamRepo.DeleteParameterValue(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Parameter value not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
