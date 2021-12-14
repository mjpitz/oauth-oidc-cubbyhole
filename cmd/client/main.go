package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/oauth2.v3/utils/uuid"
)

type rt struct {
	endpoint oauth2.Endpoint
	key      string
}

func (r *rt) RoundTrip(request *http.Request) (*http.Response, error) {
	// when we go to the auth url, add an encryption key...
	// we'll need to figure out how to instruct others
	if strings.HasPrefix(request.URL.String(), r.endpoint.AuthURL) {
		request.URL.Fragment = r.key
	}

	return http.DefaultTransport.RoundTrip(request)
}

var _ http.RoundTripper = &rt{}

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

		if token != nil {
			// store token somewhere
			log.Println(token)
			// client := oauth2Config.Client(ctx, token)
		}

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
