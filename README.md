# sqlitedb-generator

TSVファイル群から SQLite のDBファイルを生成する小さなジェネレータです。

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

## 使い方

`./tsv` 配下の `*.tsv` を読み込み、同名のテーブルに取り込んで `out.db` を作成します。

```sh
go run ./cmd/sqlitedb-generator -in ./tsv -out ./out.db -overwrite -drop -v
```

### オプション

- `-in` : 入力TSVディレクトリ（デフォルト `./tsv`）
- `-out` : 出力DBパス（デフォルト `./out.db`）
- `-overwrite` : 既存のDBファイルを削除して作り直す
- `-drop` : 既存テーブルを `DROP TABLE IF EXISTS` してから作り直す
- `-v` : 進捗を表示

## Protobuf API ドキュメント生成（protoc-gen-doc）

Docker の `pseudomuto/protoc-gen-doc` を使って `.proto` から Markdown を生成できます。

### 生成（mise）

`mise run proto-doc`

### 生成（dockerを直接実行）

```sh
docker run --rm \
	--user "$(id -u):$(id -g)" \
	-v "$PWD/api:/protos" \
	-v "$PWD/docs/proto:/out" \
	--entrypoint protoc \
	pseudomuto/protoc-gen-doc \
	-I /protos \
	-I /usr/include \
	--doc_out=/out \
	--doc_opt=markdown,api.md \
	v1/hello.proto
```

出力先は `docs/proto/api.md` です。
