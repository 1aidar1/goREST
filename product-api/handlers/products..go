// Package classification of Product APi
//
// Documentation for Product API
//
// Schemes: http
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - appllication/json
//
// Produces:
// - appllication/json
// swagger:meta
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/1aidar1/goREST/data"
	"github.com/gorilla/mux"
)

// Returns a list of products (description of response)
// swagger:response productsResponse
type productsResponse struct {
	// All products in the system
	// in:body
	Body []data.Product
}

//swagger:parameters deleteProduct
type productIDParameterWrapper struct {
	// Id of the product that should be deleted
	//in:path
	//required:true
	ID int `json:id`
}

//swagger:response noContent
type productNoContent struct {
}

// Products is a http.Handler
type Products struct {
	l *log.Logger
}

// NewProducts creates a products handler with the given logger
func NewProducts(l *log.Logger) *Products {
	return &Products{l}
}

//swagger:route GET /products products listProducts
//Returns a list of products
//responses:
//	200: productsResponse

// GetProducts returns the products from the data store

func (p *Products) GetProducts(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET Products")

	// fetch the products from the datastore
	lp := data.GetProducts()

	// serialize the list to JSON
	err := lp.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
	}
}

//swagger:route POST /products products addProduct
//Adds a product
//responses:
//	200: OK created
//	400: BadRequest check json
func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST Product")

	prod := r.Context().Value(&KeyProduct).(data.Product)
	data.AddProduct(&prod)
}

//swagger:route PUT /products/{id} products updateProduct
//Updates a product with given id
//responses:
//	200: OK updated
//	400: BadRequest check json
func (p Products) UpdateProducts(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "Unable to convert id", http.StatusBadRequest)
		return
	}

	p.l.Println("Handle PUT Product", id)

	prod := r.Context().Value(&KeyProduct).(data.Product)

	err = data.UpdateProduct(id, &prod)
	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(rw, "Product not found", http.StatusInternalServerError)
		return
	}
}

//swagger:route DELETE /products/{id} products deleteProduct
// Deletes product with given id
//responses:
//	200: noContent deleted
//	400: BadRequest bad id
func (p *Products) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		p.l.Println("Can't parse id")
		http.Error(w, fmt.Sprintf("Can't parse id. %s", err), http.StatusBadRequest)
	}
	err = data.DeleteProduct(id)
	if err != nil {
		p.l.Println("[ERROR] ", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
	}

}

var KeyProduct struct{}

func (p Products) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}

		err := prod.FromJSON(r.Body)
		if err != nil {
			p.l.Println("[ERROR] deserializing product", err)
			http.Error(rw, "Error reading product", http.StatusBadRequest)
			return
		}
		//validate json
		if err := prod.Validete(); err != nil {
			p.l.Println("[ERROR] bad json", err)
			http.Error(rw, fmt.Sprintf("JSON validation failed: %s", err), http.StatusBadRequest)
			return
		}

		// add the product to the context
		ctx := context.WithValue(r.Context(), &KeyProduct, prod)
		r = r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
