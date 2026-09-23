import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";
import type { HTMLAttributes } from "react";
import type { NodeViewProps } from "@tiptap/react";

vi.mock("@tiptap/react", () => ({
  NodeViewWrapper: ({ children, ...props }: HTMLAttributes<HTMLDivElement>) => (
    <div {...props}>{children}</div>
  ),
}));

vi.mock("../attachment", () => ({
  Attachment: () => <span className="image-node" />,
}));

import { ImageView } from "./image-view";

describe("ImageView", () => {
  it("marks the outer ProseMirror node so issue descriptions can place it in the image grid", () => {
    const props = {
      node: { attrs: { src: "https://example.test/image.png", alt: "" } },
      editor: { isEditable: true },
      selected: false,
      deleteNode: vi.fn(),
    } as unknown as NodeViewProps;
    const { container } = render(<ImageView {...props} />);

    expect(container.firstElementChild).toHaveClass("image-node-view");
    expect(container.querySelector(".image-node-view > .image-node")).not.toBeNull();
  });
});
