package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
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
		RedirectURL:  "http://localhost:9090/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{
			oidc.ScopeOpenID,
			"object:read",
			"object:write",
			"object:delete",
		},
	}

	svr := &http.Server{
		Addr:    ":9090",
		Handler: http.DefaultServeMux,
	}

	rawKey := uuid.Must(uuid.NewRandom()).Bytes()
	rawKey = append(rawKey, uuid.Must(uuid.NewRandom()).Bytes()...)
	cubbyholeKey := hex.EncodeToString(rawKey)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := uuid.Must(uuid.NewRandom()).String()
		redir := oauth2Config.AuthCodeURL(state)
		redir = redir + "#" + cubbyholeKey

		http.Redirect(w, r, redir, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			go func() {
				time.Sleep(5 * time.Second)
				svr.Shutdown(ctx)
			}()
		}()

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

		key, _ := hex.DecodeString(cubbyholeKey)
		k := sha256.Sum256(key)

		c, err := aes.NewCipher(k[:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cubbyhole, _ := hex.DecodeString(mp["cubbyhole"].(string))
		passphrase := make([]byte, len(cubbyhole))

		cipher.NewCBCDecrypter(c, make([]byte, c.BlockSize())).CryptBlocks(passphrase, cubbyhole)

		log.Println("cubbyhole value: ", strings.TrimSpace(string(passphrase)))

		// cache cubbyhole value, userInfo, and token for later use

		http.ServeContent(w, r, "", time.Now(), bytes.NewReader(nil))
	})

	var group errgroup.Group

	group.Go(func() error {
		return svr.ListenAndServe()
	})

	log.Print("Open http://localhost:9090/login in your browser")
	group.Wait()
}
