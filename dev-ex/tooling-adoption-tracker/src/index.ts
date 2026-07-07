import { ingestEvents } from "./ingest/eventReceiver.js";
import { computeFunnel } from "./analytics/funnelCalculator.js";

const events = ingestEvents();
const funnel = computeFunnel(events);

console.log("tooling-adoption-funnel", funnel);
