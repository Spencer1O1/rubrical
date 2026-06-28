import { readFileSync } from "node:fs";
import { join } from "node:path";
import { Window } from "happy-dom";

export type FixtureComposeReplace = {
  selector: string;
  fragment: string;
};

export type FixtureCompose = {
  base: string;
  replace: FixtureComposeReplace[];
};

export function composeFixtureHtml(fixturesRoot: string, spec: FixtureCompose): string {
  const baseHtml = readFileSync(join(fixturesRoot, spec.base), "utf8");
  const window = new Window();
  window.document.open();
  window.document.write(baseHtml);
  window.document.close();

  for (const { selector, fragment } of spec.replace) {
    const target = window.document.querySelector(selector);
    if (!target) {
      throw new Error(`fixture compose: selector not found: ${selector} (base ${spec.base})`);
    }
    const fragmentHtml = readFileSync(join(fixturesRoot, fragment), "utf8").trim();
    target.outerHTML = fragmentHtml;
  }

  const doctype = baseHtml.startsWith("<!DOCTYPE") ? "<!DOCTYPE html>\n" : "";
  return `${doctype}${window.document.documentElement.outerHTML}\n`;
}
