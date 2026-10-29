// @vitest-environment node
import { beforeEach, describe, expect, it, vi } from "vitest";

const { values } = vi.hoisted(() => ({ values: new Map<string, string>() }));
vi.mock("../../platform/storage", () => ({
  defaultStorage: {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => { values.set(key, value); },
    removeItem: (key: string) => { values.delete(key); },
  },
}));

import { useIssueBodyWidthStore } from "./body-width-store";

describe("issue body width preference", () => {
  beforeEach(() => {
    values.clear();
    useIssueBodyWidthStore.setState({ width: "normal" });
  });

  it("defaults to normal and restores a saved choice after rehydration", async () => {
    expect(useIssueBodyWidthStore.getState().width).toBe("normal");
    useIssueBodyWidthStore.getState().setWidth("wide");
    const saved = values.get("multica_issue_body_width")!;
    expect(JSON.parse(saved).state).toEqual({ width: "wide" });

    useIssueBodyWidthStore.setState({ width: "normal" });
    values.set("multica_issue_body_width", saved);
    await useIssueBodyWidthStore.persist.rehydrate();
    expect(useIssueBodyWidthStore.getState().width).toBe("wide");
  });

  it.each(["invalid", 3, null, {}, undefined])("falls back to normal for a stale value: %j", async (width) => {
    values.set("multica_issue_body_width", JSON.stringify({ state: { width }, version: 0 }));
    await useIssueBodyWidthStore.persist.rehydrate();
    expect(useIssueBodyWidthStore.getState().width).toBe("normal");
    expect(useIssueBodyWidthStore.getState().setWidth).toBeTypeOf("function");
  });
});
