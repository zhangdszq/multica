import type { Element, ElementContent, Root, RootContent } from "hast";

const GALLERY_DATA_TYPE = "imageGallery";

function imagesFromParagraph(node: RootContent): Element[] | null {
  if (node.type !== "element" || node.tagName !== "p") return null;

  const images: Element[] = [];
  for (const child of node.children) {
    if (child.type === "text" && child.value.trim() === "") continue;
    if (child.type === "element" && child.tagName === "br") continue;
    if (child.type === "element" && child.tagName === "img") {
      images.push(child);
      continue;
    }
    return null;
  }
  return images.length > 0 ? images : null;
}

function galleryNode(images: Element[]): Element {
  return {
    type: "element",
    tagName: "div",
    properties: { dataType: GALLERY_DATA_TYPE },
    children: images as ElementContent[],
  };
}

/**
 * Group only adjacent, top-level image-only paragraphs.
 *
 * Lists, quotes, tables, links and mixed prose retain their authored document
 * flow. The marker already passes through the canonical sanitize schema's
 * `dataType` allow-list and is converted to real gallery chrome by RichContent.
 */
export function rehypeImageGallery() {
  return (tree: Root) => {
    const next: RootContent[] = [];
    let pending: Element[] = [];

    const flush = () => {
      if (pending.length === 0) return;
      next.push(galleryNode(pending));
      pending = [];
    };

    for (const node of tree.children) {
      const images = imagesFromParagraph(node);
      if (images) {
        pending.push(...images);
        continue;
      }
      // Markdown block separation survives into HAST as whitespace text
      // between sibling paragraphs. It is layout syntax, not authored prose,
      // so it must not split an otherwise consecutive image run.
      if (pending.length > 0 && node.type === "text" && node.value.trim() === "") {
        continue;
      }
      flush();
      next.push(node);
    }
    flush();
    tree.children = next;
  };
}
