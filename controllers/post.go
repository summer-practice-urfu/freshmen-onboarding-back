package controllers

import (
	"TaskService/models"
	"TaskService/storages"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type PostController struct {
	logger *log.Logger
	stor   *storages.PostStorage
}

func NewPostController(logger *log.Logger, stor *storages.PostStorage) *PostController {
	return &PostController{
		logger: logger,
		stor:   stor,
	}
}

func (c *PostController) Register(basePath string, router *mux.Router) {
	router.HandleFunc(basePath, c.AddPost).Methods("POST")
	router.HandleFunc(basePath, c.GetPosts).Methods("GET")
	router.HandleFunc(basePath+"/{id}", c.GetPost).Methods("GET")
	router.HandleFunc(basePath+"/{id}", c.DeletePost).Methods("DELETE")
	router.HandleFunc(basePath, c.UpdatePost).Methods("PUT")

}

type PostAddDTO struct {
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Img     *string `json:"image"`
}

func (c *PostController) AddPost(w http.ResponseWriter, r *http.Request) {
	var post PostAddDTO
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		c.logger.Println("Error decoding AddPost()\n Post: ", post, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
	id, err := c.stor.Create(post.Title, post.Content, post.Img)

	if err != nil {
		c.logger.Println("Error creating post AddPost()\n Post: ", post, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
	resp := struct {
		Id string `json:"id"`
	}{Id: id}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.logger.Println("Error encoding id AddPost()\n Post: ", post, "\nId: ", id, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (c *PostController) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := c.getId(w, r)
	if err != nil {
		return
	}
	post, err := c.stor.GetOne(id)
	if err != nil {
		c.logger.Println("Error getting post in GetOne() \nError: ", err.Error())
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(post); err != nil {
		c.logger.Println("Error encoding post in GetOne() \nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (c *PostController) GetPosts(w http.ResponseWriter, _ *http.Request) {
	limit := 100
	tasks, err := c.stor.GetAll(limit)
	if err != nil {
		c.logger.Println("Error getting posts in GetPosts() \nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		c.logger.Println("Error encoding posts in GetPosts() \nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func (c *PostController) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := c.getId(w, r)
	if err != nil {
		return
	}

	if err := c.stor.Delete(id); err != nil {
		c.logger.Println("Error deleting post in DeletePost() \nError: ", err.Error())
		http.Error(w, "id not found", http.StatusBadRequest)
	}
}

func (c *PostController) UpdatePost(w http.ResponseWriter, r *http.Request) {
	var newTask *models.Post
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		c.logger.Println("Error decoding task: ", newTask, "err: ", err.Error())
		http.Error(w, "Invalid task", http.StatusBadRequest)
		return
	}
	if err := c.stor.Update(newTask); err != nil {
		c.logger.Println("Error updating task: ", newTask, "err: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func (c *PostController) getId(w http.ResponseWriter, r *http.Request) (string, error) {
	id := mux.Vars(r)["id"]
	if _, err := uuid.Parse(id); err != nil {
		c.logger.Println("Error parsing id in getId(), \nId: ", id, "\nError: ", err.Error())
		http.Error(w, "id incorrect", http.StatusBadRequest)
		return id, err
	}
	return id, nil
}
