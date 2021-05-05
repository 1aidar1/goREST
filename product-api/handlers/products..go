// Package classification of Book APi
//
// Documentation for Book API
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

// Returns a list of books (description of response)
// swagger:response booksResponse
type booksResponse struct {
	// All books in the system
	// in:body
	Body []data.Book
}

//swagger:parameters deleteBook
type bookIDParameterWrapper struct {
	// Id of the book that should be deleted
	//in:path
	//required:true
	ID int `json:id`
}

//swagger:response noContent
type bookNoContent struct {
}

// Books is a http.Handler
type Books struct {
	l *log.Logger
}

// NewBooks creates a books handler with the given logger
func NewBooks(l *log.Logger) *Books {
	return &Books{l}
}

//swagger:route GET /books books listBooks
//Returns a list of books
//responses:
//	200: booksResponse

// GetBooks returns the books from the data store

func (p *Books) GetBooks(w http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle GET Books")

	// fetch the books from the datastore
	books, err := data.GetBooks()

	if err != nil {
		http.Error(w, "[DB_ERROR]", http.StatusInternalServerError)
	}

	// serialize the list to JSON
	err = books.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
	}
}

//swagger:route POST /books books addBook
//Adds a book
//responses:
//	200: OK created
//	400: BadRequest check json
func (p *Books) AddBook(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("Handle POST Book")

	prod := r.Context().Value(&KeyBook).(data.Book)
	err := data.AddBook(&prod)

	if err != nil {
		p.l.Println(err)
	}
}

//swagger:route PUT /books/{id} books updateBook
//Updates a book with given id
//responses:
//	200: OK updated
//	400: BadRequest check json
func (p Books) UpdateBooks(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "Unable to convert id", http.StatusBadRequest)
		return
	}

	p.l.Println("Handle PUT Book", id)

	prod := r.Context().Value(&KeyBook).(data.Book)

	err = data.UpdateBook(id, &prod)
	// if err == data.ErrBookNotFound {
	// 	http.Error(rw, "Book not found", http.StatusNotFound)
	// 	return
	// }
	if err != nil {
		http.Error(rw, "Book not found", http.StatusInternalServerError)
		return
	}
}

//swagger:route DELETE /books/{id} books deleteBook
// Deletes book with given id
//responses:
//	200: noContent deleted
//	400: BadRequest bad id
func (p *Books) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		p.l.Println("Can't parse id")
		http.Error(w, fmt.Sprintf("Can't parse id. %s", err), http.StatusBadRequest)
	}
	err = data.DeleteBook(id)
	if err != nil {
		p.l.Println("[ERROR] ", err)
		http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
	}

}

var KeyBook struct{}

func (p Books) MiddlewareValidateBook(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		book := data.Book{}
		fmt.Printf("%s \n", r.Body)
		err := book.FromJSON(r.Body)
		if err != nil {
			p.l.Println("[ERROR] deserializing book", err)
			http.Error(rw, "Error reading book", http.StatusBadRequest)
			return
		}
		//validate json
		if err := book.Validate(); err != nil {
			p.l.Println("[ERROR] bad json", err)
			http.Error(rw, fmt.Sprintf("JSON validation failed: %s", err), http.StatusBadRequest)
			return
		}

		// add the book to the context
		ctx := context.WithValue(r.Context(), &KeyBook, book)
		r = r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
