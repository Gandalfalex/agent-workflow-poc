import { defineStore } from "pinia";
import {
  addGroupMember,
  addProjectGroup,
  createGroup,
  createProject,
  deleteGroupMember,
  deleteProjectGroup,
  listGroupMembers,
  listGroups,
  listProjectGroups,
  listProjects,
  listUsers,
  updateProjectGroup,
  type Group,
  type GroupMember,
  type Project,
  type ProjectGroup,
  type ProjectRole,
  type UserSummary,
} from "@/lib/api";

type LoadingState = "idle" | "loading" | "error";

export const useAdminStore = defineStore("admin", {
  state: () => ({
    projects: [] as Project[],
    groups: [] as Group[],
    projectGroups: [] as ProjectGroup[],
    groupMembers: [] as GroupMember[],
    userResults: [] as UserSummary[],
    projectStatus: "idle" as LoadingState,
    groupStatus: "idle" as LoadingState,
    projectGroupStatus: "idle" as LoadingState,
    groupMemberStatus: "idle" as LoadingState,
    userSearchStatus: "idle" as LoadingState,
    projectError: "",
    groupError: "",
    projectGroupError: "",
    groupMemberError: "",
    userSearchError: "",
  }),
  actions: {
    reset() {
      this.projects = [];
      this.groups = [];
      this.projectGroups = [];
      this.groupMembers = [];
      this.userResults = [];
      this.projectStatus = "idle";
      this.groupStatus = "idle";
      this.projectGroupStatus = "idle";
      this.groupMemberStatus = "idle";
      this.userSearchStatus = "idle";
      this.projectError = "";
      this.groupError = "";
      this.projectGroupError = "";
      this.groupMemberError = "";
      this.userSearchError = "";
    },
    clearGroupMembers() {
      this.groupMembers = [];
    },
    clearUserResults() {
      this.userResults = [];
    },
    async loadProjects() {
      this.projectStatus = "loading";
      this.projectError = "";
      try {
        const list = await listProjects();
        this.projects = list.items;
        this.projectStatus = "idle";
      } catch (err) {
        this.projectStatus = "error";
        this.projectError = "Unable to load projects.";
      }
    },
    async loadGroups() {
      this.groupStatus = "loading";
      this.groupError = "";
      try {
        const list = await listGroups();
        this.groups = list.items;
        this.groupStatus = "idle";
      } catch (err) {
        this.groupStatus = "error";
        this.groupError = "Unable to load groups.";
      }
    },
    async loadProjectGroups(projectId: string) {
      this.projectGroupStatus = "loading";
      this.projectGroupError = "";
      try {
        const list = await listProjectGroups(projectId);
        this.projectGroups = list.items;
        this.projectGroupStatus = "idle";
      } catch (err) {
        this.projectGroupStatus = "error";
        this.projectGroupError = "Unable to load project groups.";
      }
    },
    async loadGroupMembers(groupId: string) {
      this.groupMemberStatus = "loading";
      this.groupMemberError = "";
      try {
        const list = await listGroupMembers(groupId);
        this.groupMembers = list.items;
        this.groupMemberStatus = "idle";
      } catch (err) {
        this.groupMemberStatus = "error";
        this.groupMemberError = "Unable to load group members.";
      }
    },
    async searchUsers(query: string) {
      this.userSearchStatus = "loading";
      this.userSearchError = "";
      try {
        const list = await listUsers(query);
        this.userResults = list.items;
        this.userSearchStatus = "idle";
      } catch (err) {
        this.userSearchStatus = "error";
        this.userSearchError = "Unable to search users.";
      }
    },
    async createProject(input: {
      key: string;
      name: string;
      description?: string;
    }) {
      this.projectStatus = "loading";
      this.projectError = "";
      try {
        const created = await createProject(input);
        this.projects = [...this.projects, created];
        this.projectStatus = "idle";
        return created;
      } catch (err) {
        this.projectStatus = "error";
        this.projectError = "Unable to create project.";
        throw err;
      }
    },
    async createGroup(input: { name: string; description?: string }) {
      this.groupStatus = "loading";
      this.groupError = "";
      try {
        const created = await createGroup(input);
        this.groups = [...this.groups, created];
        this.groupStatus = "idle";
        return created;
      } catch (err) {
        this.groupStatus = "error";
        this.groupError = "Unable to create group.";
        throw err;
      }
    },
    async assignGroup(projectId: string, groupId: string, role: ProjectRole) {
      this.projectGroupStatus = "loading";
      this.projectGroupError = "";
      try {
        const created = await addProjectGroup(projectId, {
          groupId,
          role,
        });
        this.projectGroups = [...this.projectGroups, created];
        this.projectGroupStatus = "idle";
        return created;
      } catch (err) {
        this.projectGroupStatus = "error";
        this.projectGroupError = "Unable to assign group to project.";
        throw err;
      }
    },
    async updateGroupRole(
      projectId: string,
      groupId: string,
      role: ProjectRole,
    ) {
      this.projectGroupStatus = "loading";
      this.projectGroupError = "";
      try {
        const updated = await updateProjectGroup(projectId, groupId, {
          role,
        });
        this.projectGroups = this.projectGroups.map((item) =>
          item.groupId === groupId ? updated : item,
        );
        this.projectGroupStatus = "idle";
        return updated;
      } catch (err) {
        this.projectGroupStatus = "error";
        this.projectGroupError = "Unable to update project role.";
        throw err;
      }
    },
    async removeGroup(projectId: string, groupId: string) {
      this.projectGroupStatus = "loading";
      this.projectGroupError = "";
      try {
        await deleteProjectGroup(projectId, groupId);
        this.projectGroups = this.projectGroups.filter(
          (item) => item.groupId !== groupId,
        );
        this.projectGroupStatus = "idle";
      } catch (err) {
        this.projectGroupStatus = "error";
        this.projectGroupError = "Unable to remove group from project.";
        throw err;
      }
    },
    async addMember(groupId: string, userId: string) {
      this.groupMemberStatus = "loading";
      this.groupMemberError = "";
      try {
        const member = await addGroupMember(groupId, { userId });
        this.groupMembers = [...this.groupMembers, member];
        this.groupMemberStatus = "idle";
        return member;
      } catch (err) {
        this.groupMemberStatus = "error";
        this.groupMemberError = "Unable to add group member.";
        throw err;
      }
    },
    async removeMember(groupId: string, userId: string) {
      this.groupMemberStatus = "loading";
      this.groupMemberError = "";
      try {
        await deleteGroupMember(groupId, userId);
        this.groupMembers = this.groupMembers.filter(
          (member) => member.userId !== userId,
        );
        this.groupMemberStatus = "idle";
      } catch (err) {
        this.groupMemberStatus = "error";
        this.groupMemberError = "Unable to remove group member.";
        throw err;
      }
    },
  },
});
