import { computed, ref } from "vue";

type Locale = "en" | "de";
type TranslationTable = Record<string, string>;

const STORAGE_KEY = "app.locale";

const messages: Record<Locale, TranslationTable> = {
    en: {
        "common.clear": "Clear",
        "common.loading": "Loading...",
        "board.search.placeholder": "Filter tickets...",
        "board.filter.allStates": "All states",
        "board.filter.allAssignees": "All assignees",
        "board.filter.allPriorities": "All priorities",
        "board.filter.allTypes": "All types",
        "board.filter.blockedOnly": "Blocked only",
        "board.filter.selectPreset": "Select preset...",
        "board.filter.presetName": "Preset name",
        "board.filter.loadingPresets": "Loading presets...",
        "board.preset.save": "Save",
        "board.preset.rename": "Rename",
        "board.preset.delete": "Delete",
        "board.preset.share": "Share",
        "board.bulk.moveState": "Move state",
        "board.bulk.assignUser": "Assign user",
        "board.bulk.setPriority": "Set priority",
        "board.bulk.deleteTickets": "Delete tickets",
        "board.bulk.selectState": "Select state...",
        "board.bulk.selectAssignee": "Select assignee...",
        "board.bulk.deleteHint": "Deletes selected tickets.",
        "board.bulk.apply": "Apply bulk action",
        "board.bulk.applying": "Applying...",
        "board.view.loading": "Loading board...",
        "board.view.setup": "Setup",
        "board.view.createWorkflow": "Create your first workflow",
        "board.view.emptyWorkflow":
            "The board is empty because no workflow states exist yet. Initialize a default board to start adding tickets.",
        "board.view.creating": "Creating...",
        "board.view.initialize": "Initialize board",
        "board.view.storySection": "Board stories",
        "board.view.storyHelp":
            "Stories group tickets horizontally across workflow states.",
        "board.view.storiesCount": "{count} stories",
        "board.view.ticketsCount": "{count} tickets",
        "board.view.exitSelect": "Exit select",
        "board.view.selectTickets": "Select tickets",
        "board.view.selectedCount": "{count} selected",
        "board.view.addStory": "Add story",
        "board.view.noMatchingTitle": "No matching tickets",
        "board.view.noMatchingBody": "Nothing matched",
        "board.view.clearFilter": "Clear filter",
        "board.view.story": "Story",
        "board.view.deleteStory": "Delete story",
        "board.view.storyLabel": "Story",
        "board.view.ticketCount": "{count} tickets",
        "board.view.addTicket": "+ Add ticket",
        "board.view.blocked": "Blocked ({count})",
        "board.view.unassigned": "Unassigned",
        "board.view.dropHere": "Drop here",
        "header.board": "Board",
        "header.dashboard": "Dashboard",
        "header.settings": "Settings",
        "header.selectProject": "Select project",
        "header.inbox": "Inbox",
        "header.notifications": "Notifications",
        "header.markAllRead": "Mark all read",
        "header.close": "Close",
        "header.mentions": "Mentions",
        "header.assignments": "Assignments",
        "header.noNotifications": "No notifications.",
        "header.markRead": "Mark read",
        "header.logout": "Logout",
        "header.liveWs": "Live updates via WebSocket",
        "header.livePolling": "Fallback updates via polling",
    },
    de: {
        "common.clear": "Leeren",
        "common.loading": "Lädt...",
        "board.search.placeholder": "Tickets filtern...",
        "board.filter.allStates": "Alle Status",
        "board.filter.allAssignees": "Alle Zuständigen",
        "board.filter.allPriorities": "Alle Prioritäten",
        "board.filter.allTypes": "Alle Typen",
        "board.filter.blockedOnly": "Nur blockierte",
        "board.filter.selectPreset": "Preset wählen...",
        "board.filter.presetName": "Preset-Name",
        "board.filter.loadingPresets": "Presets laden...",
        "board.preset.save": "Speichern",
        "board.preset.rename": "Umbenennen",
        "board.preset.delete": "Löschen",
        "board.preset.share": "Teilen",
        "board.bulk.moveState": "Status ändern",
        "board.bulk.assignUser": "Zuweisen",
        "board.bulk.setPriority": "Priorität setzen",
        "board.bulk.deleteTickets": "Tickets löschen",
        "board.bulk.selectState": "Status wählen...",
        "board.bulk.selectAssignee": "Zuständige wählen...",
        "board.bulk.deleteHint": "Löscht ausgewählte Tickets.",
        "board.bulk.apply": "Sammelaktion anwenden",
        "board.bulk.applying": "Wird angewendet...",
        "board.view.loading": "Board wird geladen...",
        "board.view.setup": "Einrichtung",
        "board.view.createWorkflow": "Ersten Workflow erstellen",
        "board.view.emptyWorkflow":
            "Das Board ist leer, weil noch keine Workflow-Status existieren. Initialisiere ein Standard-Board, um Tickets anzulegen.",
        "board.view.creating": "Wird erstellt...",
        "board.view.initialize": "Board initialisieren",
        "board.view.storySection": "Board Stories",
        "board.view.storyHelp":
            "Stories gruppieren Tickets horizontal über Workflow-Status.",
        "board.view.storiesCount": "{count} Stories",
        "board.view.ticketsCount": "{count} Tickets",
        "board.view.exitSelect": "Auswahl beenden",
        "board.view.selectTickets": "Tickets auswählen",
        "board.view.selectedCount": "{count} ausgewählt",
        "board.view.addStory": "Story hinzufügen",
        "board.view.noMatchingTitle": "Keine passenden Tickets",
        "board.view.noMatchingBody": "Nichts gefunden für",
        "board.view.clearFilter": "Filter leeren",
        "board.view.story": "Story",
        "board.view.deleteStory": "Story löschen",
        "board.view.storyLabel": "Story",
        "board.view.ticketCount": "{count} Tickets",
        "board.view.addTicket": "+ Ticket hinzufügen",
        "board.view.blocked": "Blockiert ({count})",
        "board.view.unassigned": "Nicht zugewiesen",
        "board.view.dropHere": "Hier ablegen",
        "header.board": "Board",
        "header.dashboard": "Dashboard",
        "header.settings": "Einstellungen",
        "header.selectProject": "Projekt wählen",
        "header.inbox": "Inbox",
        "header.notifications": "Benachrichtigungen",
        "header.markAllRead": "Alle gelesen",
        "header.close": "Schließen",
        "header.mentions": "Erwähnungen",
        "header.assignments": "Zuweisungen",
        "header.noNotifications": "Keine Benachrichtigungen.",
        "header.markRead": "Als gelesen",
        "header.logout": "Abmelden",
        "header.liveWs": "Live-Updates per WebSocket",
        "header.livePolling": "Fallback-Updates per Polling",
    },
};

const detectLocale = (): Locale => {
    if (typeof window === "undefined") return "en";
    const saved = window.localStorage.getItem(STORAGE_KEY);
    if (saved === "en" || saved === "de") return saved;
    const browser = window.navigator.language.toLowerCase();
    return browser.startsWith("de") ? "de" : "en";
};

const activeLocale = ref<Locale>(detectLocale());

if (typeof document !== "undefined") {
    document.documentElement.lang = activeLocale.value;
}

const interpolate = (
    text: string,
    params?: Record<string, string | number>,
): string => {
    if (!params) return text;
    return Object.entries(params).reduce(
        (result, [key, value]) =>
            result.split(`{${key}}`).join(String(value)),
        text,
    );
};

export const useI18n = () => {
    const setLocale = (locale: Locale) => {
        activeLocale.value = locale;
        if (typeof document !== "undefined") {
            document.documentElement.lang = locale;
        }
        if (typeof window !== "undefined") {
            window.localStorage.setItem(STORAGE_KEY, locale);
        }
    };

    const t = (key: string, params?: Record<string, string | number>) => {
        const dict = messages[activeLocale.value] || messages.en;
        const template = dict[key] || messages.en[key] || key;
        return interpolate(template, params);
    };

    return {
        availableLocales: ["en", "de"] as const,
        locale: computed(() => activeLocale.value),
        setLocale,
        t,
    };
};
