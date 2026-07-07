import { describe, it } from "node:test";
import assert from "node:assert/strict";

import { computeFunnel } from "../src/analytics/funnelCalculator.js";

describe("computeFunnel", () => {
  it("counts each funnel stage", () => {
    const result = computeFunnel([
      { team: "a", tool: "cli", stage: "view" },
      { team: "a", tool: "cli", stage: "install" },
      { team: "a", tool: "cli", stage: "run" }
    ]);

    assert.equal(result.views, 1);
    assert.equal(result.installs, 1);
    assert.equal(result.runs, 1);
  });
});
