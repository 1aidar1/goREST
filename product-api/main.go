package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/1aidar1/goREST/data"
	"github.com/1aidar1/goREST/handlers"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const bindAddress = ":9090"

func main() {

	l := log.New(os.Stdout, "products-api ", log.LstdFlags)
	dbLog := log.New(os.Stdout, "postgres ", log.LstdFlags)

	env, err := godotenv.Read()
	if err != nil {
		l.Println("Can't locate .env")
		return
	}
	var (
		db_host     = env["DB_HOST"]
		db_user     = env["DB_USER"]
		db_name     = env["DB_NAME"]
		db_password = env["DB_PASSWORD"]
		db_port     = env["DB_PORT"]
	)

	db_uri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", db_host, db_user, db_name, db_password, db_port)

	db, err := sqlx.Connect("postgres", db_uri)

	if err != nil {
		dbLog.Fatal("[ERROR] can't connect to postgres")
		return
	}
	defer db.Close()

	data.SetDB(db)

	// create the handlers
	ph := handlers.NewBooks(l)

	// create a new serve mux and register the handlers
	sm := mux.NewRouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/books", ph.GetBooks)

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/books/{id:[0-9]+}", ph.UpdateBooks)
	putRouter.Use(ph.MiddlewareValidateBook)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/books", ph.AddBook)
	postRouter.Use(ph.MiddlewareValidateBook)

	deleteRouter := sm.Methods(http.MethodDelete).Subrouter()
	deleteRouter.HandleFunc("/books/{id:[0-9]+}", ph.DeleteBook)

	options := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	docsHandler := middleware.Redoc(options, nil)
	getRouter.Handle("/docs", docsHandler)
	getRouter.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	// create a new server
	s := http.Server{
		Addr:         bindAddress,       // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     l,                 // set the logger for the server
		ReadTimeout:  5 * time.Second,   // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}

	// start the server
	go func() {
		l.Println("Starting server on port 9090")

		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}
