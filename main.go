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
	"strings"
)

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
			app.logger.Println("Request on url: ", r.URL.String())
			if !strings.HasPrefix(r.URL.String(), "/acc/verify") {
				app.logger.Println("Setting content-type to json")
				w.Header().Add("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	})

	app.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("session_token"); err != nil {
			app.logger.Println("Got token in '/': ", c)
		} else {
			app.logger.Println("Got no cookie in '/'")
		}
	})

	base := db.Init(app.logger)
	defer base.Close()

	redis, err := db.NewRedisDb()
	if err != nil {
		app.logger.Fatal("Can't connect to redis")
	} else {
		app.logger.Println("Connected to redis")
	}

	sessionStorage := storages.NewSessionStorage(redis, app.logger)

	pc := controllers.NewPostController(app.logger, storages.NewPostStorage(base.Conn, app.logger), sessionStorage)
	pc.Register("/post", app.router)

	ac := controllers.NewAccountController(app.logger, sessionStorage)
	ac.Register("/acc", app.router)

	app.ListenAndServe(":8080")
}
