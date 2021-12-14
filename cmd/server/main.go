package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-session/session"
	"github.com/mjpitz/oauth2-oidc-key-exchange/cmd/server/web"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

func main() {
	jwtKey := "badadmin"
	access := generates.NewJWTAccessGenerate([]byte(jwtKey), jwt.SigningMethodHS512)

	clientStore := store.NewClientStore()
	_ = clientStore.Set("testClient", &models.Client{
		ID:     "testClient",
		Secret: "testSecret",
		Domain: "http://localhost",
	})

	manager := manage.NewDefaultManager()
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapAccessGenerate(access)

	manager.SetAuthorizeCodeExp(10 * time.Minute)

	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    2 * time.Hour,
		RefreshTokenExp:   14 * 24 * time.Hour,
		IsGenerateRefresh: true,
	})

	srv := server.NewDefaultServer(manager)

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
			AuthURL:     "http://localhost:9096/consent",
			TokenURL:    "http://localhost:9096/oauth/token",
			JWKSURL:     "http://localhost:9096/oauth/keys",
			UserInfoURL: "http://localhost:9096/oauth/userinfo",
			Algorithms:  []string{},
		})
	})

	http.HandleFunc("/oauth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Form == nil {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "bad form data", http.StatusBadRequest)
				return
			}
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "test" && password == "test" {
			store, err := session.Start(r.Context(), w, r)
			if err != nil {
				return
			}

			store.Set("user", "test")
			store.Save()

			http.Redirect(w, r, "/consent", http.StatusFound)
			return
		}
	})

	http.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
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

	http.Handle("/", web.Handler())

	log.Println("starting server on :9096")
	log.Fatal(http.ListenAndServe(":9096", nil))
}

type providerJSON struct {
	Issuer      string   `json:"issuer"`
	AuthURL     string   `json:"authorization_endpoint"`
	TokenURL    string   `json:"token_endpoint"`
	JWKSURL     string   `json:"jwks_uri"`
	UserInfoURL string   `json:"userinfo_endpoint"`
	Algorithms  []string `json:"id_token_signing_alg_values_supported"`
}
