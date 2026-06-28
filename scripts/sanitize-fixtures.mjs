#!/usr/bin/env node
/**
 * Strip Canvas page snapshot bloat (ENV/INST, analytics, signed URLs, PII)
 * while keeping DOM structure, data-testid, and InstUI classes tests rely on.
 *
 * Usage: node scripts/sanitize-fixtures.mjs [fixtures-dir]
 */

import { readFileSync, writeFileSync, readdirSync } from "node:fs";
import { join } from "node:path";

const FIXTURES_DIR = process.argv[2] ?? join(process.cwd(), "fixtures");
const CANVAS_HOST = "canvas.instructure.com";

function extractTitle(html) {
  const match = html.match(/<title>([^<]*)<\/title>/i);
  return match?.[1]?.trim() ?? "Canvas fixture";
}

function extractCanAttachEntries(html) {
  const match = html.match(/"can_attach_entries"\s*:\s*(true|false)/);
  return match?.[1] === "true";
}

function extractHtmlOpenTag(html) {
  const match = html.match(/<html[^>]*>/i);
  if (!match) {
    throw new Error("missing <html> tag");
  }
  return match[0];
}

function extractBody(html) {
  const start = html.search(/<body[\s>]/i);
  if (start < 0) {
    throw new Error("missing <body>");
  }

  const endTag = html.search(/<\/body>/i);
  if (endTag < 0) {
    throw new Error("missing </body>");
  }

  return html.slice(start, endTag + "</body>".length);
}

function sanitizeBody(body) {
  let out = body;

  // Institution host → generic Canvas host (relative paths unchanged).
  out = out.replaceAll("https://usu.instructure.com", `https://${CANVAS_HOST}`);
  out = out.replaceAll("http://usu.instructure.com", `https://${CANVAS_HOST}`);
  out = out.replaceAll("usu.instructure.com", CANVAS_HOST);
  out = out.replaceAll("usu.eesysoft.com", "example.edu");

  // Signed download / S3 query junk.
  out = out.replace(/([?&])verifier=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])AWSAccessKeyId=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])Signature=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])Expires=\d+/gi, "");
  out = out.replace(/([?&])response-cache-control=[^"'&\s<>]*/gi, "");
  out = out.replace(/([?&])response-expires=[^"'&\s<>]*/gi, "");
  out = out.replace(/\?&/g, "?");
  out = out.replace(/\?(?=["'\s>])/g, "");

  // Atom feed enrollment tokens.
  out = out.replace(
    /(\/feeds\/topics\/\d+\/enrollment_)[A-Za-z0-9]+/g,
    "$1fixture-token",
  );

  // PII placeholders.
  out = out.replace(/Spencer Smith/g, "Test User");
  out = out.replace(/a02412218@[^\s"<>]+/g, "student@example.edu");
  out = out.replace(
    /page_view_token=eyJ[A-Za-z0-9._-]+/g,
    "page_view_token=fixture",
  );

  // Drop inline scripts (institution analytics); DOM is what tests parse.
  out = out.replace(/<script\b[\s\S]*?<\/script>/gi, "");

  return out;
}

function buildMinimalHtml({ htmlOpenTag, title, body, canAttachEntries }) {
  const env =
    canAttachEntries === undefined
      ? "{}"
      : JSON.stringify({ can_attach_entries: canAttachEntries });

  return `<!DOCTYPE html>
${htmlOpenTag}
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
${body}
</html>
`;
}

function sanitizeFixture(html, filename) {
  const title = extractTitle(html);
  const htmlOpenTag = extractHtmlOpenTag(html);
  const canAttach = html.includes('"can_attach_entries"')
    ? extractCanAttachEntries(html)
    : undefined;
  const body = sanitizeBody(extractBody(html));
  return buildMinimalHtml({ htmlOpenTag, title, body, canAttachEntries: canAttach });
}

function assertClean(html, filename) {
  const forbidden = [
    [/AKIA[0-9A-Z]{16}/, "AWS access key"],
    [/AWSAccessKeyId=/i, "AWS signed URL"],
    [/page_view_token=eyJ/i, "JWT page view token"],
    [/a02412218@/i, "student email"],
    [/Spencer Smith/, "display name"],
    [/usu\.instructure\.com/i, "institution host"],
  ];

  for (const [pattern, label] of forbidden) {
    if (pattern.test(html)) {
      throw new Error(`${filename}: still contains ${label}`);
    }
  }
}

const files = readdirSync(FIXTURES_DIR)
  .filter((name) => name.endsWith(".html"))
  .sort();

let beforeBytes = 0;
let afterBytes = 0;

for (const name of files) {
  const path = join(FIXTURES_DIR, name);
  const original = readFileSync(path, "utf8");
  beforeBytes += original.length;

  const sanitized = sanitizeFixture(original, name);
  assertClean(sanitized, name);
  writeFileSync(path, sanitized, "utf8");
  afterBytes += sanitized.length;

  console.log(
    `${name}: ${(original.length / 1024).toFixed(1)}KB → ${(sanitized.length / 1024).toFixed(1)}KB`,
  );
}

console.log(
  `\nSanitized ${files.length} fixtures: ${(beforeBytes / 1024 / 1024).toFixed(2)}MB → ${(afterBytes / 1024 / 1024).toFixed(2)}MB`,
);
