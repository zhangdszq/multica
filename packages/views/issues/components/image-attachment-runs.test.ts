// @vitest-environment node
import { describe, expect, it } from "vitest";
import type { Attachment } from "@multica/core/types";
import { groupAttachmentRuns } from "./image-attachment-runs";

function attachment(id: string, filename: string, contentType: string): Attachment {
  return {
    id,
    filename,
    content_type: contentType,
    size_bytes: 1,
    url: `/uploads/${filename}`,
  } as Attachment;
}

describe("groupAttachmentRuns", () => {
  it("groups adjacent images without reordering files between them", () => {
    const runs = groupAttachmentRuns([
      attachment("a", "a.png", "image/png"),
      attachment("b", "b.jpg", "image/jpeg"),
      attachment("pdf", "report.pdf", "application/pdf"),
      attachment("c", "c.webp", "image/webp"),
    ]);

    expect(runs.map((run) => run.kind)).toEqual(["images", "file", "images"]);
    expect(runs[0]?.kind === "images" && runs[0].attachments.map((a) => a.id)).toEqual(["a", "b"]);
    expect(runs[1]?.kind === "file" && runs[1].attachment.id).toBe("pdf");
    expect(runs[2]?.kind === "images" && runs[2].attachments.map((a) => a.id)).toEqual(["c"]);
  });

  it("uses the shared extension fallback when content type is absent", () => {
    const runs = groupAttachmentRuns([
      attachment("svg", "diagram.svg", ""),
    ]);
    expect(runs[0]?.kind).toBe("images");
  });
});
