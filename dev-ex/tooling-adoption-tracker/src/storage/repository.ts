import type { ToolingEvent } from "../ingest/eventReceiver.js";

export class EventRepository {
  private readonly buffer: ToolingEvent[] = [];

  insert(event: ToolingEvent): void {
    this.buffer.push(event);
  }

  list(): ToolingEvent[] {
    return this.buffer;
  }
}
