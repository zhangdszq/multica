"use client";

import { useState, type ReactElement } from "react";
import type { Issue } from "@multica/core/types";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
} from "@multica/ui/components/ui/dropdown-menu";
import { ImageIcon as Images, ArrowLeftRight as MoveHorizontal } from "@multica/ui/icons";
import {
  ISSUE_BODY_WIDTH_OPTIONS,
  isIssueBodyWidth,
  useIssueBodyWidthStore,
  ISSUE_IMAGE_COLUMN_OPTIONS,
  isIssueImageColumns,
  useIssueImageLayoutStore,
} from "@multica/core/issues/stores";
import { useIssueActions } from "./use-issue-actions";
import {
  IssueActionsMenuItems,
  dropdownPrimitives,
} from "./issue-actions-menu-items";
import { useT } from "../../i18n";

interface IssueActionsDropdownProps {
  issue: Issue;
  /** A single React element cloned by Base UI as the trigger (via `render` prop). */
  trigger: ReactElement;
  align?: "start" | "end" | "center";
  /** If set, leave the page after the issue is deleted — back to wherever the
   *  user came from, or to this path when there is no in-app history. */
  onDeletedFallbackPath?: string;
}

export function IssueActionsDropdown({
  issue,
  trigger,
  align = "end",
  onDeletedFallbackPath,
}: IssueActionsDropdownProps) {
  const { t } = useT("issues");
  const actions = useIssueActions(issue);
  const [open, setOpen] = useState(false);
  const bodyWidth = useIssueBodyWidthStore((state) => state.width);
  const setBodyWidth = useIssueBodyWidthStore((state) => state.setWidth);
  const bodyWidthLabels = {
    narrow: t(($) => $.actions.body_width_narrow),
    normal: t(($) => $.actions.body_width_normal),
    wide: t(($) => $.actions.body_width_wide),
  };
  const imageColumns = useIssueImageLayoutStore((state) => state.columns);
  const setImageColumns = useIssueImageLayoutStore((state) => state.setColumns);
  const imageColumnLabels = {
    "1": t(($) => $.actions.image_layout_one_column),
    "2": t(($) => $.actions.image_layout_two_columns),
    "3": t(($) => $.actions.image_layout_three_columns),
  } as const;

  // The outer `relative inline-flex` is the picker's anchor box: the
  // absolute, pointer-events-none span inside `triggerRender` fills it, so
  // the popover positions itself relative to the dropdown's 3-dot button
  // without us having to thread a ref through Base UI's anchor API.
  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger render={trigger} />
        <DropdownMenuContent align={align} className="w-auto">
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              <MoveHorizontal className="h-3.5 w-3.5" />
              {t(($) => $.actions.body_width)}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              <DropdownMenuRadioGroup
                value={bodyWidth}
                onValueChange={(value) => {
                  if (isIssueBodyWidth(value)) setBodyWidth(value);
                }}
              >
                {ISSUE_BODY_WIDTH_OPTIONS.map((width) => (
                  <DropdownMenuRadioItem key={width} value={width}>
                    {bodyWidthLabels[width]}
                  </DropdownMenuRadioItem>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuSubContent>
          </DropdownMenuSub>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>
              <Images className="h-3.5 w-3.5" />
              {t(($) => $.actions.image_layout)}
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
              <DropdownMenuRadioGroup
                value={imageColumns}
                onValueChange={(value) => {
                  if (isIssueImageColumns(value)) setImageColumns(value);
                }}
              >
                {ISSUE_IMAGE_COLUMN_OPTIONS.map((columns) => (
                  <DropdownMenuRadioItem key={columns} value={columns}>
                    {columns === "auto"
                      ? t(($) => $.actions.image_layout_auto)
                      : imageColumnLabels[columns]}
                  </DropdownMenuRadioItem>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuSubContent>
          </DropdownMenuSub>
          <IssueActionsMenuItems
            issue={issue}
            actions={actions}
            primitives={dropdownPrimitives}
            onClose={() => setOpen(false)}
            onDeletedFallbackPath={onDeletedFallbackPath}
          />
        </DropdownMenuContent>
    </DropdownMenu>
  );
}
