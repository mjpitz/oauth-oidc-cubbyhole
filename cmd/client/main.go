package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-session/session"
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

	svr := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		rawKey := uuid.Must(uuid.NewRandom()).Bytes()
		rawKey = append(rawKey, uuid.Must(uuid.NewRandom()).Bytes()...)
		cubbyholeKey := hex.EncodeToString(rawKey)

		store, _ := session.Start(r.Context(), w, r)
		store.Set("cubbyholeKey", cubbyholeKey)
		store.Save()

		state := uuid.Must(uuid.NewRandom()).String()
		redir := oauth2Config.AuthCodeURL(state)
		redir = redir + "#" + cubbyholeKey

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

		mp := map[string]interface{}{}
		err = userInfo.Claims(&mp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		store, _ := session.Start(r.Context(), w, r)
		key, _ := store.Get("cubbyholeKey")
		cubbyholeKey, _ := hex.DecodeString(key.(string))

		c, err := aes.NewCipher(cubbyholeKey)
		if err != nil {
			log.Fatal(err)
		}

		encryptionKey := make([]byte, 0)
		c.Decrypt(encryptionKey, []byte(mp["cubbyhole"].(string)))

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
