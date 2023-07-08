package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type app struct {
	router *mux.Router
	server *http.Server
	logger *log.Logger
}

func (a *app) ListenAndServe(port string) {
	a.server.Addr = port
	a.server.Handler = a.router
	a.logger.Println("Server listening on port ", port)
	err := a.server.ListenAndServe()
	if err != nil {
		a.logger.Fatal("Can't start server, err: ", err.Error())
	}
}
