# OAuth / OIDC Cubbyhole

Mostly a demo of the work I've been designing at Storj. This demonstrates how to create a client-only secret that can
be used to create a secure cubbyhole that will allow a user to share secret data with a client application, but not the
identity provider (such as an encryption key).

A few items to note:

- The identity provider should use a single page application driven by react / vue
  - Otherwise, rendering the consent screen can cause a redirect which will occur before we can cache the secret.
  - See `server/web/src/components/Consent.vue` and `LogIn.vue` for a reference.
- When the consent page loads, it should cache the encryption key in local storage BEFORE checking user auth and 
  determining if the user needs to sign in.
- Current example is done using go, but based on a cursory search most libraries should be able to support this.
  - There might be a challenge here with some more opinionated frameworks, but special casing those shouldn't be too bad.

## Running the code

1. Run the server
   ```
   npm install -g @vue/cli
   go generate ./...
   go run ./cmd/server/main.go
   ```
   
2. Run the client
   ```
   go run ./cmd/client/main.go
   ```
   
3. Open http://localhost:8080/login in your browser. You should be redirected to the consent screen where we can check
   for a user session before redirecting them to the login page. The redirect from consent to log in is currently not in
   place. You should be able to inspect localStorage to see the associated cubbyhole key cached.

4. After logging in, the user should come back to the consent screen. In the demo case, the consent form prompts the 
   user for a passphrase. This passphrase would be encrypted using the cubbyhole key before being sent to the server as
   part of the form data.
     - For Storj, we would encrypt the derived encryption key, and not the bucket passphrase itself. The passphrase and
       cubbyhole key would never be passed along to the server, only the encrypted payload.

5. After being authorized access, the calling app should be able to obtain the cubbyhole value from the user profile
   (shown in cmd/client/main.go).
