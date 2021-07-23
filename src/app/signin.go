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
	// resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to get UserInfo %s\n", err.Error())
	// }

	return repos, nil
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

	data, err := getGithubRepos(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprint(rw, data)
}

/*
GOOGLE OAUTH

var googleOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:9785/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_SECRET_KEY"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
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

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func getGoogleUserInfo(code string) ([]byte, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}

	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to get UserInfo %s\n", err.Error())
	}

	return ioutil.ReadAll(resp.Body)
}

func (a *AppHandler) GoogleLoginHandler(rw http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(rw)
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(rw, r, url, http.StatusTemporaryRedirect)
}

func (a *AppHandler) GoogleAuthCallback(rw http.ResponseWriter, r *http.Request) {
	oauthstate, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthstate.Value {
		log.Printf("invalid google oauth state cookie:%s state:%s\n", oauthstate.Value, r.FormValue("state"))
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getGoogleUserInfo(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprint(rw, string(data))
}
*/
