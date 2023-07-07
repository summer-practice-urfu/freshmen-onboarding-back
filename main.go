package main

import (
	"TaskService/controllers"
	"TaskService/db"
	"TaskService/storages"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
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

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("No .env file found")
	}
	app := app{
		router: mux.NewRouter(),
		server: &http.Server{},
		logger: log.New(os.Stdout, "web ", log.LstdFlags),
	}

	app.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	app.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.logger.Println("serving")
	})

	base := db.Init(app.logger)
	defer base.Close()
	ts := controllers.NewPostController(app.logger, storages.NewPostStorage(base.Conn, app.logger))
	ts.Register("/post", app.router)
	app.ListenAndServe(":8080")
}
