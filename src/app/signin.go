package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	gogithub "github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:9785/auth/github/callback",
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	Scopes:       []string{"repo"},
	Endpoint:     github.Endpoint,
}

func generateStateOauthCookie(rw http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(rw, cookie)
	return state
}

func getGithubRepos(code string) ([]*gogithub.Repository, error) {
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(token)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}
	tc := oauth2.NewClient(ctx, ts)
	client := gogithub.NewClient(tc)
	repos, _, err := client.Repositories.List(ctx, "", nil)

	return repos, nil
}

func getGithubUserInfo(code string) (*gogithub.User, error) {
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(token)
	if err != nil {
		return &gogithub.User{}, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}
	tc := oauth2.NewClient(ctx, ts)
	client := gogithub.NewClient(tc)
	user, _, err := client.Users.Get(context.Background(), "")
	return user, nil
}

func (a *AppHandler) GithubLoginHandler(rw http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(rw)
	url := githubOauthConfig.AuthCodeURL(state)
	http.Redirect(rw, r, url, http.StatusTemporaryRedirect)
}

func (a *AppHandler) GithubAuthCallback(rw http.ResponseWriter, r *http.Request) {
	oauthstate, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthstate.Value {
		log.Printf("invalid github oauth state cookie:%s state:%s\n", oauthstate.Value, r.FormValue("state"))
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := getGithubUserInfo(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	// Store ID info into Session cookie
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id"] = userInfo.ID
	session.Values["login"] = userInfo.Login
	err = session.Save(r, rw)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
}
