package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/JungBin-Eom/DevEnvMaker-Envi/data"
	"github.com/JungBin-Eom/DevEnvMaker-Envi/model"
	"github.com/bndr/gojenkins"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSEION_KEY")))
var rd *render.Render = render.New()

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

type Success struct {
	Success bool `json:"success"`
}

type CreateSuccess struct {
	Success bool `json:"success"`
	Count   int  `json:"count"`
}

type BuildSuccess struct {
	Success bool   `json:"success"`
	Id      int    `json:"id"`
	Job     string `json:"job"`
}

type Status struct {
	Status  bool `json:"status"`
	Running bool `json:"running"`
}

var getSessionID = func(r *http.Request) int {
	session, err := store.Get(r, "session")
	if err != nil {
		return 0
	}

	val := session.Values["id"]
	if val == nil {
		return 0
	}
	return int(val.(int64))
}

var getSessionName = func(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}

	val := session.Values["login"]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (a *AppHandler) IndexHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/index.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) SignInHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/signin.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) SignUpHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/signup.html", http.StatusTemporaryRedirect)
}

// USER APIs
func (a *AppHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.Login
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
		// http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	login, sessionId, githubName := a.db.AuthUser(user)
	if login == true {
		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		session.Values["id"] = int64(sessionId)
		session.Values["login"] = githubName
		err = session.Save(r, rw)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		rd.JSON(rw, http.StatusOK, Success{login})
	} else {
		rd.JSON(rw, http.StatusBadRequest, Success{login})
	}
}

type Duplicated struct {
	Duplicated bool `json:"duplicated"`
}

func (a *AppHandler) DupCheckHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := vars["id"]
	dup := a.db.CheckIdDup(id) // T: Dup, F: Not Dup
	rd.JSON(rw, http.StatusOK, Duplicated{dup})
}

func (a *AppHandler) UserInfoHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	user, err := a.db.UserInfo(sessionId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rd.JSON(rw, http.StatusOK, user)
}

func (a *AppHandler) UserRegisterHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.User
	sessionId := getSessionID(r)
	githubName := getSessionName(r)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
		// http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	user.GithubName = githubName
	err = a.db.RegisterUser(user, sessionId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rd.JSON(rw, http.StatusOK, Success{true})
}

func (a *AppHandler) RegisterTokenHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	githubName := getSessionName(r)
	type Token struct {
		Token string `json:"token"`
	}
	var token Token
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&token)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	err = a.db.RegisterToken(sessionId, token.Token)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/Ricky-Envi/Envi-ArgoCD/forks", nil)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "token "+token.Token)
	req.Header.Set("user-agent", githubName)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	res.Body.Close()
	rd.JSON(rw, http.StatusOK, Success{true})
}

// PROJECT APIs
func (a *AppHandler) GetProjectsHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	list := a.db.GetProjectList(sessionId)
	rd.JSON(rw, http.StatusOK, list)
}

func (a *AppHandler) CreateProjectHandler(rw http.ResponseWriter, r *http.Request) {
	var newprj data.Project
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	if user.GithubToken == "NULL" {
		rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newprj)
		if err != nil {
			rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
		}

		count, err := a.db.CreateProject(newprj, sessionId)
		if count != 0 {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, count})
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		pw := os.Getenv("JENKINS_PW")
		jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
		_, err = jenkins.Init(r.Context())

		if err != nil {
			panic("Something Went Wrong")
		}

		jenkins.CreateFolder(r.Context(), newprj.Name)

		// ArgoCD project create
		argocdPrjTemplate := `{
			"project": {
					"metadata": { 
							"name": "` + newprj.Name + `"
					},
					"spec": {
							"description": "` + newprj.Description + `",
							"destinations": [
									{
									"server": "https://kubernetes.default.svc",
									"namespace": "` + newprj.Name + `"
									}
							],
							"clusterResourceWhitelist": [
									{
											"group": "*",
											"kind": "*"
									}
							],
							"sourceRepos": ["*"]
					}
			},
			"upsert": true
		}`

		buff := bytes.NewBuffer([]byte(argocdPrjTemplate))
		argocdToken := os.Getenv("ARGOCD_TOKEN")
		req, err := http.NewRequest("POST", "https://3.35.25.64:31286/api/v1/projects", buff)
		if err != nil {
			rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
			return
		}
		req.Header.Set("Authorization", "Bearer "+argocdToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
			return
		}
		defer res.Body.Close()

		rd.JSON(rw, http.StatusOK, CreateSuccess{true, 0})
	}
}

func (a *AppHandler) RemoveProjectHandler(rw http.ResponseWriter, r *http.Request) {
	var project data.Project
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	ok := a.db.RemoveProject(project, sessionId)
	if ok {
		// remove jenkins folder
		pw := os.Getenv("JENKINS_PW")
		jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
		_, err = jenkins.Init(r.Context())
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}

		deleteFolder, err := jenkins.DeleteJob(r.Context(), project.Name)
		if err != nil || deleteFolder == false {
			rd.JSON(rw, http.StatusOK, Success{false})
		} else {
			// remove argocd project
			argocdToken := os.Getenv("ARGOCD_TOKEN")
			req, err := http.NewRequest("DELETE", "https://3.35.25.64:31286/api/v1/projects/"+project.Name, nil)
			if err != nil {
				rd.JSON(rw, http.StatusOK, Success{false})
				return
			}
			req.Header.Set("Authorization", "Bearer "+argocdToken)

			_, err = http.DefaultClient.Do(req)
			if err != nil {
				rd.JSON(rw, http.StatusOK, Success{false})
				return
			}
			appList := a.db.GetAppList(sessionId)
			for _, app := range appList {
				if app.Project == project.Name {
					req, err := http.NewRequest("DELETE", "https://api.github.com/repos/"+user.GithubName+"/"+app.Name, nil)
					if err != nil {
						rd.JSON(rw, http.StatusInternalServerError, Success{false})
						return
					}
					req.Header.Set("content-type", "application/json")
					req.Header.Set("authorization", "token "+user.GithubToken)
					req.Header.Set("user-agent", user.GithubName)

					res, err := http.DefaultClient.Do(req)
					if err != nil {
						fmt.Println("github api error")
						rd.JSON(rw, http.StatusBadRequest, Success{false})
						return
					}
					_, err = ioutil.ReadAll(res.Body)
					if err != nil {
						http.Error(rw, "Unable to read body", http.StatusBadRequest)
					}
					res.Body.Close()
				}
			}
			rd.JSON(rw, http.StatusOK, Success{true})
		}
	} else {
		rd.JSON(rw, http.StatusOK, Success{false})
	}
	return
}

func (a *AppHandler) GetProjectDetailHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, _ := vars["name"]
	sessionId := getSessionID(r)
	project, err := a.db.GetProject(name, sessionId)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	rd.JSON(rw, http.StatusOK, project)
}

// APPLICATION APIs
func (a *AppHandler) GetAppsHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	list := a.db.GetAppList(sessionId)
	rd.JSON(rw, http.StatusOK, list)
}

func (a *AppHandler) CreateAppHandler(rw http.ResponseWriter, r *http.Request) {
	var newapp data.Application
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	if user.GithubToken == "NULL" {
		rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newapp)
		if err != nil {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
		}

		count, err := a.db.CreateApp(newapp, sessionId)
		if count != 0 {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, count})
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		pbytes, _ := json.Marshal(newapp)
		buff := bytes.NewBuffer(pbytes)
		req, err := http.NewRequest("POST", "https://api.github.com/user/repos", buff)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}

		// Add app info to values.yaml of App of Apps
		req, err = http.NewRequest("GET", "https://api.github.com/repos/"+user.GithubName+"/Envi-ArgoCD/contents/AppOfApps/values.yaml", nil)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}
		getValuesRes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
		}
		res.Body.Close()

		var values data.ValuesResponse
		json.Unmarshal(getValuesRes, &values)

		valDec, _ := base64.StdEncoding.DecodeString(values.Content)
		newAppVal := `
  - appname: ` + newapp.Name + `
    projectname: ` + newapp.Project + `
    githuburl: https://github.com/` + user.GithubName + `/` + newapp.Name

		valDec = []byte(string(valDec) + newAppVal)
		valEnc := base64.URLEncoding.EncodeToString(valDec)
		values.Content = valEnc
		values.Message = "Add " + newapp.Name + " Info to ArgoCD values.yaml"

		vbytes, _ := json.Marshal(values)
		buff = bytes.NewBuffer(vbytes)
		req, err = http.NewRequest("PUT", "https://api.github.com/repos/"+user.GithubName+"/Envi-ArgoCD/contents/AppOfApps/values.yaml", buff)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}

		cmd := exec.Command("curl", "-LJO", "https://api.github.com/repos/Ricky-Envi/Envi-"+newapp.Runtime+"/tarball")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		cmd = exec.Command("/bin/sh", "-c", "tar -xvzf Ricky-Envi-*")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		cmd = exec.Command("/bin/sh", "-c", "rm -rf Ricky-Envi-*.tar.gz")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		cmd = exec.Command("mkdir", "temp")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		cmd = exec.Command("/bin/sh", "-c", "cp -rf Ricky-Envi-*/* temp")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		cmd = exec.Command("/bin/sh", "-c", "rm -rf Ricky-Envi-*")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp"
		cmd.Run()

		// git code download
		cmd = exec.Command("ls")
		cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp/temp"
		contents, _ := cmd.Output()
		contentList := strings.Split(string(contents), "\n")

		// var files map[string]string
		for _, cont := range contentList {
			cmd = exec.Command("file", cont)
			cmd.Dir = "/Users/ricky/Desktop/2021/Envi-Temp/temp"
			output, _ := cmd.Output()
			fileInfo := strings.Split(string(output), ": ")
			fmt.Println(fileInfo)
			if len(fileInfo) == 2 {
				fileType := fileInfo[1]
				if fileType != "directory\n" {
					fileName := strings.Split(fileInfo[0], " ")[0]
					fmt.Println(fileName)
					if fileName == "Jenkinsfile" {
						fileByte, err := ioutil.ReadFile("/Users/ricky/Desktop/2021/Envi-Temp/temp/" + fileName)
						if err != nil {
							fmt.Println(err.Error())
						}
						fileContent := string(fileByte)
						fileContent = strings.Replace(fileContent, "<GITHUB_URL>", "https://github.com/"+user.GithubName+"/"+newapp.Name, -1)
						fileContent = strings.Replace(fileContent, "<APP_NAME>", newapp.Name, -1)
						fmt.Println(fileContent)
						err = ioutil.WriteFile("/Users/ricky/Desktop/2021/Envi-Temp/temp/"+fileName, []byte(fileContent), 0)
						if err != nil {
							fmt.Println(err.Error())
						}
					}
				} else {
					fmt.Println("it is directory")
				}

			}
		}

		// Jenkins Job Create
		pw := os.Getenv("JENKINS_PW")
		jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
		_, err = jenkins.Init(r.Context())

		if err != nil {
			panic("Something Went Wrong")
		}
		config := `
		<flow-definition plugin="workflow-job@2.41">
			<actions>
				<org.jenkinsci.plugins.workflow.multibranch.JobPropertyTrackerAction plugin="workflow-multibranch@2.24">
					<jobPropertyDescriptors>
						<string>org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty</string>
					</jobPropertyDescriptors>
				</org.jenkinsci.plugins.workflow.multibranch.JobPropertyTrackerAction>
				<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobAction plugin="pipeline-model-definition@1.9.1" />
				<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction plugin="pipeline-model-definition@1.9.1">
					<jobProperties />
					<triggers />
					<parameters />
					<options />
				</org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction>
			</actions>
			<description />
			<keepDependencies>false</keepDependencies>
			<properties>
				<org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
					<triggers>
						<com.cloudbees.jenkins.GitHubPushTrigger plugin="github@1.34.0">
							<spec />
						</com.cloudbees.jenkins.GitHubPushTrigger>
					</triggers>
				</org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
			</properties>
			<definition class="org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition" plugin="workflow-cps@2.93">
				<scm class="hudson.plugins.git.GitSCM" plugin="git@4.8.1">
					<configVersion>2</configVersion>
					<userRemoteConfigs>
						<hudson.plugins.git.UserRemoteConfig>
							<url>https://github.com/` + user.GithubName + `/` + newapp.Name + `</url>
						</hudson.plugins.git.UserRemoteConfig>
					</userRemoteConfigs>
					<branches>
						<hudson.plugins.git.BranchSpec>
							<name>*/main</name>
						</hudson.plugins.git.BranchSpec>
					</branches>
					<doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
					<submoduleCfg class="empty-list" />
					<extensions />
				</scm>
				<scriptPath>Jenkinsfile</scriptPath>
				<lightweight>true</lightweight>
			</definition>
			<triggers />
			<disabled>false</disabled>
		</flow-definition>
		`

		_, err = jenkins.CreateJobInFolder(r.Context(), config, newapp.Name, newapp.Project)
		if err != nil {
			http.Redirect(rw, r, "../html/404.html", http.StatusTemporaryRedirect)
		}
		rd.JSON(rw, http.StatusOK, CreateSuccess{true, 0})
	}
}

func (a *AppHandler) RemoveAppHandler(rw http.ResponseWriter, r *http.Request) {
	var app data.Application
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&app)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}

	apps := a.db.GetAppList(sessionId)
	for _, item := range apps {
		if item.Name == app.Name {
			app.Project = item.Project
			break
		}
	}

	ok := a.db.RemoveApp(app, sessionId)
	if ok {
		req, err := http.NewRequest("DELETE", "https://api.github.com/repos/"+user.GithubName+"/"+app.Name, nil)
		if err != nil {
			rd.JSON(rw, http.StatusInternalServerError, Success{false})
			return
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("github api error")
			rd.JSON(rw, http.StatusBadRequest, Success{false})
			return
		}
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
		}
		res.Body.Close()

		pw := os.Getenv("JENKINS_PW")
		jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
		_, err = jenkins.Init(r.Context())
		if err != nil {
			rd.JSON(rw, http.StatusOK, Success{false})
		}

		jobName := app.Project + "/job/" + app.Name
		deleteJob, err := jenkins.DeleteJob(r.Context(), jobName)
		if err != nil || deleteJob == false {
			rd.JSON(rw, http.StatusOK, Success{false})
		} else {
			rd.JSON(rw, http.StatusOK, Success{true})
		}
	} else {
		rd.JSON(rw, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) GetAppDetailHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project, _ := vars["projname"]
	name, _ := vars["appname"]
	sessionId := getSessionID(r)
	app, err := a.db.GetApp(project, name, sessionId)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	rd.JSON(rw, http.StatusOK, app)
}

func (a *AppHandler) BuildAppHandler(rw http.ResponseWriter, r *http.Request) {
	var app data.Application
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&app)
	if err != nil {
		http.Redirect(rw, r, "../html/404.html", http.StatusTemporaryRedirect)
	}
	sessionId := getSessionID(r)
	apps := a.db.GetAppList(sessionId)
	for _, item := range apps {
		if item.Name == app.Name {
			app.Project = item.Project
			break
		}
	}

	pw := os.Getenv("JENKINS_PW")
	jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
	_, err = jenkins.Init(r.Context())

	if err != nil {
		panic("Something Went Wrong")
	}
	jobName := app.Project + "/job/" + app.Name
	_, err = jenkins.BuildJob(r.Context(), jobName, nil)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	// count := 0
	before, _ := jenkins.GetAllBuildIds(r.Context(), jobName)
	after, _ := jenkins.GetAllBuildIds(r.Context(), jobName)
	for {
		after, _ = jenkins.GetAllBuildIds(r.Context(), jobName)
		if len(before) == 0 && len(after) != 0 {
			break
		} else if len(before) != len(after) {
			break
		}
	}
	rd.JSON(rw, http.StatusOK, BuildSuccess{true, int(after[0].Number), app.Project + "-" + app.Name})
}

func (a *AppHandler) GetBuildStatusHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := vars["id"]
	job, _ := vars["job"]

	path := strings.Split(job, "-")
	jobName := path[0] + "/job/" + path[1]
	pw := os.Getenv("JENKINS_PW")

	jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
	_, err := jenkins.Init(r.Context())

	if err != nil {
		panic("Something Went Wrong")
	}
	buildId, err := strconv.Atoi(id)
	if err != nil {
		http.Redirect(rw, r, "../html/404.html", http.StatusTemporaryRedirect)
	}
	build, err := jenkins.GetBuild(r.Context(), jobName, int64(buildId))
	if err != nil {
		fmt.Println("Can't get build")
		rd.JSON(rw, http.StatusInternalServerError, Status{false, false})
		return
	}

	if build.IsRunning(r.Context()) { // now running
		rd.JSON(rw, http.StatusOK, Status{true, true})
	} else if build.IsGood(r.Context()) { // build finish successfully
		rd.JSON(rw, http.StatusOK, Status{true, false})
	} else { // build error
		rd.JSON(rw, http.StatusOK, Status{false, false})
	}
}

func (a *AppHandler) DeployAppHandler(rw http.ResponseWriter, r *http.Request) {
	var app data.Application
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&app)
	if err != nil {
		http.Redirect(rw, r, "../html/404.html", http.StatusTemporaryRedirect)
	}
	sessionId := getSessionID(r)
	apps := a.db.GetAppList(sessionId)
	for _, item := range apps {
		if item.Name == app.Name {
			app.Project = item.Project
			break
		}
	}

	pw := os.Getenv("JENKINS_PW")
	jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
	_, err = jenkins.Init(r.Context())

	if err != nil {
		panic("Something Went Wrong")
	}
	jobName := app.Project + "/job/" + app.Name
	_, err = jenkins.BuildJob(r.Context(), jobName, nil)
	if err != nil {
		rd.JSON(rw, http.StatusInternalServerError, Success{false})
		return
	}
	// count := 0
	before, _ := jenkins.GetAllBuildIds(r.Context(), jobName)
	after, _ := jenkins.GetAllBuildIds(r.Context(), jobName)
	for {
		after, _ = jenkins.GetAllBuildIds(r.Context(), jobName)
		if len(before) == 0 && len(after) != 0 {
			break
		} else if len(before) != len(after) {
			break
		}
	}
	rd.JSON(rw, http.StatusOK, BuildSuccess{true, int(after[0].Number), app.Project + "-" + app.Name})
}

// GITHUB APIs
func (a *AppHandler) GetGitHubNameHandler(rw http.ResponseWriter, r *http.Request) {
	githubName := getSessionName(r)
	rd.Text(rw, http.StatusOK, githubName)
}

func CheckSignin(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if strings.Contains(r.URL.Path, "/sign") || strings.Contains(r.URL.Path, "/auth") {
		next(rw, r)
		return
	}

	// if user already signed in
	sessionID := getSessionID(r)
	if sessionID != 0 {
		next(rw, r)
		return
	}

	// redirect signin.html
	http.Redirect(rw, r, "/html/signin.html", http.StatusTemporaryRedirect)
}

func MakeHandler(filepath string) *AppHandler {
	r := mux.NewRouter()

	neg := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.HandlerFunc(CheckSignin), negroni.NewStatic(http.Dir("public")))
	neg.UseHandler(r)

	a := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(filepath),
	}

	r.HandleFunc("/", a.IndexHandler)
	r.HandleFunc("/signin", a.SignInHandler).Methods("GET")
	r.HandleFunc("/signin", a.LoginHandler).Methods("POST")
	r.HandleFunc("/signup", a.SignUpHandler)
	r.HandleFunc("/signup/idcheck/{id:[a-zA-Z0-9]+}", a.DupCheckHandler).Methods("GET")
	r.HandleFunc("/signup/register", a.UserRegisterHandler).Methods("POST")

	r.HandleFunc("/user", a.UserInfoHandler).Methods("GET")
	r.HandleFunc("/user/token", a.RegisterTokenHandler).Methods("POST")

	r.HandleFunc("/project", a.GetProjectsHandler).Methods("GET")
	r.HandleFunc("/project", a.RemoveProjectHandler).Methods("DELETE")
	r.HandleFunc("/project", a.CreateProjectHandler).Methods("POST")
	r.HandleFunc("/project/{name:[a-zA-Z0-9]+}", a.GetProjectDetailHandler).Methods("GET")

	r.HandleFunc("/app", a.GetAppsHandler).Methods("GET")
	r.HandleFunc("/app", a.CreateAppHandler).Methods("POST")
	r.HandleFunc("/app", a.RemoveAppHandler).Methods("DELETE")
	r.HandleFunc("/app/{projname:[a-zA-Z0-9]+}/{appname:[a-zA-Z0-9]+}", a.GetAppDetailHandler).Methods("GET")
	r.HandleFunc("/app/build", a.BuildAppHandler).Methods("POST")
	r.HandleFunc("/app/build/status/{job:[a-zA-Z0-9-]+}/{id:[0-9]+}", a.GetBuildStatusHandler).Methods("GET")
	r.HandleFunc("/app/deploy", a.DeployAppHandler).Methods("POST")
	// r.HandleFunc("/app/build/status/{job:[a-zA-Z0-9-]+}/{id:[0-9]+}", a.GetBuildStatusHandler).Methods("GET")

	// r.HandleFunc("/repos", a.Repository).Methods("GET")

	r.HandleFunc("/github/name", a.GetGitHubNameHandler).Methods("GET")

	r.HandleFunc("/auth/github/login", a.GithubLoginHandler)
	r.HandleFunc("/auth/github/callback", a.GithubAuthCallback)

	// Swagger Handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.yml"}
	sh := middleware.Redoc(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yml", http.FileServer(http.Dir("./")))

	return a
}
