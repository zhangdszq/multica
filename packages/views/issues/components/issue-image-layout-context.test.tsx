import { beforeEach, describe, expect, it } from "vitest";
import { render } from "@testing-library/react";
import { useIssueImageLayoutStore } from "@multica/core/issues/stores";
import {
  IssueDescriptionImageLayout,
  IssueImageGallery,
  IssueImageLayoutProvider,
} from "./issue-image-layout-context";

describe("issue image layout scopes", () => {
  beforeEach(() => {
    useIssueImageLayoutStore.setState({ columns: "auto" });
  });

  it.each(["2", "3"] as const)(
    "keeps a single description image in the explicit %s-column scope",
    (columns) => {
      useIssueImageLayoutStore.setState({ columns });

      const { container } = render(
        <IssueImageLayoutProvider>
          <IssueDescriptionImageLayout>
            <div data-testid="single-image" />
          </IssueDescriptionImageLayout>
        </IssueImageLayoutProvider>,
      );

      const scope = container.querySelector(".issue-description-image-layout");
      expect(scope).toHaveAttribute("data-image-columns", columns);
      expect(scope?.children).toHaveLength(1);
    },
  );

  it.each(["2", "3"] as const)(
    "keeps a single comment image in the explicit %s-column scope",
    (columns) => {
      useIssueImageLayoutStore.setState({ columns });

      const { container } = render(
        <IssueImageLayoutProvider>
          <IssueImageGallery>
            <div data-testid="single-image" />
          </IssueImageGallery>
        </IssueImageLayoutProvider>,
      );

      const scope = container.querySelector(".issue-image-gallery-shell");
      expect(scope).toHaveAttribute("data-image-columns", columns);
      expect(container.querySelectorAll(".issue-image-gallery > *")).toHaveLength(1);
    },
  );
});
