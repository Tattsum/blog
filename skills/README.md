# skills（blog モノレポ）

このリポジトリ用の **Agent Skill ソース**。編集の正本は `src/` 以下。

## 方針

- **`src/<skill-id>/SKILL.md`** … 各 skill の本体。
- **設置** … `make install` で `.cursor/skills/<name>/` にコピーし、Cursor が読み込む。生成先は `.gitignore` 済み（正本は `src/` のみコミット）。

## レイアウト

```text
skills/
├── README.md
├── Makefile
└── src/
    ├── blog-lint-and-test/SKILL.md
    └── blog-handover/SKILL.md
```

## コマンド

- **`make -C skills list`** … `src/` 以下の skill 一覧
- **`make -C skills install`** … `src/<name>/` を `.cursor/skills/<name>/` にコピー
