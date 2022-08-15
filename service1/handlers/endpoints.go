package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"service1/models"
	"service1/services"

	validation "github.com/go-ozzo/ozzo-validation"
)

// ErrNotCorrectMsg .
var ErrNotCorrectMsg = errors.New("not correct msg")

// Handler .
type Handler struct {
	service services.Service
}

// NewHandler .
func NewHandler(s services.Service) *Handler {
	return &Handler{service: s}
}

// IncrementByHandler .
func (h *Handler) IncrementByHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msgin models.IncrMsgIn
		if err := json.NewDecoder(r.Body).Decode(&msgin); err != nil {
			respondError(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := msgin.Validate(); err != nil {
			respondError(w, r, http.StatusBadRequest, err)
			return
		}

		res, err := h.service.IncrementBy(r.Context(), msgin.Key, msgin.Val)
		if err != nil {
			respondError(w, r, http.StatusInternalServerError, err)
			return
		}

		respond(w, r, http.StatusOK, res)
	}

}

// HashStringHandler .
func (h *Handler) HashStringHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msgin *models.HashMsgIn

		if err := json.NewDecoder(r.Body).Decode(&msgin); err != nil {
			fmt.Println(err)
			respondError(w, r, http.StatusInternalServerError, ErrNotCorrectMsg)
			return
		}

		if err := msgin.Validate(); err != nil {
			fmt.Println(err)
			respondError(w, r, http.StatusBadRequest, ErrNotCorrectMsg)
			return
		}

		res := h.service.HashString(r.Context(), msgin.S, msgin.Key)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(res))
	}
}

// MulStringValHandler .
func (h *Handler) MulStringValHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var pairs []*models.Pair
		if err := json.NewDecoder(r.Body).Decode(&pairs); err != nil {
			fmt.Println(err)
			respondError(w, r, http.StatusInternalServerError, ErrNotCorrectMsg)
			return
		}

		if err := validation.Validate(pairs); err != nil {
			fmt.Println(err)
			respondError(w, r, http.StatusBadRequest, ErrNotCorrectMsg)
			return
		}

		res, err := h.service.MulStringVal(r.Context(), pairs)
		if err != nil {
			respondError(w, r, http.StatusInternalServerError, err)
			return
		}

		respond(w, r, http.StatusOK, res)
	}
}

func respondError(w http.ResponseWriter, r *http.Request, code int, err error) {
	respond(w, r, code, map[string]string{"error": err.Error()})
}

func respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
