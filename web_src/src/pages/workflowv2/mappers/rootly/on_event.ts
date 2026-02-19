import { getBackgroundColorClass } from "@/utils/colors";
import { formatTimeAgo } from "@/utils/date";
import { TriggerEventContext, TriggerRenderer, TriggerRendererContext } from "../types";
import { TriggerProps } from "@/ui/trigger";
import rootlyIcon from "@/assets/icons/integrations/rootly.svg";
import { Incident } from "./types";

interface OnEventEventData {
  id?: string;
  event?: string;
  kind?: string;
  visibility?: string;
  occurred_at?: string;
  created_at?: string;
  user_display_name?: string;
  event_source?: string;
  incident?: Incident;
}

export const onEventTriggerRenderer: TriggerRenderer = {
  getTitleAndSubtitle: (context: TriggerEventContext) => {
    const eventData = context.event?.data as OnEventEventData;
    const content = eventData?.event
      ? eventData.event.length > 60
        ? eventData.event.substring(0, 60) + "..."
        : eventData.event
      : "Incident Event";
    const parts = [eventData?.kind, eventData?.user_display_name].filter(Boolean).join(" · ");
    const timeAgo = context.event?.createdAt ? formatTimeAgo(new Date(context.event.createdAt)) : "";
    const subtitle = [parts, timeAgo].filter(Boolean).join(" · ");
    return { title: content, subtitle };
  },

  getRootEventValues: (context: TriggerEventContext) => {
    const eventData = context.event?.data as OnEventEventData;
    const details: Record<string, string> = {};
    if (eventData?.id) details["ID"] = eventData.id;
    if (eventData?.event) details["Event"] = eventData.event;
    if (eventData?.kind) details["Kind"] = eventData.kind;
    if (eventData?.visibility) details["Visibility"] = eventData.visibility;
    if (eventData?.user_display_name) details["Created By"] = eventData.user_display_name;
    if (eventData?.occurred_at) details["Occurred At"] = new Date(eventData.occurred_at).toLocaleString();
    if (eventData?.created_at) details["Created At"] = new Date(eventData.created_at).toLocaleString();
    if (eventData?.incident?.id) details["Incident ID"] = eventData.incident.id;
    if (eventData?.incident?.title) details["Incident Title"] = eventData.incident.title;
    if (eventData?.incident?.status) details["Incident Status"] = eventData.incident.status;
    if (eventData?.incident?.severity) details["Incident Severity"] = eventData.incident.severity;
    return details;
  },

  getTriggerProps: (context: TriggerRendererContext) => {
    const { node, definition, lastEvent } = context;
    const configuration = node.configuration as {
      incidentStatus?: string;
      severity?: string;
      service?: string;
      team?: string;
      eventSource?: string;
      visibility?: string;
      eventKind?: string;
    };
    const metadataItems = [];

    if (configuration?.eventKind) {
      metadataItems.push({ icon: "tag", label: `Kind: ${configuration.eventKind}` });
    }
    if (configuration?.visibility) {
      metadataItems.push({ icon: "eye", label: `Visibility: ${configuration.visibility}` });
    }
    if (configuration?.incidentStatus) {
      metadataItems.push({ icon: "activity", label: `Status: ${configuration.incidentStatus}` });
    }
    if (configuration?.severity) {
      metadataItems.push({ icon: "alert-triangle", label: `Severity: ${configuration.severity}` });
    }
    if (configuration?.service) {
      metadataItems.push({ icon: "server", label: `Service: ${configuration.service}` });
    }
    if (configuration?.team) {
      metadataItems.push({ icon: "users", label: `Team: ${configuration.team}` });
    }

    const props: TriggerProps = {
      title: node.name!,
      iconSrc: rootlyIcon,
      collapsedBackground: getBackgroundColorClass(definition.color),
      metadata: metadataItems,
    };

    if (lastEvent) {
      const eventData = lastEvent.data as OnEventEventData;
      const content = eventData?.event
        ? eventData.event.length > 60
          ? eventData.event.substring(0, 60) + "..."
          : eventData.event
        : "Incident Event";
      const parts = [eventData?.kind, eventData?.user_display_name].filter(Boolean).join(" · ");
      const timeAgo = formatTimeAgo(new Date(lastEvent.createdAt));
      const subtitle = [parts, timeAgo].filter(Boolean).join(" · ");

      props.lastEventData = {
        title: content,
        subtitle,
        receivedAt: new Date(lastEvent.createdAt),
        state: "triggered",
        eventId: lastEvent.id,
      };
    }

    return props;
  },
};
