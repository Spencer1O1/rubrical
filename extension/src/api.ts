import { RUBRICAL_API_BASES } from "./api-bases";
import { executeRubricalFetch, type RubricalFetchRequest } from "./api-fetch";

export { RUBRICAL_API_BASES };

async function fetchRubricalApi<T>(
  path: string,
  init?: RequestInit,
): Promise<{ data: T; base: string }> {
  const headers: Record<string, string> = {};
  if (init?.headers) {
    if (init.headers instanceof Headers) {
      init.headers.forEach((value, key) => {
        headers[key] = value;
      });
    } else if (Array.isArray(init.headers)) {
      for (const [key, value] of init.headers) {
        headers[key] = value;
      }
    } else {
      Object.assign(headers, init.headers);
    }
  }

  const request: RubricalFetchRequest = {
    path,
    method: init?.method,
    headers: Object.keys(headers).length > 0 ? headers : undefined,
    body: typeof init?.body === "string" ? init.body : undefined,
  };

  const result = await executeRubricalFetch(request);
  if (!result.ok) {
    throw new Error(result.error);
  }

  return { data: result.data as T, base: result.base };
}

export async function getRubricalJson<T>(path: string): Promise<T | null> {
  try {
    const { data } = await fetchRubricalApi<T>(path, { method: "GET" });
    return data;
  } catch {
    return null;
  }
}

export async function postImport(
  payload: unknown,
): Promise<{ data: { id?: number; redirect?: string }; base: string }> {
  const { data, base } = await fetchRubricalApi<{ id?: number; redirect?: string }>("/imports", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return { data, base };
}
