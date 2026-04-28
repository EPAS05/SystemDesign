package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type NodeHandler struct {
	Repo repository.NodeRepository
}

type CreateNodeRequest struct {
	Name      string `json:"name"`
	ParentID  *int   `json:"parent_id,omitempty"`
	UnitID    *int   `json:"unit_id,omitempty"`
	SortOrder *int   `json:"sort_order,omitempty"`
}

type SetParentRequest struct {
	NewParentID *int `json:"new_parent_id,omitempty"`
}

type SetNameRequest struct {
	Name string `json:"name"`
}

type SetOrderRequest struct {
	Order int `json:"order"`
}

func (h *NodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req CreateNodeRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	createReq := models.CreateNodeRequest{
		Name:      req.Name,
		ParentID:  req.ParentID,
		UnitID:    req.UnitID,
		SortOrder: req.SortOrder,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	node, err := h.Repo.CreateNode(ctx, createReq)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, node)
}

func (h *NodeHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	node, err := h.Repo.GetNode(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, node)
}

func (h *NodeHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	children, err := h.Repo.GetChildren(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, children)
}

func (h *NodeHandler) GetAllDescendants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	descendants, err := h.Repo.GetAllDescendants(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, descendants)
}

func (h *NodeHandler) GetAllAncestors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	ancestors, err := h.Repo.GetAllAncestors(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, ancestors)
}

func (h *NodeHandler) GetParent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	parent, err := h.Repo.GetParent(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if parent == nil {
		response.WriteError(w, http.StatusNotFound, "Node has no parent")
		return
	}
	response.WriteJSON(w, http.StatusOK, parent)
}

func (h *NodeHandler) SetParent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req SetParentRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	modelReq := models.SetParentRequest{
		NodeId:      id,
		NewParentID: req.NewParentID,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	updatedNode, err := h.Repo.SetParent(ctx, modelReq)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else if err == repository.ErrInvalidParent {
			response.WriteError(w, http.StatusBadRequest, "Invalid parent")
		} else if err == repository.ErrCycleDetected {
			response.WriteError(w, http.StatusBadRequest, "Cycle detected")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, updatedNode)
}

func (h *NodeHandler) SetName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req SetNameRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	modelReq := models.SetNameRequest{
		NodeId: id,
		Name:   req.Name,
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	updatedNode, err := h.Repo.SetName(ctx, modelReq)
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

func (h *NodeHandler) SetNodeOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	var req SetOrderRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.SetNodeOrder(ctx, id, req.Order)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *NodeHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	err = h.Repo.DeleteNode(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			response.WriteError(w, http.StatusNotFound, "Node not found")
		} else if err == repository.ErrCannotDeleteTrash {
			response.WriteError(w, http.StatusBadRequest, "Cannot delete trash node")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *NodeHandler) GetTerminalDescendants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	ctx, cancel := response.RequestContext(r)
	defer cancel()
	terminals, err := h.Repo.GetAllTerminalDescendants(ctx, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, terminals)
}
