import {
  ComponentBaseContext,
  ComponentBaseMapper,
  ExecutionDetailsContext,
  ExecutionInfo,
  NodeInfo,
  OutputPayload,
  SubtitleContext,
} from "../types";
import { ComponentBaseProps, ComponentBaseSpec, EventSection } from "@/ui/componentBase";
import { getBackgroundColorClass, getColorClass } from "@/utils/colors";
import { getState, getStateMap, getTriggerRenderer } from "..";
import { MetadataItem } from "@/ui/metadataList";
import slackIcon from "@/assets/icons/integrations/slack.svg";
import { formatTimeAgo } from "@/utils/date";

interface SendAndWaitForResponseConfiguration {
  channel?: string;
  message?: string;
  timeout?: number;
  buttons?: Array<{
    name: string;
    value: string;
  }>;
}

interface SendAndWaitForResponseMetadata {
  channel?: {
    id?: string;
    name?: string;
  };
  messageTs?: string;
  responseTs?: string;
  selectedValue?: string;
  state?: string;
}

export const sendAndWaitForResponseMapper: ComponentBaseMapper = {
  props(context: ComponentBaseContext): ComponentBaseProps {
    const lastExecution = context.lastExecutions.length > 0 ? context.lastExecutions[0] : null;
    const componentName = context.componentDefinition.name || "unknown";

    return {
      title:
        context.node.name ||
        context.componentDefinition.label ||
        context.componentDefinition.name ||
        "Unnamed component",
      iconSrc: slackIcon,
      iconSlug: "slack",
      iconColor: getColorClass(context.componentDefinition.color),
      collapsedBackground: getBackgroundColorClass(context.componentDefinition.color),
      collapsed: context.node.isCollapsed,
      eventSections: lastExecution
        ? sendAndWaitForResponseEventSections(context.nodes, lastExecution, componentName)
        : undefined,
      includeEmptyState: !lastExecution,
      metadata: sendAndWaitForResponseMetadataList(context.node),
      specs: sendAndWaitForResponseSpecs(context.node),
      eventStateMap: getStateMap(componentName),
    };
  },

  getExecutionDetails(context: ExecutionDetailsContext): Record<string, string> {
    const metadata = context.execution.metadata as SendAndWaitForResponseMetadata | undefined;
    const outputs = context.execution.outputs as
      | { received?: OutputPayload[]; timeout?: OutputPayload[] }
      | undefined;

    // Check which output channel was used
    const receivedData = outputs?.received?.[0]?.data as Record<string, unknown> | undefined;
    const timeoutData = outputs?.timeout?.[0]?.data as Record<string, unknown> | undefined;

    const details: Record<string, string> = {
      Channel: metadata?.channel?.name || "-",
      State: metadata?.state || "-",
    };

    if (receivedData) {
      details["Selected Value"] = stringOrDash(receivedData.value);
      details["Response Time"] = formatSlackTimestamp(receivedData.responseTs) || "-";
      details["User"] = stringOrDash((receivedData.user as Record<string, unknown>)?.name);
    } else if (timeoutData) {
      details.Status = "Timeout (no response received)";
    }

    return details;
  },

  subtitle(context: SubtitleContext): string {
    if (!context.execution.createdAt) return "";
    const metadata = context.execution.metadata as SendAndWaitForResponseMetadata | undefined;

    if (metadata?.state === "waiting") {
      return "Waiting for response...";
    }

    if (metadata?.state === "responded") {
      return `Responded: ${metadata.selectedValue}`;
    }

    if (metadata?.state === "timeout") {
      return "Timed out";
    }

    return formatTimeAgo(new Date(context.execution.createdAt));
  },
};

function sendAndWaitForResponseMetadataList(node: NodeInfo): MetadataItem[] {
  const metadata: MetadataItem[] = [];
  const nodeMetadata = node.metadata as SendAndWaitForResponseMetadata | undefined;
  const configuration = node.configuration as SendAndWaitForResponseConfiguration | undefined;

  const channelLabel = nodeMetadata?.channel?.name || configuration?.channel;
  if (channelLabel) {
    metadata.push({ icon: "hash", label: channelLabel });
  }

  // Show timeout if configured
  if (configuration?.timeout) {
    metadata.push({
      icon: "clock",
      label: `${configuration.timeout}s timeout`,
    });
  }

  return metadata;
}

function sendAndWaitForResponseSpecs(node: NodeInfo): ComponentBaseSpec[] {
  const specs: ComponentBaseSpec[] = [];
  const configuration = node.configuration as SendAndWaitForResponseConfiguration | undefined;

  // Show buttons as badges
  if (configuration?.buttons && configuration.buttons.length > 0) {
    const buttonLabels = configuration.buttons.map((b) => b.name);
    specs.push({
      title: "buttons",
      tooltipTitle: "Buttons",
      iconSlug: "square",
      items: buttonLabels,
    });
  }

  return specs;
}

function sendAndWaitForResponseEventSections(
  nodes: NodeInfo[],
  execution: ExecutionInfo,
  componentName: string,
): EventSection[] {
  const rootTriggerNode = nodes.find((n) => n.id === execution.rootEvent?.nodeId);
  const rootTriggerRenderer = getTriggerRenderer(rootTriggerNode?.componentName!);
  const { title } = rootTriggerRenderer.getTitleAndSubtitle({ event: execution.rootEvent });

  return [
    {
      receivedAt: new Date(execution.createdAt!),
      eventTitle: title,
      eventSubtitle: formatTimeAgo(new Date(execution.createdAt!)),
      eventState: getState(componentName)(execution),
      eventId: execution.rootEvent!.id!,
    },
  ];
}

function stringOrDash(value?: unknown): string {
  if (value === undefined || value === null || value === "") {
    return "-";
  }

  return String(value);
}

function formatSlackTimestamp(value?: unknown): string | undefined {
  if (value === undefined || value === null || value === "") {
    return undefined;
  }

  const raw = String(value);
  const seconds = Number.parseFloat(raw);
  if (!Number.isNaN(seconds)) {
    return new Date(seconds * 1000).toLocaleString();
  }

  const asDate = new Date(raw);
  if (!Number.isNaN(asDate.getTime())) {
    return asDate.toLocaleString();
  }

  return raw;
}
