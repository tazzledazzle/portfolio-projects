export function computeFunnel(events) {
    return {
        views: events.filter((event) => event.stage === "view").length,
        installs: events.filter((event) => event.stage === "install").length,
        runs: events.filter((event) => event.stage === "run").length
    };
}
