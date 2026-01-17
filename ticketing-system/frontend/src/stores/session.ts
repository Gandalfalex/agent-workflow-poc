import { defineStore } from "pinia";
import { getMe, login, logout, type AuthUser } from "@/lib/api";

type AuthStatus = "checking" | "authenticated" | "unauthenticated";

export const useSessionStore = defineStore("session", {
  state: () => ({
    status: "checking" as AuthStatus,
    user: null as AuthUser | null,
  }),
  actions: {
    reset() {
      this.status = "unauthenticated";
      this.user = null;
    },
    async checkSession() {
      this.status = "checking";
      try {
        const me = await getMe();
        this.user = me;
        this.status = "authenticated";
        return me;
      } catch (err) {
        this.user = null;
        this.status = "unauthenticated";
        throw err;
      }
    },
    async login(identifier: string, password: string) {
      const response = await login(identifier, password);
      this.user = response.user;
      this.status = "authenticated";
      return response.user;
    },
    async logout() {
      try {
        await logout();
      } finally {
        this.user = null;
        this.status = "unauthenticated";
      }
    },
  },
});
