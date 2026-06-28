import { executeRubricalFetchDirect, executeRubricalMultipartDirect } from "./api-direct";
import { isRubricalApiMessage } from "./api-messages";
import { handleStagedFilesMessage } from "./staged-files/idb";
import type { StagedFilesMessage } from "./staged-files/types";

function isStagedFilesMessage(message: unknown): message is StagedFilesMessage {
  return (
    typeof message === "object" &&
    message !== null &&
    "type" in message &&
    typeof (message as { type?: unknown }).type === "string" &&
    (message as { type: string }).type.startsWith("staged-files:")
  );
}

chrome.runtime.onMessage.addListener((message: unknown, _sender, sendResponse) => {
  if (isStagedFilesMessage(message)) {
    void handleStagedFilesMessage(message).then(sendResponse);
    return true;
  }

  if (!isRubricalApiMessage(message)) {
    return false;
  }

  if (message.type === "rubrical-api:fetch") {
    void executeRubricalFetchDirect(message.request, message.maxAttempts).then(sendResponse);
    return true;
  }

  void executeRubricalMultipartDirect(message).then(sendResponse);
  return true;
});
