import type { RubricalFetchRequest, RubricalFetchResult } from "./api-fetch-types";
import type { MultipartFetchResult, RubricalMultipartMessage } from "./api-multipart-types";

export type RubricalApiFetchMessage = {
  type: "rubrical-api:fetch";
  request: RubricalFetchRequest;
  maxAttempts?: number;
};

export type RubricalApiMessage = RubricalApiFetchMessage | RubricalMultipartMessage;

export type RubricalApiResponse = RubricalFetchResult | MultipartFetchResult;

export function isRubricalApiMessage(message: unknown): message is RubricalApiMessage {
  if (typeof message !== "object" || message === null || !("type" in message)) {
    return false;
  }

  const type = (message as { type: string }).type;
  return type === "rubrical-api:fetch" || type === "rubrical-api:multipart";
}
