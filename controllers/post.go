package controllers

import (
	"TaskService/models"
	"TaskService/models/cache"
	"TaskService/storages"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PostController struct {
	logger             *log.Logger
	postStor           *storages.PostStorage
	sessionStor        *storages.SessionStorage
	userPostRatingStor *storages.UserPostRatingStorage
}

func NewPostController(logger *log.Logger, stor *storages.PostStorage,
	sessionStor *storages.SessionStorage,
	userPostRatingStor *storages.UserPostRatingStorage) *PostController {
	return &PostController{
		logger:             logger,
		postStor:           stor,
		sessionStor:        sessionStor,
		userPostRatingStor: userPostRatingStor,
	}
}

func (c *PostController) Register(basePath string, router *mux.Router) {
	router.HandleFunc(basePath, c.AddPost).Methods("POST")
	router.HandleFunc(basePath, c.GetPosts).Methods("GET")
	router.HandleFunc(basePath, c.UpdatePost).Methods("PUT")

	router.HandleFunc(basePath+"/search", c.Search).Methods("GET")

	router.HandleFunc(basePath+"/{id}", c.GetPost).Methods("GET")
	router.HandleFunc(basePath+"/{id}", c.DeletePost).Methods("DELETE")

	router.HandleFunc(basePath+"/{id}/inc", c.Increment).Methods("PUT")
	router.HandleFunc(basePath+"/{id}/dec", c.Decrement).Methods("PUT")
}

type TokenDTO struct {
	SessionToken string `json:"sessionToken"`
}

func (c *PostController) Search(w http.ResponseWriter, r *http.Request) {
	size, page, search, err := c.getSearchParams(w, r)
	if err != nil {
		return
	}

	esRes, err := c.postStor.SearchES(search, size, page)
	if err != nil {
		c.logger.Println("Error in Search(), query: ", search, ", page: ", page, ", size", size, "\nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
	c.logger.Println("Got ids: ", esRes)

	posts, err := c.postStor.GetMany(esRes.Ids)
	if err != nil {
		c.logger.Println("Error getting many posts by ids in Search(), Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total": esRes.Total,
		"posts": posts,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Printf("Error encoding posts in Search(), posts: %v", posts)
	}
}

func (c *PostController) Increment(w http.ResponseWriter, r *http.Request) {
	tokenDto, err := c.getSessionToken(w, r)
	if err != nil {
		return
	}

	userVk, err := c.tryGetSession(tokenDto, w)
	if err != nil {
		return
	}

	c.logger.Println("Token in Increment(): ", tokenDto.SessionToken)

	c.changeRating('-', userVk, w, r)
}

func (c *PostController) Decrement(w http.ResponseWriter, r *http.Request) {
	tokenDto, err := c.getSessionToken(w, r)
	if err != nil {
		return
	}

	userVk, err := c.tryGetSession(tokenDto, w)
	if err != nil {
		return
	}

	c.logger.Println("Token in Decrement(): ", tokenDto.SessionToken)

	c.changeRating('+', userVk, w, r)
}

func (c *PostController) changeRating(oper rune, userVk *cache.UserVk, w http.ResponseWriter, r *http.Request) {
	id, err := c.getId(w, r)
	if err != nil {
		return
	}

	if !c.userPostRatingStor.OperAllowed(oper) {
		c.logger.Println("Invalid oper in changeRating(), oper: ", string(oper))
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	post, err := c.postStor.GetOne(id)
	if err != nil || post == nil {
		c.logger.Println("Unexisted id: ", id)
		http.Error(w, "Id does not exist, id: "+id, http.StatusBadRequest)
		return
	}

	userPostRating := &models.UserPostRating{
		UserId: userVk.Info.Id,
		PostId: id,
		Oper:   oper,
	}
	if err := c.userPostRatingStor.SetUserOper(userPostRating); err != nil {
		c.logger.Printf("Error changing rating in changeRating(), Oper: %v, Error: %v", oper, err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	delta := c.userPostRatingStor.OperToDelta(oper)

	post.Rating += delta
	if err := c.postStor.Update(post); err != nil {
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

	if post.Title == "" {
		c.logger.Println("Error post create without title,\n Post: ", post)
		http.Error(w, "Creating post without title", http.StatusBadRequest)
		return
	}

	id, err := c.postStor.Create(post.Title, post.Content, post.Img)

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
	post, err := c.postStor.GetOne(id)
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
	tasks, err := c.postStor.GetAll(limit)

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

	if err := c.postStor.Delete(id); err != nil {
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
	if err := c.postStor.Update(newTask); err != nil {
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

func (c *PostController) tryGetSession(tokenDto *TokenDTO, w http.ResponseWriter) (*cache.UserVk, error) {
	if tokenDto.SessionToken == "" {
		c.logger.Println("Got no token in Increment()")
		http.Error(w, "Got no sessonToken", http.StatusUnauthorized)
		return nil, errors.New("no token")
	}

	userVk, err := c.sessionStor.GetSession(tokenDto.SessionToken)
	if err != nil || !userVk.Valid() {
		c.logger.Println("Expired token in Increment(), token: ", tokenDto.SessionToken)
		if err != nil {
			c.logger.Println("Error: ", err.Error())
		}
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return nil, errors.New("token expired")
	}

	return userVk, nil
}

func (c *PostController) getSearchParams(w http.ResponseWriter, r *http.Request) (size int, page int, search string, err error) {
	query := r.URL.Query()
	pageQuery, ok := query["page"]
	if !ok || len(pageQuery) < 1 {
		c.logger.Println("Has no page in query for Search()")
		http.Error(w, "Has no page in query for Search()", http.StatusBadRequest)
		return 0, 0, "", errors.New("no page")
	}
	sizeQuery, ok := query["size"]
	if !ok || len(sizeQuery) < 1 {
		c.logger.Println("Has no size in query for Search()")
		http.Error(w, "Has no size in query for Search()", http.StatusBadRequest)
		return 0, 0, "", errors.New("no size")
	}
	searchQuery := query["search"]
	page, err = strconv.Atoi(pageQuery[0])
	if err != nil {
		c.logger.Println("Page in Search() is not integer")
		http.Error(w, "Page is not integer", http.StatusBadRequest)
		return 0, 0, "", errors.New("page is not int")
	}
	if page < 1 {
		c.logger.Println("Page in Search() is less than 1")
		http.Error(w, "Page is less than 1", http.StatusBadRequest)
		return 0, 0, "", errors.New("page is less than 1")
	}

	size, err = strconv.Atoi(sizeQuery[0])
	if err != nil {
		c.logger.Println("Size in Search() is not integer")
		http.Error(w, "Size is not integer", http.StatusBadRequest)
		return 0, 0, "", errors.New("size is not int")
	}
	if size < 1 {
		c.logger.Println("Size in Search() is less than 1")
		http.Error(w, "Size is less than 1", http.StatusBadRequest)
		return 0, 0, "", errors.New("size is less than 1")
	}

	search = ""
	if len(searchQuery) > 0 {
		search = strings.TrimSpace(searchQuery[0])
	}

	return size, page, search, nil
}
