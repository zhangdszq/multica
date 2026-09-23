// @vitest-environment node
import { describe, expect, it } from "vitest";
import type { Element, Root, RootContent } from "hast";
import { rehypeImageGallery } from "./rehype-image-gallery";

function element(tagName: string, children: Element["children"] = []): Element {
  return { type: "element", tagName, properties: {}, children };
}

function image(src: string): Element {
  return { ...element("img"), properties: { src } };
}

function paragraph(...children: Element["children"]): Element {
  return element("p", children.flat());
}

function apply(children: RootContent[]): Root {
  const tree: Root = { type: "root", children };
  rehypeImageGallery()(tree);
  return tree;
}

describe("rehypeImageGallery", () => {
  it("groups adjacent top-level image-only paragraphs", () => {
    const tree = apply([
      paragraph(image("a.png")),
      { type: "text", value: "\n" },
      paragraph(image("b.png")),
      { type: "text", value: "\n" },
      paragraph({ type: "text", value: "caption" }),
    ]);

    expect(tree.children).toHaveLength(2);
    const gallery = tree.children[0] as Element;
    expect(gallery.tagName).toBe("div");
    expect(gallery.properties.dataType).toBe("imageGallery");
    expect(gallery.children).toHaveLength(2);
  });

  it("leaves mixed prose, nested lists and code untouched", () => {
    const mixed = paragraph(image("a.png"), { type: "text", value: "caption" });
    const list = element("ul", [element("li", [paragraph(image("b.png"))])]);
    const code = element("pre", [element("code", [{ type: "text", value: "![](c.png)" }])]);
    const tree = apply([mixed, list, code]);

    expect(tree.children).toEqual([mixed, list, code]);
  });

  it("accepts line breaks between images in the same paragraph", () => {
    const tree = apply([
      paragraph(image("a.png"), element("br"), { type: "text", value: "\n" }, image("b.png")),
    ]);
    expect((tree.children[0] as Element).children).toHaveLength(2);
  });
});
