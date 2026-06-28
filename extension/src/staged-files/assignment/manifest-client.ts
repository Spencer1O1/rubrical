import { getRubricalJson } from "../../api";
import { normalizeSourceUrl } from "../normalize-source-url";
import type { DraftManifest } from "../types";

let cachedManifest: DraftManifest | null = null;
let manifestFetched = false;
let manifestLoad: Promise<DraftManifest> | null = null;

async function fetchDraftManifest(sourceUrl = window.location.href): Promise<DraftManifest | null> {
  const normalized = normalizeSourceUrl(sourceUrl);
  if (!normalized) {
    return null;
  }

  const query = new URLSearchParams({ sourceUrl: normalized });
  const data = await getRubricalJson<DraftManifest>(`/assignments/draft-manifest?${query.toString()}`);
  if (!data || !Array.isArray(data.files)) {
    return null;
  }

  return data;
}

/** In-memory manifest from the single page-load GET (or after import/panel refresh). */
export function getDraftManifest(): DraftManifest {
  return cachedManifest ?? { files: [] };
}

/** One network fetch per assignment visit. */
export async function fetchDraftManifestOnce(): Promise<DraftManifest> {
  if (manifestFetched) {
    return getDraftManifest();
  }

  if (!manifestLoad) {
    manifestLoad = fetchDraftManifest()
      .then((manifest) => {
        if (manifest) {
          cachedManifest = manifest;
        }
        manifestFetched = true;
        return getDraftManifest();
      })
      .catch(() => {
        manifestFetched = true;
        return getDraftManifest();
      })
      .finally(() => {
        manifestLoad = null;
      });
  }

  return manifestLoad;
}

export async function reloadDraftManifest(): Promise<DraftManifest> {
  const manifest = await fetchDraftManifest();
  if (manifest) {
    manifestFetched = true;
    cachedManifest = manifest;
  }
  return getDraftManifest();
}
