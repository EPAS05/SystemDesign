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

	customer, err := h.Repo.CreateCustomer(ctx, models.CreateCustomerRequest{
		Name:    req.Name,
		TaxID:   req.TaxID,
		Address: req.Address,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customer, err := h.Repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (h *CustomerHandler) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	customers, err := h.Repo.GetAllCustomers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	current, err := h.Repo.GetCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	if name == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
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

	err = h.Repo.UpdateCustomer(ctx, models.UpdateCustomerRequest{
		ID:      id,
		Name:    name,
		TaxID:   taxID,
		Address: address,
	})
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = h.Repo.DeleteCustomer(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound || err == repository.ErrCustomerNotFound {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
