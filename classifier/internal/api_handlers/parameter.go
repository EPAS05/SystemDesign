package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"fmt"
	"net/http"
	"strconv"

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

type FindProductsByParametersRequest struct {
	Filters []models.ParameterFilter `json:"filters"`
}

func (h *ParameterHandler) CreateParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid node_id")
		return
	}

	var req CreateParamDefRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.ParameterType != "number" && req.ParameterType != "enum" {
		response.WriteError(w, http.StatusBadRequest, "parameter_type must be 'number' or 'enum'")
		return
	}
	if req.ParameterType == "enum" && req.EnumID == nil {
		response.WriteError(w, http.StatusBadRequest, "enum_id is required for enum parameter")
		return
	}
	if req.ParameterType == "enum" && req.UnitID != nil {
		response.WriteError(w, http.StatusBadRequest, "unit_id is not allowed for enum parameter")
		return
	}
	if req.ParameterType == "number" && req.EnumID != nil {
		response.WriteError(w, http.StatusBadRequest, "enum_id is not allowed for number parameter")
		return
	}
	if req.Constraints != nil && req.Constraints.MinValue != nil && req.Constraints.MaxValue != nil {
		if *req.Constraints.MinValue > *req.Constraints.MaxValue {
			response.WriteError(w, http.StatusBadRequest, "min_value cannot be greater than max_value")
			return
		}
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	_, err = h.NodeRepo.GetNode(ctx, nodeID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if req.UnitID != nil {
		_, err = h.UnitRepo.GetUnit(ctx, *req.UnitID)
		if err != nil {
			if err == repository.ErrNotFound {
				response.WriteError(w, http.StatusNotFound, "Unit not found")
			} else {
				response.WriteError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
	}

	if req.EnumID != nil {
		_, err = h.EnumRepo.GetEnum(ctx, *req.EnumID)
		if err != nil {
			if err == repository.ErrNotFound {
				response.WriteError(w, http.StatusNotFound, "Enum not found")
			} else {
				response.WriteError(w, http.StatusInternalServerError, err.Error())
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
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, param)
}

func (h *ParameterHandler) GetParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	param, err := h.ParamRepo.GetParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter definition not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, param)
}

func (h *ParameterHandler) GetParameterDefinitionsForClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid node_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	params, err := h.ParamRepo.GetParameterDefinitionsForClass(ctx, nodeID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, params)
}

func (h *ParameterHandler) GetParameterConstraints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	constraints, err := h.ParamRepo.GetParameterConstraints(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter constraints not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, constraints)
}

func (h *ParameterHandler) UpdateParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req UpdateParamDefRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	param, err := h.ParamRepo.GetParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter definition not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
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
					response.WriteError(w, http.StatusNotFound, "Unit not found")
				} else {
					response.WriteError(w, http.StatusInternalServerError, err.Error())
				}
				return
			}
		}
	}

	if param.ParameterType == "enum" {
		if req.UnitID != nil {
			response.WriteError(w, http.StatusBadRequest, "unit_id is not allowed for enum parameter")
			return
		}
		if req.EnumID != nil {
			_, err = h.EnumRepo.GetEnum(ctx, *req.EnumID)
			if err != nil {
				if err == repository.ErrNotFound {
					response.WriteError(w, http.StatusNotFound, "Enum not found")
				} else {
					response.WriteError(w, http.StatusInternalServerError, err.Error())
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

	updatedParam, err := h.ParamRepo.UpdateParameterDefinition(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter definition not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, updatedParam)
}

func (h *ParameterHandler) DeleteParameterDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.ParamRepo.DeleteParameterDefinition(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter definition not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ParameterHandler) SetParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product_id")
		return
	}

	var req SetParameterValueRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.ProductID != 0 && req.ProductID != productID {
		response.WriteError(w, http.StatusBadRequest, "product_id mismatch")
		return
	}
	if req.ParamDefID == 0 {
		response.WriteError(w, http.StatusBadRequest, "param_def_id is required")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	product, err := h.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Product not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	params, err := h.ParamRepo.GetParameterDefinitionsForClass(ctx, product.ClassNodeID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
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
		response.WriteError(w, http.StatusBadRequest, "Parameter does not belong to this product class")
		return
	}

	if target.ParameterType == "number" {
		if req.ValueNumeric == nil {
			response.WriteError(w, http.StatusBadRequest, "value_numeric is required for number parameter")
			return
		}
		if req.ValueEnumID != nil {
			response.WriteError(w, http.StatusBadRequest, "value_enum_id is not allowed for number parameter")
			return
		}
		constraints, err := h.ParamRepo.GetParameterConstraints(ctx, target.ID)
		if err == nil && constraints != nil {
			if constraints.MinValue != nil && *req.ValueNumeric < *constraints.MinValue {
				response.WriteError(w, http.StatusBadRequest, "value_numeric is below minimum")
				return
			}
			if constraints.MaxValue != nil && *req.ValueNumeric > *constraints.MaxValue {
				response.WriteError(w, http.StatusBadRequest, "value_numeric is above maximum")
				return
			}
		}
	}

	if target.ParameterType == "enum" {
		if req.ValueEnumID == nil {
			response.WriteError(w, http.StatusBadRequest, "value_enum_id is required for enum parameter")
			return
		}
		if req.ValueNumeric != nil {
			response.WriteError(w, http.StatusBadRequest, "value_numeric is not allowed for enum parameter")
			return
		}
		if target.EnumID == nil {
			response.WriteError(w, http.StatusBadRequest, "Enum is not attached to this parameter")
			return
		}
		values, err := h.EnumRepo.GetEnumValues(ctx, *target.EnumID)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
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
			response.WriteError(w, http.StatusBadRequest, "Selected enum value does not belong to this enum")
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
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, pv)
}

func (h *ParameterHandler) GetParameterValuesForProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product_id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	values, err := h.ParamRepo.GetParameterValuesForProduct(ctx, productID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, values)
}

func (h *ParameterHandler) UpdateParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req UpdateParameterValueRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.ValueNumeric == nil && req.ValueEnumID == nil {
		response.WriteError(w, http.StatusBadRequest, "value is required")
		return
	}
	if req.ValueNumeric != nil && req.ValueEnumID != nil {
		response.WriteError(w, http.StatusBadRequest, "only one value can be set")
		return
	}
	updateReq := models.UpdateParameterValueRequest{
		ID:           id,
		ValueNumeric: req.ValueNumeric,
		ValueEnumID:  req.ValueEnumID,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	updatedValue, err := h.ParamRepo.UpdateParameterValue(ctx, updateReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter value not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, updatedValue)
}

func (h *ParameterHandler) DeleteParameterValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.ParamRepo.DeleteParameterValue(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Parameter value not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ParameterHandler) FindProductsByParameters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil || nodeID <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid node_id")
		return
	}

	var req FindProductsByParametersRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	for i, filter := range req.Filters {
		if filter.ParamDefID <= 0 {
			response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("filters[%d].param_def_id must be greater than zero", i))
			return
		}
		switch filter.Operator {
		case "=", "<", ">", "<=", ">=":
		default:
			response.WriteError(w, http.StatusBadRequest, fmt.Sprintf("filters[%d].operator is not supported", i))
			return
		}
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	products, err := h.ParamRepo.FindProductsByParameters(ctx, nodeID, req.Filters)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "No products found matching the criteria")
			return
		}
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, products)
}
