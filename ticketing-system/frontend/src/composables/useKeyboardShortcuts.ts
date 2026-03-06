import { onMounted, onUnmounted, ref } from "vue";

export type KeyboardShortcut = {
    key: string;
    modifier?: "ctrl" | "meta" | "alt" | "shift";
    handler: () => void;
    preventDefault?: boolean;
    description?: string;
};

export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[]) {
    const isEnabled = ref(true);

    const handleKeyDown = (event: KeyboardEvent) => {
        if (!isEnabled.value) return;

        // Don't trigger shortcuts when typing in input fields
        const target = event.target as HTMLElement;
        if (
            target.tagName === "INPUT" ||
            target.tagName === "TEXTAREA" ||
            target.isContentEditable
        ) {
            return;
        }

        for (const shortcut of shortcuts) {
            const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase();

            let modifierMatches = true;
            if (shortcut.modifier) {
                switch (shortcut.modifier) {
                    case "ctrl":
                        modifierMatches = event.ctrlKey;
                        break;
                    case "meta":
                        modifierMatches = event.metaKey;
                        break;
                    case "alt":
                        modifierMatches = event.altKey;
                        break;
                    case "shift":
                        modifierMatches = event.shiftKey;
                        break;
                }
            } else {
                // If no modifier specified, ensure no modifiers are pressed
                modifierMatches = !event.ctrlKey && !event.metaKey && !event.altKey;
            }

            if (keyMatches && modifierMatches) {
                if (shortcut.preventDefault !== false) {
                    event.preventDefault();
                }
                shortcut.handler();
                break;
            }
        }
    };

    const enable = () => {
        isEnabled.value = true;
    };

    const disable = () => {
        isEnabled.value = false;
    };

    onMounted(() => {
        document.addEventListener("keydown", handleKeyDown);
    });

    onUnmounted(() => {
        document.removeEventListener("keydown", handleKeyDown);
    });

    return {
        isEnabled,
        enable,
        disable,
    };
}

// Common shortcuts for the ticketing app
export function createTicketingShortcuts(options: {
    onNewTicket: () => void;
    onSearch?: () => void;
    onCloseModal?: () => void;
    onRefresh?: () => void;
    onToggleTheme?: () => void;
}) {
    const shortcuts: KeyboardShortcut[] = [
        {
            key: "n",
            handler: options.onNewTicket,
            description: "Create new ticket",
        },
        {
            key: "/",
            handler: options.onSearch || (() => {}),
            description: "Search tickets",
        },
        {
            key: "k",
            modifier: "meta",
            handler: options.onSearch || (() => {}),
            description: "Search tickets (Cmd+K)",
        },
        {
            key: "k",
            modifier: "ctrl",
            handler: options.onSearch || (() => {}),
            description: "Search tickets (Ctrl+K)",
        },
        {
            key: "Escape",
            handler: options.onCloseModal || (() => {}),
            description: "Close modal or panel",
        },
        {
            key: "r",
            modifier: "meta",
            handler: options.onRefresh || (() => {}),
            description: "Refresh data (Cmd+R without browser reload)",
            preventDefault: false, // Allow browser refresh if handler doesn't prevent
        },
    ];

    return shortcuts;
}
