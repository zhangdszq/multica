import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";
import { defaultStorage } from "../../platform/storage";

export const ISSUE_BODY_WIDTH_OPTIONS = ["narrow", "normal", "wide"] as const;
export type IssueBodyWidth = (typeof ISSUE_BODY_WIDTH_OPTIONS)[number];

export function isIssueBodyWidth(value: unknown): value is IssueBodyWidth {
  return ISSUE_BODY_WIDTH_OPTIONS.includes(value as IssueBodyWidth);
}

interface IssueBodyWidthStore {
  width: IssueBodyWidth;
  setWidth: (width: IssueBodyWidth) => void;
}

/** Device-local reading preference, shared by Web and Desktop issue views. */
export const useIssueBodyWidthStore = create<IssueBodyWidthStore>()(
  persist(
    (set) => ({
      width: "normal",
      setWidth: (width) => set({ width }),
    }),
    {
      name: "multica_issue_body_width",
      storage: createJSONStorage(() => defaultStorage),
      partialize: ({ width }) => ({ width }),
      merge: (persisted, current) => {
        const width = (persisted as Partial<IssueBodyWidthStore> | null)?.width;
        return { ...current, width: isIssueBodyWidth(width) ? width : "normal" };
      },
    },
  ),
);
