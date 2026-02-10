import { ComponentBaseMapper, EventStateRegistry, TriggerRenderer } from "../types";
import { onAppMentionTriggerRenderer } from "./on_app_mention";
import { sendTextMessageMapper } from "./send_text_message";
import { sendAndWaitForResponseMapper } from "./send_and_wait_for_response";
import { buildActionStateRegistry } from "../utils";

export const componentMappers: Record<string, ComponentBaseMapper> = {
  sendTextMessage: sendTextMessageMapper,
  sendAndWaitForResponse: sendAndWaitForResponseMapper,
};

export const triggerRenderers: Record<string, TriggerRenderer> = {
  onAppMention: onAppMentionTriggerRenderer,
};

export const eventStateRegistry: Record<string, EventStateRegistry> = {
  sendTextMessage: buildActionStateRegistry("sent"),
  sendAndWaitForResponse: buildActionStateRegistry("waiting"),
};
