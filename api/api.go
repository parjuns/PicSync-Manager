package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/product", makeHTTPHandleFunc(s.handleCreateProduct)).Methods("POST")
	router.HandleFunc("/product/{id}", makeHTTPHandleFunc(s.handleGetProduct)).Methods("GET")
	router.HandleFunc("/product/{id}", makeHTTPHandleFunc(s.handleUpdateProduct)).Methods("POST")
	log.Println("API Server running on port", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleCreateProduct(w http.ResponseWriter, r *http.Request) error {

	var productParams CreateProductParams

	// decoding request body to Product object
	err := json.NewDecoder(r.Body).Decode(&productParams)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, err)
	}
	// check for missing fields
	if productParams.Name == "" || productParams.Description == "" || len(productParams.Images) == 0 || productParams.Price == "" || productParams.UserID == 0 {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "missing fields"})
	}
	//check if user id present in database
	err = s.store.CheckUserID(productParams.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "user id not found"})
		}
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}
	productid, err := s.store.CreateProduct(productParams)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	return WriteJSON(w, http.StatusCreated, fmt.Sprintf("product added successfully with product id:%d", productid))
}

func (s *APIServer) handleGetProduct(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)
	productId, err := strconv.Atoi(params["id"])
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "bad product id"})
	}

	product, err := s.store.GetProduct(productId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "product id not found"})
	}

	return WriteJSON(w, http.StatusOK, product)
}
func (s *APIServer) handleUpdateProduct(w http.ResponseWriter, r *http.Request) error {

	params := mux.Vars(r)
	productId, err := strconv.Atoi(params["id"])
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "bad product id"})
	}

	// decoding request body to Product object
	var imageLocations []string
	err = json.NewDecoder(r.Body).Decode(&imageLocations)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "payload decode error " + err.Error()})
	}

	var productParams AddProductCompressImagesParams
	productParams.ID = productId
	productParams.CompressedImages = imageLocations

	err = s.store.AddProductCompressImages(productParams)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, err)
	}

	product, err := s.store.GetProduct(productId)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	return WriteJSON(w, http.StatusOK, product)
}

// utils

type apiFunc func(http.ResponseWriter, *http.Request) error
type ApiError struct {
	Error string
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
