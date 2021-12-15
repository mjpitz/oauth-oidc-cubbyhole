<template>
  <div class="container">
    <div class="row">
      <div class="col-sm-10 col-sm-offset-1">
        <div v-if="this.errors.length > 0">
          <p v-for="err in this.errors" :key="err">{{ err }}</p>
        </div>

        <div v-else>
          <form>
            <label for="project">Project</label>
            <input type="text" id="project" placeholder="Project" v-model="project"/>

            <label for="bucket">Bucket</label>
            <input type="text" id="bucket" placeholder="Bucket" v-model="bucket"/>

            <label for="passphrase">Passphrase</label>
            <input type="password" id="passphrase" placeholder="Passphrase" v-model="passphrase"/>
          </form>

          <form method="post">
            <input type="hidden" name="scope" v-model="fullScope"/>
            <input type="hidden" name="redirect_uri" v-model="appInfo.redirectURI"/>
            <input type="hidden" name="client_id" v-model="appInfo.clientID"/>
            <input type="hidden" name="state" v-model="appInfo.state"/>
            <input type="hidden" name="response_type" v-model="appInfo.responseType"/>

            <input type="submit" value="Authorize"/>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import CryptoJS from "crypto-js";
import cubbyhole from "./cubbyhole";

export default {
  data() {
    return {
      appInfo: {},
      project: "",
      bucket: "",
      passphrase: "",
      cubbyholeKey: "",
      errors: [],
    }
  },

  mounted() {
    let [secret, err] = cubbyhole(window?.location?.hash?.substr(1))
    if (err) {
      this.errors = [err.toString()]
      return
    }

    let url = new URL(window?.location?.href)
    if (url.search?.length > 1) {
      this.appInfo = {
        redirectURI: url.searchParams.get("redirect_uri"),
        clientID: url.searchParams.get("client_id"),
        state: url.searchParams.get("state"),
        responseType: url.searchParams.get("response_type"),
        scope: url.searchParams.get("scope"),
      }
    } else {
      let info = localStorage.getItem("appInfo")
      localStorage.removeItem("appInfo")

      try {
        this.appInfo = JSON.parse(info)
      } catch (err) {
        this.errors = ["app info missing"]
      }
    }

    let user = localStorage.getItem("user")
    if (!user) {
      // cache this for when we come back
      localStorage.setItem("appInfo", JSON.stringify(this.appInfo))
      this.$router.push("/login")
      return
    }

    localStorage.removeItem("user")
    this.cubbyholeKey = secret
  },

  computed: {
    fullScope() {
      return [
        this.appInfo.scope,
        `project:${this.project}`,
        `bucket:${this.bucket}`,
        `cubbyhole:${this.cubbyhole}`
      ].join(' ')
    },

    cubbyhole() {
      let key = Buffer.from(this.cubbyholeKey, "hex").toString()

      return CryptoJS.AES.encrypt(this.passphrase, key, {})
    }
  },
}
</script>

<style scoped>
</style>
