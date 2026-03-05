import { defineStore } from "pinia";
import {
  getPortfolioStats,
  exportPortfolioStats,
  type PortfolioStats,
  type PortfolioTotals,
  type ProjectPortfolioEntry,
} from "@/lib/api";

export type { PortfolioStats, PortfolioTotals, ProjectPortfolioEntry };

export const usePortfolioStore = defineStore("portfolio", {
  state: () => ({
    stats: null as PortfolioStats | null,
    loading: false,
    error: "",
  }),
  getters: {
    totals: (state): PortfolioTotals | null => state.stats?.totals ?? null,
    projects: (state): ProjectPortfolioEntry[] => state.stats?.projects ?? [],
    atRiskCount: (state): number =>
      (state.stats?.projects ?? []).filter((p) => p.urgentOpen > 0).length,
  },
  actions: {
    async load(filter?: { ownerId?: string; groupId?: string }) {
      this.loading = true;
      this.error = "";
      try {
        this.stats = await getPortfolioStats(filter);
      } catch (err) {
        this.error = (err as Error).message || "Failed to load portfolio";
      } finally {
        this.loading = false;
      }
    },
    async exportCsv() {
      const file = await exportPortfolioStats();
      const url = URL.createObjectURL(file.blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = file.filename;
      a.click();
      URL.revokeObjectURL(url);
    },
  },
});
