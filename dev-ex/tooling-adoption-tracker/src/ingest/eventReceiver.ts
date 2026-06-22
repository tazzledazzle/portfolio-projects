export type ToolingEvent = {
  team: string;
  tool: string;
  stage: "view" | "install" | "run";
};

export function ingestEvents(): ToolingEvent[] {
  return [
    { team: "platform", tool: "ide-plugin", stage: "view" },
    { team: "platform", tool: "ide-plugin", stage: "install" },
    { team: "platform", tool: "ide-plugin", stage: "run" }
  ];
}
