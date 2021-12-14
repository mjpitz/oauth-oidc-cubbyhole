import {createApp} from 'vue'
import {createRouter, createWebHistory} from 'vue-router'

import Consent from "./components/Consent.vue"
import LogIn from "./components/LogIn.vue"
import Main from "./components/Main.vue"

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes: [
        {path: "/login", name: "LogIn", component: LogIn},
        {path: "/consent", name: "Consent", component: Consent},
    ],
})

createApp(Main)
    .use(router)
    .mount('#app')
