import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";
import { defaultStorage } from "../../platform/storage";

export const ISSUE_IMAGE_COLUMN_OPTIONS = ["auto", "1", "2", "3"] as const;
export type IssueImageColumns = (typeof ISSUE_IMAGE_COLUMN_OPTIONS)[number];

export function isIssueImageColumns(value: unknown): value is IssueImageColumns {
  return ISSUE_IMAGE_COLUMN_OPTIONS.includes(value as IssueImageColumns);
}

interface IssueImageLayoutStore {
  columns: IssueImageColumns;
  setColumns: (columns: IssueImageColumns) => void;
}

/**
 * Personal image-density preference for issue detail pages.
 *
 * The value is a maximum column count: responsive layout may step down on a
 * narrow content panel. It is device-local like the sticky comment composer
 * preference and never changes issue data.
 */
export const useIssueImageLayoutStore = create<IssueImageLayoutStore>()(
  persist(
    (set) => ({
      columns: "auto",
      setColumns: (columns) => set({ columns }),
    }),
    {
      name: "multica_issue_image_layout",
      storage: createJSONStorage(() => defaultStorage),
      merge: (persisted, current) => {
        const candidate = (persisted as Partial<IssueImageLayoutStore> | null)?.columns;
        return {
          ...current,
          columns: isIssueImageColumns(candidate) ? candidate : "auto",
        };
      },
    },
  ),
);
