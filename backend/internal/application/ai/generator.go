// Package ai はテキスト生成（要約・下書き支援等）のアプリケーション層ポートを定義する。
// 実装は infrastructure（Vertex Gemini / 将来 Partner MaaS / OpenAI 直 等）に置く。
package ai

import "context"

// TextGenerator はプロンプトからテキストを生成する共通ポート。
// Summarize / DraftSupport は RPC 層でプロンプトを組み立て、この IF に渡す。
// nil のときは RPC 層がローカル要約／プレースホルダにフォールバックする。
type TextGenerator interface {
	// GenerateText は単一プロンプトで生成テキストを返す。空文字のみ返してよい。
	GenerateText(ctx context.Context, prompt string) (string, error)
}
