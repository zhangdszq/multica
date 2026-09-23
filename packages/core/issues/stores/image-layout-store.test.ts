// @vitest-environment jsdom
import { beforeAll, beforeEach, describe, expect, it } from "vitest";
import {
  isIssueImageColumns,
  useIssueImageLayoutStore,
} from "./image-layout-store";

beforeAll(() => {
  if (typeof globalThis.localStorage?.clear === "function") return;
  const values = new Map<string, string>();
  const storage: Storage = {
    get length() { return values.size; },
    clear: () => values.clear(),
    getItem: (key) => values.get(key) ?? null,
    key: (index) => Array.from(values.keys())[index] ?? null,
    removeItem: (key) => { values.delete(key); },
    setItem: (key, value) => { values.set(key, value); },
  };
  Object.defineProperty(globalThis, "localStorage", { configurable: true, value: storage });
  Object.defineProperty(window, "localStorage", { configurable: true, value: storage });
});

describe("issue image layout store", () => {
  beforeEach(() => {
    localStorage.clear();
    useIssueImageLayoutStore.setState({ columns: "auto" });
  });

  it("defaults to automatic layout and accepts every supported column value", () => {
    expect(useIssueImageLayoutStore.getState().columns).toBe("auto");

    for (const columns of ["1", "2", "3", "auto"] as const) {
      useIssueImageLayoutStore.getState().setColumns(columns);
      expect(useIssueImageLayoutStore.getState().columns).toBe(columns);
    }
  });

  it("rejects stale or malformed persisted values", () => {
    expect(isIssueImageColumns("auto")).toBe(true);
    expect(isIssueImageColumns("3")).toBe(true);
    expect(isIssueImageColumns(3)).toBe(false);
    expect(isIssueImageColumns("4")).toBe(false);
  });

  it("persists the personal preference on this device", () => {
    useIssueImageLayoutStore.getState().setColumns("2");

    const saved = JSON.parse(localStorage.getItem("multica_issue_image_layout") ?? "{}");
    expect(saved.state.columns).toBe("2");
  });
});
