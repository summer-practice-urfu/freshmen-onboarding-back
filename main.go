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
	_ = godotenv.Load()

	app := app{
		router: mux.NewRouter(),
		server: &http.Server{},
		logger: log.New(os.Stdout, "web ", log.LstdFlags),
	}

	app.router.Use(corsMiddleware)

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

	app.logger.Println("Redis url: ", os.Getenv("REDIS_URL"))
	redis, err := db.NewRedisDb()
	if err != nil {
		app.logger.Fatal("Can't connect to redis")
	} else {
		app.logger.Println("Connected to redis")
	}

	es := db.NewEsDb(app.logger)

	postStorage := storages.NewPostStorage(app.logger, base.Conn, es)
	sessionStorage := storages.NewSessionStorage(redis, app.logger)
	userPostRatingStorage := storages.NewUserPostRatingStorage(app.logger, base.Conn)

	pc := controllers.NewPostController(app.logger, postStorage, sessionStorage, userPostRatingStorage)
	pc.Register("/post", app.router)

	ac := controllers.NewAccountController(app.logger, sessionStorage)
	ac.Register("/acc", app.router)

	app.ListenAndServe(":8080")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
