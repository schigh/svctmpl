package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/myservice/internal/model"
)

// createRequest is the JSON body for resource creation.
type createRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// updateRequest is the JSON body for resource updates.
type updateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// createResource returns a handler for POST /api/resources.
func createResource(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}

		res, err := svc.CreateResource(r.Context(), req.Name, req.Description)
		if err != nil {
			respondServiceError(w, err)
			return
		}

		respondJSON(w, http.StatusCreated, res)
	}
}

// listResources returns a handler for GET /api/resources.
func listResources(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resources, err := svc.ListResources(r.Context())
		if err != nil {
			respondServiceError(w, err)
			return
		}
		if resources == nil {
			resources = make([]*model.Resource, 0)
		}
		respondJSON(w, http.StatusOK, resources)
	}
}

// getResource returns a handler for GET /api/resources/{id}.
func getResource(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		res, err := svc.GetResource(r.Context(), id)
		if err != nil {
			respondServiceError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, res)
	}
}

// updateResource returns a handler for PUT /api/resources/{id}.
func updateResource(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		var req updateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}

		res, err := svc.UpdateResource(r.Context(), id, req.Name, req.Description)
		if err != nil {
			respondServiceError(w, err)
			return
		}

		respondJSON(w, http.StatusOK, res)
	}
}

// deleteResource returns a handler for DELETE /api/resources/{id}.
func deleteResource(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := svc.DeleteResource(r.Context(), id)
		if err != nil {
			respondServiceError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
