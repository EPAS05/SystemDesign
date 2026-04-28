package api_handlers

import (
	"classifier/internal/http/response"
	"classifier/internal/models"
	"classifier/internal/repository"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CustomerHandler struct {
	Repo repository.CustomerRepository
}

type CreateCustomerRequest struct {
	Name    string  `json:"name"`
	TaxID   *string `json:"tax_id,omitempty"`
	Address *string `json:"address,omitempty"`
}

type UpdateCustomerRequest struct {
	Name    *string `json:"name,omitempty"`
	TaxID   *string `json:"tax_id,omitempty"`
	Address *string `json:"address,omitempty"`
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
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

	customer, err := h.Repo.CreateCustomer(ctx, models.CreateCustomerRequest{
		Name:    req.Name,
		TaxID:   req.TaxID,
		Address: req.Address,
	})
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, customer)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	customer, err := h.Repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			response.WriteError(w, http.StatusNotFound, "Customer not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := response.RequestContext(r)
	defer cancel()

	customers, err := h.Repo.GetAllCustomers(ctx)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, customers)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	var req UpdateCustomerRequest
	if err := response.ReadJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	current, err := h.Repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			response.WriteError(w, http.StatusNotFound, "Customer not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	if name == "" {
		response.WriteError(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	taxID := current.TaxID
	if req.TaxID != nil {
		taxID = req.TaxID
	}

	address := current.Address
	if req.Address != nil {
		address = req.Address
	}

	customer, err := h.Repo.UpdateCustomer(ctx, models.UpdateCustomerRequest{
		ID:      id,
		Name:    name,
		TaxID:   taxID,
		Address: address,
	})
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			response.WriteError(w, http.StatusNotFound, "Customer not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	ctx, cancel := response.RequestContext(r)
	defer cancel()

	err = h.Repo.DeleteCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			response.WriteError(w, http.StatusNotFound, "Customer not found")
		} else {
			response.WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
