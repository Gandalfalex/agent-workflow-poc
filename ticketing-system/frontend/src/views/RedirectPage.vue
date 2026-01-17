<script setup lang="ts">
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { useAdminStore } from "@/stores/admin";

const router = useRouter();
const adminStore = useAdminStore();

onMounted(async () => {
  if (adminStore.projects.length === 0) {
    await adminStore.loadProjects();
  }
  const project = adminStore.projects[0];
  if (project) {
    await router.replace(`/projects/${project.id}/board`);
  }
});
</script>

<template>
  <div class="rounded-3xl border border-border bg-card/80 p-6 text-sm text-muted-foreground">
    Loading workspace...
  </div>
</template>
