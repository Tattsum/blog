/**
 * 編集画面の「公開する」「下書きに戻す」ボタンで PublishPost に渡す unpublish の値を決める。
 * - 下書きのとき「公開する」→ unpublish: false
 * - 公開済みのとき「下書きに戻す」→ unpublish: true
 * この対応をテストで担保するためだけに切り出している。
 */
export function getUnpublishFlagForPublishButton(isPublished: boolean): boolean {
  return isPublished;
}
