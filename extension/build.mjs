import * as esbuild from "esbuild";

const dev = process.argv.includes("--dev");
const watch = process.argv.includes("--watch");
const apiBase = dev ? "http://localhost:8787" : "https://rubrical.spencerls.dev";

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
    outfile: "background.js",
    format: "iife",
  },
  {
    ...shared,
    entryPoints: ["src/popup.ts"],
    outfile: "popup.js",
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
