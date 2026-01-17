import { createRouter, createWebHistory } from "vue-router";
import BoardPage from "@/views/BoardPage.vue";
import SettingsPage from "@/views/SettingsPage.vue";
import RedirectPage from "@/views/RedirectPage.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: RedirectPage },
    {
      path: "/projects/:projectId/board",
      name: "board",
      component: BoardPage,
      props: true,
    },
    {
      path: "/projects/:projectId/settings",
      name: "settings",
      component: SettingsPage,
      props: true,
    },
  ],
});

export default router;
