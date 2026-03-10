# skills（blog モノレポ）

[shuymn/dotfiles の skills](https://github.com/shuymn/dotfiles/tree/main/skills) のレイアウトを参考にした、**このリポジトリ用の Agent Skill ソース**です。

## 方針

- **`src/`** … 編集の正本。各 skill は `src/<skill-id>/SKILL.md` を持つ。
- **ビルドツール** … dotfiles 版の `skit` / `skitkit` は入れていない。必要なら `Makefile` の `install` で `.cursor/skills/` にコピーするだけでよい。
- **設置** … Cursor でプロジェクト skill として使う場合は `.cursor/skills/` に配置する（README の install 参照）。

## レイアウト

```text
skills/
├── README.md           # このファイル
├── Makefile            # list / install など
└── src/
    ├── blog-lint-and-test/   # 完了前の lint・test 必須
    │   └── SKILL.md
    └── blog-handover/        # 引き継ぎ・次の作業
        └── SKILL.md
```

## コマンド

- **`make -C skills list`** … `src/` 以下の skill 一覧
- **`make -C skills install`** … `src/<name>/` を `.cursor/skills/<name>/` にコピー（Cursor が読み込む）

## 注意

- `.cursor/skills/` にインストールした skill は **リポジトリにコミットするか**はチーム方針で決める（個人のみなら .gitignore でも可）。
- dotfiles のように `*.md.tmpl` + fragments でビルドする構成は、必要になったら拡張する。
