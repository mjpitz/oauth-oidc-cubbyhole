<template>
  <div class="container">
    <div class="row">
      <div class="col-sm-10 col-sm-offset-1">
        <div v-if="this.errors.length > 0">
          <p v-for="err in this.errors" :key="err">{{ err }}</p>
        </div>
        <form v-else method="post" action="/oauth/login">
          <label for="username">Username</label>
          <input type="text" id="Username" placeholder="Username"/>
          <label for="password">Password</label>
          <input type="password" id="password" placeholder="Password"/>
          <input type="submit" value="Log In" />
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import cubbyhole from "./cubbyhole";

export default {
  data() {
    return {
      errors: [],
    }
  },

  mounted() {
    let [ secret, err ] = cubbyhole(window?.location?.hash?.substr(1))
    if (err) {
      this.error = this.errors.concat([ err.toString() ])
      return
    }

    console.log(secret)
  },
}
</script>

<style scoped>
</style>
