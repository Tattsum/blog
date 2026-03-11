import { describe, expect, it } from "vitest";
import { getUnpublishFlagForPublishButton } from "./publish-utils";

describe("getUnpublishFlagForPublishButton", () => {
  it("下書きのとき「公開する」で unpublish: false を渡す", () => {
    expect(getUnpublishFlagForPublishButton(false)).toBe(false);
  });

  it("公開済みのとき「下書きに戻す」で unpublish: true を渡す", () => {
    expect(getUnpublishFlagForPublishButton(true)).toBe(true);
  });
});
