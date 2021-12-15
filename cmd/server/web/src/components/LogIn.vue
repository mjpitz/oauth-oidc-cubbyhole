<template>
  <div class="container">
    <div class="row">
      <div class="col-sm-10 col-sm-offset-1">
        <div v-if="this.errors.length > 0">
          <p v-for="err in this.errors" :key="err">{{ err }}</p>
        </div>
        <form v-else @submit.prevent="submitForm">
          <label for="username">Username</label>
          <input type="text" id="Username" placeholder="Username" v-model="username"/>
          <label for="password">Password</label>
          <input type="password" id="password" placeholder="Password" v-model="password"/>
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
      username: "",
      password: "",
      errors: [],
    };
  },

  mounted() {
    let [ , err ] = cubbyhole(window?.location?.hash?.substr(1));
    if (err) {
      this.errors = ([ err.toString() ]);
    }
  },

  methods: {
    async submitForm() {
      const resp = await fetch("/oauth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: this.username,
          password: this.password,
        }),
      });

      await resp.text()

      if (resp.status >= 400) {
        this.errors = [ resp.statusText ];
        return;
      }

      localStorage.setItem("user", "username");

      return this.$router.push("/oauth/authorize");
    }
  }
}
</script>

<style scoped>
</style>
