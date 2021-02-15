package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/aflog/todolist/item"
	"github.com/aflog/todolist/repository"
	"github.com/gorilla/mux"
)

//Handler holds set up for a to do list items hadler
type Handler struct {
	storage repository.Repository
}

//New creates and sets up a new items handler
func New(s repository.Repository) (*Handler, error) {
	if s == nil {
		return nil, errors.New("storage can not be nil")
	}
	return &Handler{storage: s}, nil
}

// List searches for all items and returns them through the http response
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.storage.GetItems(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "We could not retrieve the to do list items.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

// Select searches for an item based on an id from the request url and returns it in the http response
// Returns StatusNotFound if requested item id does not exist
func (h *Handler) Select(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	items, err := h.storage.GetItem(r.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(w, "We could not retrieve the requested item.", http.StatusInternalServerError)
		return
	}
	if items == nil {
		http.Error(w, "Item not found.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

// Add stores new item and returns its id in the http response
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	// get item from request body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var inItem item.Item
	err = json.Unmarshal(b, &inItem)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// validate item data
	err = inItem.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create item
	id, err := h.storage.CreateItem(r.Context(), inItem)
	if err != nil {
		log.Println(err)
		http.Error(w, "We could not create new item.", http.StatusInternalServerError)
		return
	}
	// send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		ID int `json:"id"`
	}{id}
	json.NewEncoder(w).Encode(response)
}
