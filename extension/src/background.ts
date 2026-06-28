import { executeRubricalFetchDirect, executeRubricalMultipartDirect } from "./api-direct";
import { isRubricalApiMessage } from "./api-messages";

chrome.runtime.onMessage.addListener((message: unknown, _sender, sendResponse) => {
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
