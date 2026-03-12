# skills（blog モノレポ）

このリポジトリ用の **Cursor Agent Skill**。編集の正本は `src/` 以下。

## 方針

- **`src/<skill-id>/SKILL.md`** … 各 skill の本体。ここを編集する。
- **Cursor での利用** … `make install` で `.cursor/skills/<name>/` にコピーする。`.cursor/skills/` はリポジトリにコミットしてあり、Cursor がプロジェクトを開いたときに自動で読み込む。
- **更新手順** … `src/` を編集 → リポジトリルートで `make -C skills install` → `.cursor/skills/` の変更をコミット。

## レイアウト

```text
skills/
├── README.md
├── Makefile
└── src/
    ├── blog-handover/SKILL.md
    ├── blog-lint-and-test/SKILL.md
    └── blog-security-performance/SKILL.md
```

展開先（コミット済み）:

```text
.cursor/skills/
├── blog-handover/SKILL.md
├── blog-lint-and-test/SKILL.md
└── blog-security-performance/SKILL.md
```

## コマンド

- **`make -C skills list`** … `src/` 以下の skill 一覧
- **`make -C skills install`** … `src/<name>/` を `.cursor/skills/<name>/` にコピー（編集後に実行してからコミット）
