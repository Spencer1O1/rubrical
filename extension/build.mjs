import * as esbuild from "esbuild";

const watch = process.argv.includes("--watch");

// Same knob as the Go server (PUBLIC_URL). Default matches internal/config.DefaultPublicURL.
function resolveApiBase() {
  const fromEnv = process.env.PUBLIC_URL?.trim().replace(/\/$/, "");
  if (fromEnv) {
    return fromEnv;
  }
  return "http://localhost:8787";
}

const apiBase = resolveApiBase();

const shared = {
  define: {
    __RUBRICAL_API_BASE__: JSON.stringify(apiBase),
  },
  target: "chrome120",
  logLevel: "info",
  bundle: true,
};

const builds = [
  {
    ...shared,
    entryPoints: ["src/content.ts"],
    outfile: "dist/content.js",
    format: "iife",
    loader: { ".css": "text" },
  },
  {
    ...shared,
    entryPoints: ["src/background.ts"],
    outfile: "dist/background.js",
    format: "iife",
  },
  {
    ...shared,
    entryPoints: ["src/popup.ts"],
    outfile: "dist/popup.js",
    format: "esm",
  },
];

if (watch) {
  const contexts = await Promise.all(builds.map((options) => esbuild.context(options)));
  await Promise.all(contexts.map((context) => context.watch()));
  console.log(`Watching extension (API base: ${apiBase})`);
} else {
  await Promise.all(builds.map((options) => esbuild.build(options)));
  console.log(`Built extension (API base: ${apiBase})`);
}
