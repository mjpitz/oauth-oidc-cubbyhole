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

