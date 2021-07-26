package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func getGithubUserInfo(code string) ([]byte, error) {
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(token)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}
	tc := oauth2.NewClient(ctx, ts)
	client := gogithub.NewClient(tc)
	_, resp, err := client.Users.Get(ctx, "")

	return ioutil.ReadAll(resp.Body)
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

	data, err := getGithubUserInfo(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	var user map[string]interface{}
	json.Unmarshal(data, &user)
	fmt.Println(user)

	// 	// Store ID info into Session cookie
	// 	var userInfo GoogleUserID
	// 	err = json.Unmarshal(data, &userInfo)
	// 	if err != nil {
	// 		log.Println(err.Error())
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	session, err := store.Get(r, "session")
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	session.Values["id"] = userInfo.ID
	// 	err = session.Save(r, w)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// }
}
