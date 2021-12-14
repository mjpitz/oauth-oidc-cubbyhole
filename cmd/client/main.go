package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/oauth2.v3/utils/uuid"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:9096")
	if err != nil {
		log.Fatal(err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     "testClient",
		ClientSecret: "testSecret",
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID},
	}

	userKey := "testEncryptionKey"
	c, err := aes.NewCipher([]byte(userKey))
	if err != nil {
		log.Fatal(err)
	}

	svr := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.Must(uuid.NewRandom())

		redir := oauth2Config.AuthCodeURL(uuid.String())
		redir = redir + "#" + userKey

		http.Redirect(w, r, redir, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		token, err := oauth2Config.Exchange(ctx, q.Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2Config.TokenSource(ctx, token))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mp := map[string]string{}
		err = userInfo.Claims(mp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		encryptionKey := make([]byte, 0)
		c.Decrypt(encryptionKey, []byte(mp["cubbyhole"]))

		// cache encryption key, userInfo, and token for later use

		defer svr.Shutdown(ctx)

		http.ServeContent(w, r, "", time.Now(), bytes.NewReader(nil))
	})

	var group errgroup.Group

	group.Go(func() error {
		return svr.ListenAndServe()
	})

	log.Print("Open http://localhost:8080/login in your browser")
	group.Wait()
}
