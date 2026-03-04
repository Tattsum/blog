# ADR-001: API に Connect RPC（connect-go）を採用する

## ステータス

Accepted

## コンテキスト

個人ブログのフロントエンド（Next.js）とバックエンド（Go）の間で、記事・タグ・認証・AI 連携などの API をどう定義・実装するかが課題だった。REST / GraphQL / gRPC など複数の方式があり、型安全性・開発体験・インフラ（GCP Cloud Run、Cloudflare Pages）との相性・将来の拡張を考慮する必要があった。バックエンドは Go、DB は Cloud SQL (MySQL)、AI は Vertex AI (Gemini)、フロントは Cloudflare Pages で配信する前提である。

## 決定

**API 方式として Connect RPC を採用し、バックエンドは connect-go、フロントは @connectrpc/connect（proto から自動生成クライアント）で実装する。**

Protocol Buffers で API を定義し、Connect プロトコル（HTTP/1.1 ベース、gRPC 互換オプションあり）で通信する。

## 選択肢の比較

| 選択肢 | メリット | デメリット |
| --- | --- | --- |
| **Connect RPC** | proto で型・契約が明確。Go/TS 双方でコード生成。HTTP/JSON 互換でブラウザ・CDN と相性が良い。gRPC との互換オプションあり。 | エコシステムが gRPC より小さい。学習コストが REST より高い。 |
| **REST (OpenAPI)** | 広く普及しており、ツール・ドキュメントが豊富。HTTP の仕組みに素直。 | スキーマ駆動の型生成が Connect より手間。過剰/過少取得の制御が煩雑。 |
| **GraphQL** | クライアントが必要なフィールドだけ取得できる。単一エンドポイント。 | バックエンドは Go の成熟したエコシステムが REST/gRPC より小さい。キャッシュ・レート制限の設計が複雑。 |
| **gRPC (純正)** | 高性能・型安全。Go との相性が良い。 | ブラウザからは gRPC-Web が必要。Cloudflare 等で HTTP/2 やストリーミングの制約が出ることがある。 |

## 理由

- **型安全性と契約の一元化**: proto を単一のソースオブトゥルースにし、connect-go と @connectrpc/connect でサーバ・クライアントを生成することで、API の変更時に型不一致をコンパイル/ビルド段階で検知できる。
- **HTTP/JSON 互換**: Connect の JSON モードにより、Cloudflare Pages 上の Next.js から同一オリジンまたは CORS で Cloud Run の API を呼びやすく、デバッグや既存の HTTP ツールも使いやすい。
- **Go との相性**: connect-go は Go 標準の net/http に載り、Cloud Run のコンテナでそのまま運用できる。gRPC 互換が必要になった場合も移行しやすい。
- **トレードオフ**: REST に比べると「API を JSON で手書きで叩く」という運用は減り、proto とコード生成の利用が前提になる。その代わり、フロント・バック間の契約のずれを減らし、リファクタ時の安心感を優先した。

## 影響・結果

### ポジティブ

- 記事・タグ・認証・AI などの RPC が proto で一覧でき、[docs/api-specification.md](../api-specification.md) と整合した仕様管理がしやすい。
- フロント・バックで同じメッセージ定義を共有でき、リクエスト/レスポンスの型が揃う。
- 将来、他言語のクライアントや別サービスを追加する場合も proto から生成できる。

### ネガティブ

- チームや協力者が REST のみ経験の場合、Connect / proto の理解が必要になる。
- コード生成のパイプライン（buf 等）の導入・維持が必要である。
- 細かい API の「REST らしい URL 設計」や HTTP キャッシュを URL ベースで細かく制御するような要件には、Connect 単体では補いづらい部分がある（必要なら CDN やプロキシで対応する想定）。

## 参考

- [Connect RPC - Go](https://connectrpc.com/docs/go/getting-started)
- [Connect RPC - TypeScript/Connect for web](https://connectrpc.com/docs/web/getting-started)
- [Protocol Buffers - Language Guide](https://protobuf.dev/programming-guides/proto3/)
- [Buf - CLI](https://buf.build/docs/reference/cli/buf/)
