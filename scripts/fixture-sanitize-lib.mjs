/**
 * Shared PII / head sanitization for Canvas HTML fixtures.
 */

export const CANVAS_HOST = "canvas.instructure.com";
export const REPLACEMENT_DISPLAY_NAME = "Test User";
export const REPLACEMENT_EMAIL = "student@example.edu";

export function extractTitle(html) {
  const match = html.match(/<title>([^<]*)<\/title>/i);
  return match?.[1]?.trim() ?? "Canvas fixture";
}

export function extractCanAttachEntries(html) {
  const match = html.match(/"can_attach_entries"\s*:\s*(true|false)/);
  return match?.[1] === "true";
}

export function extractHtmlOpenTag(html) {
  const match = html.match(/<html[^>]*>/i);
  if (!match) {
    throw new Error("missing <html> tag");
  }
  return match[0];
}

export function extractCourseAssignmentPath(html) {
  const match = html.match(/\/courses\/\d+\/(?:assignments|discussion_topics)\/\d+/);
  return match?.[0] ?? null;
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

/**
 * @param {string} body
 * @param {{ displayName?: string, email?: string, instHosts?: string[] }} [options]
 */
export function sanitizeBody(body, options = {}) {
  let out = body;
  const displayName = options.displayName?.trim() ?? "";
  const email = options.email?.trim() ?? "";
  const instHosts = options.instHosts ?? [];

  for (const host of instHosts) {
    const trimmed = host.trim();
    if (!trimmed) {
      continue;
    }
    out = out.replaceAll(`https://${trimmed}`, `https://${CANVAS_HOST}`);
    out = out.replaceAll(`http://${trimmed}`, `https://${CANVAS_HOST}`);
    out = out.replaceAll(trimmed, CANVAS_HOST);
  }

  out = out.replace(/([?&])verifier=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])AWSAccessKeyId=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])Signature=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])Expires=\d+/gi, "");
  out = out.replace(/([?&])response-cache-control=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])response-expires=[^"'&\s<>]*/gi, "");
  out = out.replace(/\?&/g, "?");
  out = out.replace(/\?(?=["'\s>])/g, "");

  out = out.replace(
    /(\/feeds\/topics\/\d+\/enrollment_)[A-Za-z0-9]+/g,
    "$1fixture-token",
  );

  if (displayName) {
    out = out.replaceAll(displayName, REPLACEMENT_DISPLAY_NAME);
  }
  if (email) {
    out = out.replaceAll(email, REPLACEMENT_EMAIL);
  }

  out = out.replace(
    /page_view_token=eyJ[A-Za-z0-9._-]+/g,
    "page_view_token=fixture",
  );

  out = out.replace(/<script\b[\s\S]*?<\/script>/gi, "");

  return out;
}

export function normalizeHtmlOpenTag(htmlOpenTag) {
  if (/position:\s*fixed/i.test(htmlOpenTag)) {
    return htmlOpenTag;
  }
  return '<html lang="en">';
}

export function buildMinimalHtml({ htmlOpenTag, title, body, canAttachEntries, coursePath }) {
  const env =
    canAttachEntries === undefined
      ? "{}"
      : JSON.stringify({ can_attach_entries: canAttachEntries });

  const pathLink =
    coursePath === null
      ? ""
      : `\n  <a href="https://${CANVAS_HOST}${coursePath}" style="display:none" id="fixture-path-link">fixture path</a>`;

  const innerBody = body.replace(/^<body[^>]*>/i, "").replace(/<\/body>\s*$/i, "");

  return `<!DOCTYPE html>
${normalizeHtmlOpenTag(htmlOpenTag)}
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${title}</title>
  <script>
    INST = {};
    ENV = ${env};
    REMOTES = {};
  </script>
</head>
<body>${innerBody}${pathLink}
</body>
</html>
`;
}

/**
 * @param {string} html
 * @param {string} filename
 * @param {{ displayName?: string, email?: string, instHosts?: string[] }} [options]
 */
export function assertClean(html, filename, options = {}) {
  const forbidden = [
    [/AKIA[0-9A-Z]{16}/, "AWS access key"],
    [/AWSAccessKeyId=/i, "AWS signed URL"],
    [/page_view_token=eyJ/i, "JWT page view token"],
  ];

  const displayName = options.displayName?.trim() ?? "";
  const email = options.email?.trim() ?? "";
  if (displayName) {
    forbidden.push([new RegExp(escapeRegExp(displayName), "g"), "display name"]);
  }
  if (email) {
    forbidden.push([new RegExp(escapeRegExp(email), "g"), "email address"]);
  }
  for (const host of options.instHosts ?? []) {
    const trimmed = host.trim();
    if (trimmed) {
      forbidden.push([new RegExp(escapeRegExp(trimmed), "i"), `institution host ${trimmed}`]);
    }
  }

  for (const [pattern, label] of forbidden) {
    if (pattern.test(html)) {
      throw new Error(`${filename}: still contains ${label}`);
    }
  }
}
