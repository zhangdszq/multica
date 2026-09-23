import { isImageAttachment } from "@multica/core/attachments/image-sequence";
import type { Attachment } from "@multica/core/types";

export type AttachmentRun =
  | { kind: "images"; attachments: Attachment[] }
  | { kind: "file"; attachment: Attachment };

/** Preserve attachment order while grouping only adjacent images. */
export function groupAttachmentRuns(attachments: Attachment[]): AttachmentRun[] {
  const runs: AttachmentRun[] = [];

  for (const attachment of attachments) {
    if (isImageAttachment(attachment.content_type, attachment.filename)) {
      const previous = runs.at(-1);
      if (previous?.kind === "images") {
        previous.attachments.push(attachment);
      } else {
        runs.push({ kind: "images", attachments: [attachment] });
      }
      continue;
    }
    runs.push({ kind: "file", attachment });
  }

  return runs;
}
