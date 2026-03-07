import { ref, onUnmounted } from "vue";
import { upsertPresence, type PresenceEntry } from "@/lib/api";

const ENABLED =
    (import.meta.env.VITE_LIVE_COLLAB_ENABLED || "").toLowerCase() === "true";

export function usePresence(projectId: () => string) {
    const presenceUsers = ref<PresenceEntry[]>([]);
    let timer: number | null = null;

    const heartbeat = async (view: "board" | "ticket", ticketId?: string) => {
        if (!ENABLED || !projectId()) return;
        try {
            const res = await upsertPresence(projectId(), view, ticketId);
            presenceUsers.value = res.users;
        } catch {
            // non-critical, ignore
        }
    };

    const startHeartbeat = (view: "board" | "ticket", ticketId?: string) => {
        if (!ENABLED) return;
        heartbeat(view, ticketId);
        timer = window.setInterval(() => heartbeat(view, ticketId), 15_000);
    };

    const stopHeartbeat = () => {
        if (timer !== null) {
            window.clearInterval(timer);
            timer = null;
        }
    };

    const handlePresenceUpdate = (users: PresenceEntry[]) => {
        presenceUsers.value = users;
    };

    onUnmounted(() => {
        stopHeartbeat();
        if (ENABLED && projectId()) {
            upsertPresence(projectId(), "board", undefined).catch(() => {});
        }
    });

    return {
        presenceUsers,
        startHeartbeat,
        stopHeartbeat,
        handlePresenceUpdate,
        isEnabled: ENABLED,
    };
}
