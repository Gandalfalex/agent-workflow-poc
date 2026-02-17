export type ActionResult = {
    value: string;
    start: number;
    end: number;
};

export type ToolbarAction = {
    id: string;
    label: string;
    icon: string;
    shortcutKey?: string;
    compact: boolean;
    apply: (el: HTMLTextAreaElement) => ActionResult;
};

function wrapSelection(
    el: HTMLTextAreaElement,
    before: string,
    after: string,
): ActionResult {
    const { value, selectionStart: s, selectionEnd: e } = el;
    const selected = value.slice(s, e);

    // If already wrapped, unwrap
    const bLen = before.length;
    const aLen = after.length;
    if (
        s >= bLen &&
        value.slice(s - bLen, s) === before &&
        value.slice(e, e + aLen) === after
    ) {
        return {
            value:
                value.slice(0, s - bLen) + selected + value.slice(e + aLen),
            start: s - bLen,
            end: e - bLen,
        };
    }

    const wrapped = before + (selected || "text") + after;
    return {
        value: value.slice(0, s) + wrapped + value.slice(e),
        start: s + bLen,
        end: s + bLen + (selected ? selected.length : 4),
    };
}

function prefixLines(
    el: HTMLTextAreaElement,
    prefixer: (line: string, index: number) => string,
): ActionResult {
    const { value, selectionStart: s, selectionEnd: e } = el;

    // Find full lines covering the selection
    const lineStart = value.lastIndexOf("\n", s - 1) + 1;
    const lineEnd = value.indexOf("\n", e);
    const blockEnd = lineEnd === -1 ? value.length : lineEnd;

    const block = value.slice(lineStart, blockEnd);
    const lines = block.split("\n");
    const prefixed = lines.map(prefixer).join("\n");

    return {
        value: value.slice(0, lineStart) + prefixed + value.slice(blockEnd),
        start: lineStart,
        end: lineStart + prefixed.length,
    };
}

export const toolbarActions: ToolbarAction[] = [
    {
        id: "bold",
        label: "Bold",
        icon: "Bold",
        shortcutKey: "b",
        compact: true,
        apply: (el) => wrapSelection(el, "**", "**"),
    },
    {
        id: "italic",
        label: "Italic",
        icon: "Italic",
        shortcutKey: "i",
        compact: true,
        apply: (el) => wrapSelection(el, "*", "*"),
    },
    {
        id: "code",
        label: "Code",
        icon: "Code",
        shortcutKey: "e",
        compact: true,
        apply: (el) => wrapSelection(el, "`", "`"),
    },
    {
        id: "link",
        label: "Link",
        icon: "Link",
        shortcutKey: "k",
        compact: true,
        apply: (el) => {
            const { value, selectionStart: s, selectionEnd: e } = el;
            const selected = value.slice(s, e);
            const linkText = selected || "text";
            const insertion = `[${linkText}](url)`;
            return {
                value: value.slice(0, s) + insertion + value.slice(e),
                start: s + linkText.length + 3,
                end: s + linkText.length + 6,
            };
        },
    },
    {
        id: "unordered-list",
        label: "Bullet list",
        icon: "List",
        compact: false,
        apply: (el) => prefixLines(el, (_line, _i) => `- ${_line}`),
    },
    {
        id: "ordered-list",
        label: "Numbered list",
        icon: "ListOrdered",
        compact: false,
        apply: (el) => prefixLines(el, (line, i) => `${i + 1}. ${line}`),
    },
    {
        id: "quote",
        label: "Quote",
        icon: "Quote",
        compact: false,
        apply: (el) => prefixLines(el, (line) => `> ${line}`),
    },
    {
        id: "heading",
        label: "Heading",
        icon: "Heading2",
        compact: false,
        apply: (el) => {
            const { value, selectionStart: s } = el;
            const lineStart = value.lastIndexOf("\n", s - 1) + 1;
            const lineEnd = value.indexOf("\n", s);
            const end = lineEnd === -1 ? value.length : lineEnd;
            const line = value.slice(lineStart, end);

            // Toggle: if already has ##, remove it
            if (line.startsWith("## ")) {
                const newLine = line.slice(3);
                return {
                    value:
                        value.slice(0, lineStart) +
                        newLine +
                        value.slice(end),
                    start: lineStart,
                    end: lineStart + newLine.length,
                };
            }

            const newLine = `## ${line}`;
            return {
                value:
                    value.slice(0, lineStart) + newLine + value.slice(end),
                start: lineStart + 3,
                end: lineStart + newLine.length,
            };
        },
    },
];

export function handleTab(
    el: HTMLTextAreaElement,
    shiftKey: boolean,
): ActionResult {
    const { value, selectionStart: s, selectionEnd: e } = el;

    if (shiftKey) {
        // Outdent: remove up to 2 leading spaces from current line
        const lineStart = value.lastIndexOf("\n", s - 1) + 1;
        const line = value.slice(lineStart);
        const spaces = line.match(/^ {1,2}/)?.[0]?.length ?? 0;
        if (spaces === 0) return { value, start: s, end: e };
        return {
            value:
                value.slice(0, lineStart) +
                value.slice(lineStart + spaces),
            start: Math.max(lineStart, s - spaces),
            end: Math.max(lineStart, e - spaces),
        };
    }

    // Indent: insert 2 spaces at cursor
    return {
        value: value.slice(0, s) + "  " + value.slice(e),
        start: s + 2,
        end: s + 2,
    };
}
