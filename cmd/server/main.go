package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"github.com/mjpitz/oauth2-oidc-key-exchange/cmd/server/web"
)

func main() {
	jwtKey := "badadmin"
	access := generates.NewJWTAccessGenerate("", []byte(jwtKey), jwt.SigningMethodHS512)

	tokenStore, _ := store.NewMemoryTokenStore()

	clientStore := store.NewClientStore()
	_ = clientStore.Set("testClient", &models.Client{
		ID:     "testClient",
		Secret: "testSecret",
		Domain: "http://localhost:9090",
	})

	manager := manage.NewDefaultManager()
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)
	manager.MapAccessGenerate(access)

	manager.SetAuthorizeCodeExp(10 * time.Minute)
	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    2 * time.Hour,
		RefreshTokenExp:   14 * 24 * time.Hour,
		IsGenerateRefresh: true,
	})

	srv := server.NewDefaultServer(manager)

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		data, _ := json.Marshal(re)
		log.Println(string(data))
	})

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println(err)
		return nil
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		store, err := session.Start(r.Context(), w, r)
		if err != nil {
			return
		}

		uid, ok := store.Get("user")
		if !ok {
			return "", nil
		}

		return uid.(string), nil
	})

	http.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&providerJSON{
			Issuer:      "http://localhost:9096",
			AuthURL:     "http://localhost:9096/oauth/authorize",
			TokenURL:    "http://localhost:9096/oauth/token",
			UserInfoURL: "http://localhost:9096/oauth/userinfo",
		})
	})

	http.HandleFunc("/oauth/login", func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]string)
		err := json.NewDecoder(r.Body).Decode(&values)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		username := values["username"]
		password := values["password"]

		if username == "test" && password == "test" {
			store, err := session.Start(r.Context(), w, r)
			if err != nil {
				return
			}

			store.Set("user", "test")
			store.Save()
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	webHandle := web.Handler()

	http.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			webHandle.ServeHTTP(w, r)
			return
		}

		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, r *http.Request) {
		access := r.Header.Get("Authorization")
		access = strings.TrimPrefix(access, "Bearer ")
		access = strings.TrimPrefix(access, "bearer ")

		info, err := tokenStore.GetByAccess(r.Context(), access)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
		}

		project, cubbyhole := "", ""
		var buckets []string

		scopes := strings.Split(info.GetScope(), " ")
		for _, scope := range scopes {
			switch {
			case strings.HasPrefix(scope, "project:"):
				if project != "" {
					http.Error(w, "", http.StatusBadRequest)
					return
				}
				project = strings.TrimPrefix(scope, "project:")
			case strings.HasPrefix(scope, "bucket:"):
				buckets = append(buckets, strings.TrimPrefix(scope, "bucket:"))
			case strings.HasPrefix(scope, "cubbyhole:"):
				if cubbyhole != "" {
					http.Error(w, "", http.StatusBadRequest)
					return
				}
				cubbyhole = strings.TrimPrefix(scope, "cubbyhole:")
			}
		}

		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"sub":            info.GetUserID(),
			"profile":        "",
			"email":          "",
			"email_verified": false,
			"project":        project,
			"buckets":        buckets,
			"cubbyhole":      cubbyhole,
		})

		if err != nil  {
			http.Error(w, "", http.StatusInternalServerError)
		}
	})

	http.Handle("/", webHandle)

	log.Println("starting server on :9096")
	log.Fatal(http.ListenAndServe("0.0.0.0:9096", nil))
}

type providerJSON struct {
	Issuer      string   `json:"issuer"`
	AuthURL     string   `json:"authorization_endpoint"`
	TokenURL    string   `json:"token_endpoint"`
	JWKSURL     string   `json:"jwks_uri"`
	UserInfoURL string   `json:"userinfo_endpoint"`
	Algorithms  []string `json:"id_token_signing_alg_values_supported"`
}
