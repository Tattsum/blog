# backend/internal

このディレクトリには、ドメイン駆動設計のレイヤごとのパッケージを配置する。

- `domain/`: エンティティ・値オブジェクト・ドメインサービス
- `application/`: ユースケース（サービス）レイヤ
- `infrastructure/`: DB・外部サービスとの実装
- `interface/`: transport 層（connect-go ハンドラなど）
