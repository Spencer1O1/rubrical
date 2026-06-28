#!/usr/bin/env node
/**
 * Parse CLI args and optional prompts for fixture PII redaction.
 */

import { createInterface } from "node:readline/promises";
import { stdin as input, stdout as output } from "node:process";
import { join } from "node:path";

function splitList(value) {
  return value
    .split(",")
    .map((part) => part.trim())
    .filter(Boolean);
}

export function parseCleanFixtureArgs(argv) {
  const options = {
    fixturesDir: join(process.cwd(), "fixtures"),
    files: [],
    displayName: process.env.RUBRICAL_FIXTURE_DISPLAY_NAME?.trim() ?? "",
    email: process.env.RUBRICAL_FIXTURE_EMAIL?.trim() ?? "",
    instHosts: splitList(process.env.RUBRICAL_FIXTURE_INST_HOSTS ?? ""),
    skipPrompt: false,
  };

  for (let i = 2; i < argv.length; i++) {
    const arg = argv[i];
    switch (arg) {
      case "--name":
        options.displayName = argv[++i]?.trim() ?? "";
        break;
      case "--email":
        options.email = argv[++i]?.trim() ?? "";
        break;
      case "--inst-host":
        options.instHosts.push(argv[++i]?.trim() ?? "");
        break;
      case "--no-prompt":
        options.skipPrompt = true;
        break;
      case "--help":
      case "-h":
        options.help = true;
        break;
      default:
        if (arg.endsWith(".html")) {
          options.files.push(arg.includes("/") ? arg : join(options.fixturesDir, arg));
        } else if (!arg.startsWith("-")) {
          options.fixturesDir = arg;
        }
        break;
    }
  }

  return options;
}

export function printCleanFixtureHelp() {
  console.log(`Usage: pnpm clean:fixtures [options] [fixtures-dir] [file.html ...]

Prune Canvas HTML snapshots and redact PII before committing fixtures.

Options:
  --name <string>       Canvas display name to replace with "Test User"
  --email <string>      Email address to replace with "student@example.edu"
  --inst-host <host>    Institution Canvas host to replace (repeatable)
  --no-prompt           Do not ask for missing redaction values
  -h, --help            Show this help

Environment (non-interactive):
  RUBRICAL_FIXTURE_DISPLAY_NAME
  RUBRICAL_FIXTURE_EMAIL
  RUBRICAL_FIXTURE_INST_HOSTS   comma-separated hosts

When stdin is a TTY and values are still missing, the script prompts for them.
`);
}

export async function resolveRedactionOptions(options) {
  if (options.skipPrompt || !process.stdin.isTTY) {
    return options;
  }

  const needsName = !options.displayName;
  const needsEmail = !options.email;
  const needsHost = options.instHosts.length === 0;
  if (!needsName && !needsEmail && !needsHost) {
    return options;
  }

  console.log("Fixture redaction — your name/email never get written to the repo.");
  const rl = createInterface({ input, output });
  try {
    if (needsName) {
      options.displayName = (await rl.question("Canvas display name to redact: ")).trim();
    }
    if (needsEmail) {
      options.email = (await rl.question("Email address to redact: ")).trim();
    }
    if (needsHost) {
      const host = (await rl.question("Institution Canvas host to redact (optional): ")).trim();
      if (host) {
        options.instHosts.push(host);
      }
    }
  } finally {
    rl.close();
  }

  return options;
}
