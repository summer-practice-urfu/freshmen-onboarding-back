package controllers

import (
	"TaskService/models"
	"TaskService/storages"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type PostController struct {
	logger      *log.Logger
	stor        *storages.PostStorage
	sessionStor *storages.SessionStorage
}

func NewPostController(logger *log.Logger, stor *storages.PostStorage, sessionStor *storages.SessionStorage) *PostController {
	return &PostController{
		logger:      logger,
		stor:        stor,
		sessionStor: sessionStor,
	}
}

func (c *PostController) Register(basePath string, router *mux.Router) {
	router.HandleFunc(basePath, c.AddPost).Methods("POST")
	router.HandleFunc(basePath, c.GetPosts).Methods("GET")
	router.HandleFunc(basePath, c.UpdatePost).Methods("PUT")

	router.HandleFunc(basePath+"/{id}", c.GetPost).Methods("GET")
	router.HandleFunc(basePath+"/{id}", c.DeletePost).Methods("DELETE")

	router.HandleFunc(basePath+"/{id}/inc", c.Increment).Methods("PUT")
	router.HandleFunc(basePath+"/{id}/dec", c.Decrement).Methods("PUT")
}

type TokenDTO struct {
	SessionToken string `json:"sessionToken"`
}

func (c *PostController) Increment(w http.ResponseWriter, r *http.Request) {
	tokenDto, err := c.getSessionToken(w, r)
	if err != nil {
		return
	}

	if err := c.checkSessionToken(tokenDto, w); err != nil {
		return
	}

	c.logger.Println("Token in Increment(): ", tokenDto.SessionToken)

	c.changeRating(1, w, r)
}

func (c *PostController) Decrement(w http.ResponseWriter, r *http.Request) {
	tokenDto, err := c.getSessionToken(w, r)
	if err != nil {
		return
	}

	if err := c.checkSessionToken(tokenDto, w); err != nil {
		return
	}

	c.logger.Println("Token in Decrement(): ", tokenDto.SessionToken)

	c.changeRating(-1, w, r)
}

func (c *PostController) changeRating(delta int, w http.ResponseWriter, r *http.Request) {
	id, err := c.getId(w, r)
	if err != nil {
		return
	}
	post, err := c.stor.GetOne(id)
	if err != nil || post == nil {
		c.logger.Println("Unexisted id: ", id)
		http.Error(w, "Id does not exist, id: "+id, http.StatusBadRequest)
		return
	}
	post.Rating += delta
	if err := c.stor.Update(post); err != nil {
		c.logger.Println("Error updating post, \nPost: ", post, "\n Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func (c *PostController) AddPost(w http.ResponseWriter, r *http.Request) {
	var post models.PostAddDTO
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		c.logger.Println("Error decoding AddPost()\n Post: ", post, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	c.logger.Println("Decoded post in AddPost(),\n Post: ", post)

	if post.Title == nil {
		c.logger.Println("Error post create without title,\n Post: ", post)
		http.Error(w, "Creating post without title", http.StatusBadRequest)
		return
	}

	id, err := c.stor.Create(post.Title, post.Content, post.Img)

	if err != nil {
		c.logger.Println("Error creating post AddPost()\n Post: ", post, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		Id string `json:"id"`
	}{Id: id}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.logger.Println("Error encoding id AddPost()\n Post: ", post, "\nId: ", id, "\nError: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
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
		return
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

func (c *PostController) getSessionToken(w http.ResponseWriter, r *http.Request) (*TokenDTO, error) {
	var tokenDto *TokenDTO
	if err := json.NewDecoder(r.Body).Decode(&tokenDto); err != nil {
		c.logger.Println("Error decoding tokenDto in Increment(), Error: ", err.Error())
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return nil, errors.New("invalid body")
	}
	return tokenDto, nil
}

func (c *PostController) checkSessionToken(tokenDto *TokenDTO, w http.ResponseWriter) error {
	if tokenDto.SessionToken == "" {
		c.logger.Println("Got no token in Increment()")
		http.Error(w, "Got no sessonToken", http.StatusUnauthorized)
		return errors.New("no token")
	}

	if userVk, err := c.sessionStor.GetSession(tokenDto.SessionToken); err != nil || !userVk.Valid() {
		c.logger.Println("Expired token in Increment(), token: ", tokenDto.SessionToken)
		if err != nil {
			c.logger.Println("Error: ", err.Error())
		}
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return errors.New("token expired")
	}

	return nil
}
