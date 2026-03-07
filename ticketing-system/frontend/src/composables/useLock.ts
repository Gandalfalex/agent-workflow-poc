import { ref, onUnmounted } from "vue";
import {
    acquireTicketLock,
    renewTicketLock,
    releaseTicketLock,
    type TicketLock,
    type TicketLockConflict,
} from "@/lib/api";

const ENABLED =
    (import.meta.env.VITE_LIVE_COLLAB_ENABLED || "").toLowerCase() === "true";

// Renew 20s before the 2-minute TTL expires
const RENEW_INTERVAL_MS = 90_000;

export function useLock() {
    const activeLock = ref<TicketLock | null>(null);
    const lockConflict = ref<TicketLockConflict | null>(null);
    let renewTimer: number | null = null;
    let currentTicketId: string | null = null;

    const stopRenew = () => {
        if (renewTimer !== null) {
            window.clearInterval(renewTimer);
            renewTimer = null;
        }
    };

    const tryAcquire = async (ticketId: string): Promise<boolean> => {
        if (!ENABLED) return true;
        currentTicketId = ticketId;
        try {
            const lock = await acquireTicketLock(ticketId);
            activeLock.value = lock;
            lockConflict.value = null;
            // Schedule renewal
            stopRenew();
            renewTimer = window.setInterval(async () => {
                if (!currentTicketId) return;
                try {
                    const renewed = await renewTicketLock(currentTicketId);
                    activeLock.value = renewed;
                } catch {
                    activeLock.value = null;
                }
            }, RENEW_INTERVAL_MS);
            return true;
        } catch (err: unknown) {
            const error = err as Error & { status?: number };
            if (error.status === 409) {
                try {
                    const body = JSON.parse(error.message) as TicketLockConflict;
                    lockConflict.value = body;
                } catch {
                    lockConflict.value = {
                        error: "ticket_locked",
                        lockedBy: "",
                        lockedByName: "another user",
                        expiresAt: "",
                    };
                }
            }
            activeLock.value = null;
            return false;
        }
    };

    const release = async () => {
        stopRenew();
        if (!ENABLED || !currentTicketId) return;
        try {
            await releaseTicketLock(currentTicketId);
        } catch {
            // best-effort
        }
        activeLock.value = null;
        lockConflict.value = null;
        currentTicketId = null;
    };

    const clearConflict = () => {
        lockConflict.value = null;
    };

    onUnmounted(() => {
        release();
    });

    return {
        activeLock,
        lockConflict,
        tryAcquire,
        release,
        clearConflict,
        isEnabled: ENABLED,
    };
}
