/**
 * Simple template rendering utility
 * Replaces {{variable}} placeholders with values
 */

export function renderTemplate(
  template: string,
  context: Record<string, unknown>
): string {
  let result = template;

  for (const [key, value] of Object.entries(context)) {
    if (value === null || value === undefined) {
      continue;
    }

    const placeholder = new RegExp(`{{\\s*${key}\\s*}}`, "g");
    const stringValue = typeof value === "string" ? value : JSON.stringify(value);
    result = result.replace(placeholder, stringValue);
  }

  return result;
}

export function renderTemplateStrict(
  template: string,
  context: Record<string, unknown>
): string {
  const result = renderTemplate(template, context);

  // Check for unreplaced placeholders
  const unreplaced = result.match(/{{.*?}}/g);
  if (unreplaced) {
    throw new Error(
      `Unreplaced template placeholders: ${unreplaced.join(", ")}`
    );
  }

  return result;
}
