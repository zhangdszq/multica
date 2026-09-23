"use client";

import { createContext, useContext, type ReactNode } from "react";
import {
  useIssueImageLayoutStore,
  type IssueImageColumns,
} from "@multica/core/issues/stores";
import { cn } from "@multica/ui/lib/utils";
import "./issue-image-gallery.css";

const IssueImageLayoutContext = createContext<IssueImageColumns | null>(null);

export function IssueImageLayoutProvider({ children }: { children: ReactNode }) {
  const columns = useIssueImageLayoutStore((state) => state.columns);
  return (
    <IssueImageLayoutContext.Provider value={columns}>
      {children}
    </IssueImageLayoutContext.Provider>
  );
}

export function useIssueImageLayout(): IssueImageColumns | null {
  return useContext(IssueImageLayoutContext);
}

export function IssueImageGallery({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const columns = useIssueImageLayout();
  if (columns === null) return <>{children}</>;

  return (
    <div
      className={cn("issue-image-gallery-shell", className)}
      data-image-columns={columns}
    >
      <div className="issue-image-gallery">{children}</div>
    </div>
  );
}

export function IssueDescriptionImageLayout({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const columns = useIssueImageLayout();
  if (columns === null) return <>{children}</>;

  return (
    <div
      className={cn("issue-description-image-layout", className)}
      data-image-columns={columns}
    >
      {children}
    </div>
  );
}
