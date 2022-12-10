package gh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type AppToken struct {
	Token     string
	ExpiresAt time.Time
}
type Installations []struct {
	ID      int `json:"id"`
	Account struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		Type      string `json:"type"`
		SiteAdmin bool   `json:"site_admin"`
	} `json:"account"`
	RepositorySelection string `json:"repository_selection"`
	AccessTokensURL     string `json:"access_tokens_url"`
	RepositoriesURL     string `json:"repositories_url"`
	AppID               int    `json:"app_id"`
	AppSlug             string `json:"app_slug"`
	TargetID            int    `json:"target_id"`
	TargetType          string `json:"target_type"`
	Permissions         struct {
		Actions        string `json:"actions"`
		Metadata       string `json:"metadata"`
		Administration string `json:"administration"`
	} `json:"permissions"`
}

type InstallationToken struct {
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permissions struct {
		Actions        string `json:"actions"`
		Administration string `json:"administration"`
		Metadata       string `json:"metadata"`
	} `json:"permissions"`
	RepositorySelection string `json:"repository_selection"`
}

type RunnerToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WebHookEvent struct {
	Action      string `json:"action"`
	WorkflowJob struct {
		Status      string      `json:"status"`
		Conclusion  interface{} `json:"conclusion"`
		StartedAt   time.Time   `json:"started_at"`
		CompletedAt interface{} `json:"completed_at"`
		Name        string      `json:"name"`
		Labels      []string    `json:"labels"`
	} `json:"workflow_job"`
	Repository struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	} `json:"repository"`
	Installation struct {
		ID     int    `json:"id"`
		NodeID string `json:"node_id"`
	} `json:"installation"`
}

// TODO: implement get/refresh?

const APP_ID = "269165"
const APP_URL = "https://api.github.com/app" // TODO: Validate token on startup
const REGISTER_RUNNER_URL = "https://api.github.com/repos/%s/%s/actions/runners/registration-token"

func MakeAppToken() AppToken {
	secretKey, err := ioutil.ReadFile("multi-arch-builder.2022-12-05.private-key.pem")
	if err != nil {
		panic(err)
	}
	expires_at := time.Now().Add(1 * time.Minute)
	jwt, err := generateJWT(secretKey, expires_at)
	if err != nil {
		panic(err)
	}
	return jwt
}

func _http(method string, url string, bearer_token string) ([]byte, error) {
	//fmt.Printf("%s to %s with %s\n", method, url, bearer_token)
	//request, err := http.NewRequest("POST", url, nil)
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer_token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	//fmt.Println("response Status:", response.Status)
	body, _ := ioutil.ReadAll(response.Body)

	return body, nil
}
func generateJWT(key []byte, expires_at time.Time) (AppToken, error) {
	apptoken := AppToken{}
	pkey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return apptoken, err
	}
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Add(-1 * time.Minute).Unix()
	claims["exp"] = expires_at.Unix()
	claims["iss"] = APP_ID

	//fmt.Printf("%#v\n", claims)
	tokenString, err := token.SignedString(pkey)
	if err != nil {
		return apptoken, err
	}

	return AppToken{Token: tokenString, ExpiresAt: expires_at}, nil
}

func GetInstallationToken(installation int, bearer_token string) (RunnerToken, error) {
	rToken := RunnerToken{}
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installation)
	reply, err := _http("POST", url, bearer_token)
	if err != nil {
		return rToken, err
	}
	instToken := InstallationToken{}
	//fmt.Printf("RepoSelection: %s\n", instToken.RepositorySelection)
	json.Unmarshal(reply, &instToken)
	url = fmt.Sprintf(REGISTER_RUNNER_URL, "davidventura", "tpom")
	reply, err = _http("POST", url, instToken.Token)
	if err != nil {
		return rToken, err
	}
	json.Unmarshal(reply, &rToken)
	return rToken, nil
}
