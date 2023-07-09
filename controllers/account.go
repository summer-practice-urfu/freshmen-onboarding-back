package controllers

import (
	"TaskService/models/cache"
	"TaskService/storages"
	"context"
	"encoding/json"
	"github.com/go-vk-api/vk"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	oauthVk "golang.org/x/oauth2/vk"
	"log"
	"net/http"
	"os"
)

type AccountController struct {
	logger      *log.Logger
	conf        *oauth2.Config
	sessionStor *storages.SessionStorage
}

var tokens map[string]oauth2.Token

func NewAccountController(logger *log.Logger, sessionStor *storages.SessionStorage) *AccountController {
	//tokenURL := os.Getenv("VK_TOKEN")

	conf := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes:       []string{},
		Endpoint:     oauthVk.Endpoint,
	}

	return &AccountController{
		logger:      logger,
		conf:        conf,
		sessionStor: sessionStor,
	}
}

func (c *AccountController) Register(basePath string, router *mux.Router) {
	router.HandleFunc(basePath+"/login", c.LogIn).Methods("POST")
	router.HandleFunc(basePath+"/logout", c.LogOut).Methods("POST")
	router.HandleFunc(basePath+"/url", c.GetUrl).Methods("GET")
	router.HandleFunc(basePath+"/verify", c.Verify)
}

type logDTO struct {
	SessionToken string `json:"sessionToken"`
}

func (c *AccountController) LogIn(w http.ResponseWriter, r *http.Request) {
	logDto, err := c.getLogDto(w, r)
	if err != nil {
		return
	}

	userVk, err := c.sessionStor.GetSession(logDto.SessionToken)
	if err != nil {
		c.logger.Println("Can't get session, Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(userVk.Info); err != nil {
		c.logger.Println("Error encoding userInfo, \nuserVk: ", userVk, "\nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func (c *AccountController) LogOut(w http.ResponseWriter, r *http.Request) {
	logDto, err := c.getLogDto(w, r)
	if err != nil {
		return
	}

	if err := c.sessionStor.DeleteSession(logDto.SessionToken); err != nil {
		c.logger.Println("Error deleting session, token: ", logDto.SessionToken, "\nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func (c *AccountController) Verify(w http.ResponseWriter, r *http.Request) {
	c.logger.Println("Verify() url: ", r.URL)
	queryCode := r.URL.Query()["code"]
	if len(queryCode) < 1 {
		c.logger.Println("Invalid code param: ", queryCode)
		http.Error(w, "Invalid code param", http.StatusBadRequest)
		return
	}
	code := queryCode[0]
	ctx := context.Background()
	token, err := c.conf.Exchange(ctx, code)
	if err != nil {
		c.logger.Println("Error exchanging token, Code: ", code, "Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	c.logger.Println("Token: ", token)
	client, err := vk.NewClientWithOptions(vk.WithToken(token.AccessToken))
	if err != nil {
		c.logger.Println("Error creating vk client, Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	user, err := getCurrentUser(client)
	if err != nil {
		c.logger.Println("Error")
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	c.logger.Println("User: ", user)

	/*http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "sessionToken",
		Expires: time.Now().Add(5 * time.Minute),
	})*/

	userVk := &cache.UserVk{
		Token: token,
		Info: &cache.UserInfo{
			Id:         user.ID,
			Name:       user.Name,
			SecondName: user.SecondName,
			Photo:      user.Photo,
		},
	}
	sessionToken, err := c.sessionStor.CreateSession(userVk)
	if err != nil {
		c.logger.Println("Error creating session, Error: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:8080/?sessionToken="+sessionToken, http.StatusSeeOther)
}

type User struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	SecondName string `json:"SecondName"`
	Photo      string `json:"photo_400_orig"`
}

func (c *AccountController) GetUrl(w http.ResponseWriter, r *http.Request) {
	url := c.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.logger.Println("Generated url in Authenricate(), \nURL: ", url)
	resp := struct {
		url string
	}{url: url}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.logger.Println("Error encoding url in Authenticate(), \nError: ", err.Error())
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
}

func getCurrentUser(api *vk.Client) (*User, error) {
	var users []User

	if err := api.CallMethod("users.get", vk.RequestParams{
		"fields": "photo_400_orig",
	}, &users); err != nil {
		return nil, err
	}

	return &users[0], nil
}

func (c *AccountController) getLogDto(w http.ResponseWriter, r *http.Request) (*logDTO, error) {
	var logDto *logDTO
	if err := json.NewDecoder(r.Body).Decode(logDto); err != nil {
		c.logger.Println("Error decoding logDto in LogIn(), Error: ", err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return nil, err
	}
	return logDto, nil
}
