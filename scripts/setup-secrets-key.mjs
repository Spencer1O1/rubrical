#!/usr/bin/env node
/**
 * Ensure .env.local exists and contains SECRETS_ENCRYPTION_KEY.
 * Idempotent — skips when the key is already set to a non-empty value.
 *
 * Usage: pnpm setup:secrets-key
 */

import { copyFileSync, existsSync, readFileSync, writeFileSync } from "node:fs";
import { randomBytes } from "node:crypto";
import { join } from "node:path";

const KEY = "SECRETS_ENCRYPTION_KEY";
const envLocal = join(process.cwd(), ".env.local");
const envExample = join(process.cwd(), ".env.example");

function hasConfiguredKey(content) {
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) {
      continue;
    }
    if (!trimmed.startsWith(`${KEY}=`)) {
      continue;
    }
    const value = trimmed.slice(KEY.length + 1).trim();
    if (value.length > 0) {
      return true;
    }
  }
  return false;
}

function upsertKey(content, value) {
  const lines = content.split("\n");
  let replaced = false;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (trimmed.startsWith(`${KEY}=`) || trimmed.startsWith(`# ${KEY}=`) || trimmed === `#${KEY}=`) {
      lines[i] = `${KEY}=${value}`;
      replaced = true;
      break;
    }
  }

  if (!replaced) {
    const suffix = content.endsWith("\n") || content.length === 0 ? "" : "\n";
    return `${content}${suffix}${KEY}=${value}\n`;
  }

  return lines.join("\n");
}

if (!existsSync(envLocal)) {
  if (existsSync(envExample)) {
    copyFileSync(envExample, envLocal);
    console.log("Created .env.local from .env.example");
  } else {
    writeFileSync(envLocal, "", "utf8");
    console.log("Created empty .env.local");
  }
}

const content = readFileSync(envLocal, "utf8");
if (hasConfiguredKey(content)) {
  console.log(`${KEY} already set in .env.local`);
  process.exit(0);
}

const value = randomBytes(32).toString("base64");
writeFileSync(envLocal, upsertKey(content, value), "utf8");
console.log(`Wrote ${KEY} to .env.local`);
