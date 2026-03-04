# ADR-002: フロントを Cloudflare Pages、API を GCP Cloud Run でホスティングする

## ステータス

Accepted

## コンテキスト

個人ブログのフロントエンド（Next.js）とバックエンド（Go + connect-go）をどこにデプロイするか検討する必要があった。モノレポであり、フロントと API は別サービスとしてスケール・コスト・運用を分離したい。技術スタックは Go / connect-go / GCP Cloud Run / Cloud SQL (MySQL) / Vertex AI / Cloudflare Pages を前提とする。

## 決定

**フロントエンドは Cloudflare Pages に、API（connect-go）は GCP Cloud Run にデプロイする。** 同一の Vercel や Cloud Run のみに寄せるのではなく、フロントと API でプロバイダを分ける。

## 選択肢の比較

| 選択肢 | メリット | デメリット |
| --- | --- | --- |
| **Cloudflare Pages + Cloud Run** | フロントはエッジ配信でレイテンシ・キャッシュに強い。API は GCP 内で Cloud SQL・Vertex AI と同一ネットワークにでき、Secret Manager 等と連携しやすい。 | プロバイダが二つになり、認証・ドメイン・監視の設定が分散する。 |
| **Vercel にフロント・API を集約** | 単一プロバイダで Next.js の DX が良い。Serverless Functions で API も載せられる。 | API を Go で書く場合、Vercel の Go ランタイムは connect-go の運用事例が少ない。Cloud SQL・Vertex AI は GCP 外となり、接続・認証の設計が重くなる。 |
| **Cloud Run にフロント・API を両方** | GCP 一本で、Cloud SQL・Vertex AI と近い。 | 静的アセットを Cloud Run で配信するとコスト・レイテンシの面でエッジ CDN に劣る。フロントのキャッシュ戦略を自前で設計する必要がある。 |
| **Cloudflare Pages + Cloudflare Workers (API)** | フロント・API を同一プロバイダにまとめられる。エッジで API も動かせる。 | Workers で Go は動かせず、API は JS/TS か他のランタイムになる。Cloud SQL・Vertex AI との連携は GCP 側に別サービスが必要で、構成が複雑化しうる。 |

## 理由

- **API は GCP 内に置きたい**: Cloud SQL (MySQL) と Vertex AI (Gemini) をすでに採用しており、Cloud Run 上に API を置けば VPC コネクタでプライベート接続し、Secret Manager で認証情報を渡せる。同一リージョン・プロジェクトにまとめることでレイテンシと運用の一貫性を確保した。
- **フロントはエッジで配信したい**: Next.js のビルド成果物は静的または SSR であり、Cloudflare Pages でエッジキャッシュ・HTTPS・カスタムドメインをそのまま利用できる。読者向けトラフィックは CDN で賄い、API は必要なときだけ Cloud Run に飛ばす形にした。
- **トレードオフ**: プロバイダが GCP と Cloudflare の二つになるため、ドメイン設定（例: 同一ドメインで /api を Cloud Run にプロキシする等）や、監視・ログの集約は設計時に意識する必要がある。その代わり、フロントの配信と API・DB・AI の処理を役割ごとに最適な場所に置けた。

## 影響・結果

### ポジティブ

- フロントの配信は Cloudflare のグローバルエッジでキャッシュされ、読者体験とコストのバランスが取りやすい。
- API は GCP 内で Cloud SQL・Vertex AI と密結合でき、セキュリティとパフォーマンスの見通しが良い。
- フロントと API を独立してスケール・デプロイでき、障害の影響範囲を分離しやすい。

### ネガティブ

- 二つのクラウド／CDN の契約・請求・権限管理が必要になる。
- CORS や API のベース URL（本番・プレビュー）の管理をフロントの環境変数等で明示する必要がある。
- 将来、フロントを GCP に寄せる（例: Cloud Run + Cloud CDN）などに変更する場合、デプロイパイプラインとドキュメントの更新が発生する。

## 参考

- [Cloudflare Pages - Documentation](https://developers.cloudflare.com/pages/)
- [Cloud Run - ドキュメント](https://cloud.google.com/run/docs)
- [アーキテクチャ設計書 - インフラ構成](../architecture.md#5-インフラ構成)
