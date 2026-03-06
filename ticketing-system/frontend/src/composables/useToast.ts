import { ref } from "vue";

export type ToastType = "success" | "error" | "info" | "warning";

export type Toast = {
    id: string;
    message: string;
    type: ToastType;
    duration: number;
};

const toasts = ref<Toast[]>([]);
let counter = 0;

export function useToast() {
    const push = (message: string, type: ToastType = "info", duration = 4000) => {
        const id = `toast-${++counter}`;
        toasts.value.push({ id, message, type, duration });
        setTimeout(() => dismiss(id), duration);
        return id;
    };

    const dismiss = (id: string) => {
        const idx = toasts.value.findIndex((t) => t.id === id);
        if (idx !== -1) toasts.value.splice(idx, 1);
    };

    const success = (message: string, duration?: number) => push(message, "success", duration);
    const error = (message: string, duration?: number) => push(message, "error", duration);
    const warning = (message: string, duration?: number) => push(message, "warning", duration);
    const info = (message: string, duration?: number) => push(message, "info", duration);

    return { toasts, push, dismiss, success, error, warning, info };
}
