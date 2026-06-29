export type MultipartFetchResult =
  | { ok: true; base: string }
  | { ok: false; error: string; authRequired?: boolean; base?: string };

export type RubricalMultipartMessage = {
  type: "rubrical-api:multipart";
  path: string;
  fileName: string;
  mimeType: string;
  bytesBase64: string;
  canvasFileId?: string;
};
