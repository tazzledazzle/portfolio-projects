import type { ToolingEvent } from "../ingest/eventReceiver.js";

export function computeFunnel(events: ToolingEvent[]): Record<string, number> {
  return {
    views: events.filter((event) => event.stage === "view").length,
    installs: events.filter((event) => event.stage === "install").length,
    runs: events.filter((event) => event.stage === "run").length
  };
}
