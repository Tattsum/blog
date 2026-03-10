# AI モデル・プロバイダ拡張（設計メモ）

**目的**: 現状は Vertex 上の **Gemini のみ** を `google.golang.org/genai` で呼び出している。ここを **Vertex AI 経由の他モデル** および **必要に応じた外部 API** も選べるようにするための方針を先にドキュメント化する（実装は後続）。

**関連**: [implementation-plan.md](implementation-plan.md) フェーズ 3（AI）、[setup-deploy-checklist.md](setup-deploy-checklist.md) §6.3 Vertex AI。

---

## 1. 現状

| 項目 | 内容 |
| --- | --- |
| 実装箇所 | `backend/internal/infrastructure/vertexai`、`rpc.AIServer` |
| 利用 SDK | `google.golang.org/genai`（`BackendVertexAI`） |
| モデル指定 | 環境変数 `VERTEX_GEMINI_MODEL`（既定 `gemini-2.0-flash-001` 等） |
| フォールバック | `GOOGLE_CLOUD_PROJECT` 未設定時はローカル要約／プレースホルダ |

---

## 2. ゴール（要望の整理）

- **Gemini 以外も選択したい**（同一プロバイダ内のモデル切替に加え、**Claude / OpenAI（ChatGPT）/ DeepSeek 等**も扱いたい）。
- **可能な限り Vertex AI 経由**に寄せると、課金・ネットワーク（出口 IP）・IAM・監査ログを GCP 側に集約できる。
- Vertex で扱えない／運用で直 API がよい場合は **抽象インターフェースの裏側に別アダプタ**を差し込めるようにする。

---

## 3. Vertex AI でのモデル種別（公式の枠組み）

Google Cloud のドキュメント上、大きく次のような提供形態がある（時期によりモデル名・利用可否が変わるため、実装時は Model Garden / 公式一覧で確認すること）。

1. **第一党モデル（Gemini）**  
   - 現行どおり genai SDK や Generative AI API で利用。

2. **パートナーモデル（MaaS）**  
   - 例: **Anthropic Claude**（Claude Sonnet / Opus / Haiku 系）を Vertex 上でマネージド API として利用。  
   - 参照: [Vertex AI partner models](https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/use-partner-models)、[Claude on Model Garden](https://cloud.google.com/products/model-garden/claude)。

3. **オープンモデル（MaaS）**  
   - 例: **DeepSeek**（R1 / V3 系 等）、Llama、Mistral 等を Vertex 上でマネージド利用。  
   - 参照: [DeepSeek on Vertex AI](https://cloud.google.com/vertex-ai/generative-ai/docs/maas/deepseek)、[Open models for MaaS](https://cloud.google.com/vertex-ai/generative-ai/docs/maas/use-open-models)。

4. **その他**  
   - ドキュメント上、一部オープン重量モデルや OSS 系が Model Garden に追加される前提で記載がある。実装時は **利用可能エンドポイントと認証方式**（OAuth / API キー / ADC）を都度確認する。

**ChatGPT（OpenAI 本家 API）** は Vertex の「パートナー」としてではなく、**OpenAI API 直** または **Azure OpenAI** など別経路になることが多い。Vertex 内で OpenAI ブランドの商品として提供される場合と、MaaS の gpt-oss 等のオープン系が混在するため、**「Vertex 経由で使えるもの」と「直 API 必須のもの」を区別**して設計する。

---

## 4. 推奨アーキテクチャ（実装前の方針）

### 4.1 ドメイン／アプリケーション層

- **`TextGenerator` インターフェースを単一化**（既存の `vertexGenerator` を拡張）。  
  - メソッド例: `Generate(ctx, req GenerateRequest) (GenerateResponse, error)`  
  - `GenerateRequest` に **プロバイダ ID・モデル ID・プロンプト・温度等**を含め、RPC 層はこの IF のみに依存する。

### 4.2 プロバイダ別アダプタ

| アダプタ | 役割 | 認証・設定 |
| --- | --- | --- |
| `vertex/gemini` | 現行 genai（Gemini） | ADC / `GOOGLE_CLOUD_PROJECT` + Location |
| `vertex/partner` または `vertex/maas` | Claude / DeepSeek 等（Vertex の Partner / MaaS API） | 公式 SDK または REST。リージョン・モデル ID は Vertex ドキュメント準拠 |
| `openai` | OpenAI API（ChatGPT 等） | `OPENAI_API_KEY` を Secret Manager に格納。プロンプトはログに残さない |
| `anthropic` | Anthropic API 直（Vertex 経由でなくす場合） | `ANTHROPIC_API_KEY` 等を Secret Manager に |
| `deepseek` | DeepSeek API 直（Vertex 経由でない場合） | 同様にキー管理 |

- **同一プロバイダでも「Vertex 経由」と「直 API」は別アダプタ**にすると、課金主体とレート制限が切り分けやすい。

### 4.3 設定の持ち方（段階的）

1. **Phase A（まず実装しやすい）**  
   - 環境変数または Secret Manager で  
     `AI_PROVIDER=vertex-gemini`  
     `AI_MODEL=gemini-2.0-flash-001`  
     のように **単一プロバイダ・単一モデル**。

2. **Phase B**  
   - 上記に加え、RPC のリクエストに **オプションで model_override** を付けない（管理者のみ API キーで触る想定なら、サーバ側設定のみでも可）。

3. **Phase C（将来）**  
   - DB に「既定プロバイダ／モデル」を保存し、管理画面から切替。監査ログと合わせて検討。

### 4.4 セキュリティ

- **API キーは Secret Manager**。Cloud Run からは env の `value_source.secret_key_ref` で注入。
- **プロンプト・レスポンス本文をアプリログに出さない**（デバッグ時もマスクまたはサンプリングのみ）。
- **レート制限・入力長上限**はプロバイダ共通でミドルウェア化（既存の truncate を拡張）。

### 4.5 エラーとフォールバック

- 外部 API 障害時は **`CodeInternal` + 汎用メッセージ**（既存 `MapHandlerError` と整合）。
- 任意で **フォールバックチェーン**（例: Vertex Claude 失敗 → Gemini）を設定可能にするかはコストと一貫性のトレードオフで ADR 化する。

---

## 5. 実装タスク（チェックリスト）

- [ ] `TextGenerator` IF と `GenerateRequest` / `GenerateResponse` の定義（`internal/domain` または `internal/application`）
- [ ] 現行 `vertexai.Client` を `vertex/gemini` アダプタとしてラップ
- [ ] Vertex Partner / MaaS 用クライアント調査（公式 Go SDK の有無、REST の場合の署名）
- [ ] OpenAI / Anthropic 直用の薄いクライアント（http.Client + タイムアウト）
- [ ] 設定読み込み（env / tfvars の方針）
- [ ] proto に `model_id` 等を追加するかは UI 要件次第（まずはサーバ設定のみでも可）
- [ ] 手順書更新（どの IAM / どの API 有効化が必要かを setup-deploy-checklist に追記）

---

## 6. 参照リンク（2026 年 3 月時点の出発点）

- [Vertex AI partner models](https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/use-partner-models)
- [DeepSeek models on Vertex AI](https://cloud.google.com/vertex-ai/generative-ai/docs/maas/deepseek)
- [Open models MaaS](https://cloud.google.com/vertex-ai/generative-ai/docs/maas/use-open-models)
- [Google Gen AI Go SDK (genai)](https://pkg.go.dev/google.golang.org/genai) — Gemini 向け。パートナーモデルは別クライアントの可能性あり。

---

このドキュメントは **実装の前段**として置く。実装時は ADR を 1 本追加し、最初に **Vertex 上の 1 パートナーモデル（例: Claude または DeepSeek）を 1 本だけ**繋いでから、OpenAI 直などを増やすと安全。
