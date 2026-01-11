# sqlitedb-generator

## 開発環境

- Go:mise でバージョンを固定しています。バージョンは`.mise.toml` を参照してください。
- タスク実行/ツール管理: [mise](https://mise.jdx.dev/)

### セットアップ（mise）

1. mise をインストール
	- 例: `curl https://mise.run | sh`
	- インストール後、シェル統合（`mise activate`）を有効化してください（手順は mise の案内に従ってください）
2. ツールをインストール
	- `mise install`

### タスク実行

このリポジトリでは、開発用コマンドを mise のタスクとして実行します。
タスク定義は `.mise.toml` を参照してください。

## 環境変数

設定は環境変数で行います。ローカル開発では `.env.example` をコピーして `.env` を作り、必要な値に変更してください。

## コードスタイル（EditorConfig）

このリポジトリは [EditorConfig](https://editorconfig.org/) を利用して、エディタ／IDE間での基本的な書式（改行コード、インデント、末尾改行、行末空白など）を統一します。

- 設定ファイル: `.editorconfig`
- 目的: 開発環境に依存せず、差分のノイズを減らしてレビューしやすくする

### 使い方

- VS Code: 拡張機能 `EditorConfig for VS Code` を入れると自動で反映されます
- JetBrains / IntelliJ 系: 多くの製品で標準対応しています（未対応の場合は EditorConfig プラグインを追加してください）

設定の詳細は `.editorconfig` を参照してください。
