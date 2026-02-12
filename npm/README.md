# misaki-banner

8×8ドットの[美咲フォント](https://littlelimit.net/misaki.htm)を使って、日本語テキストをターミナル上にバナー表示するCLIツールです。

## インストール

### Go

```bash
go install github.com/qraqras/misaki-banner/cmd/misaki-banner@latest
```

### npm

```bash
# グローバルインストール
npm install -g misaki-banner

# または npx で一時実行
npx misaki-banner "こんにちは"
```

### バイナリダウンロード

[GitHub Releases](https://github.com/qraqras/misaki-banner/releases) から OS 別のバイナリをダウンロードできます。

## 使い方

```bash
misaki-banner [オプション] <テキスト>
```

### オプション

| フラグ | 説明 | デフォルト |
|---|---|---|
| `-font` | フォント名: `misaki_gothic`, `misaki_gothic_2nd`, `misaki_mincho` | `misaki_gothic_2nd` |
| `-shadow` | 影スタイル: `outline` (罫線), `solid` (シェーディング) | なし |
| `-color` | テキスト色: プリセット名, hex (`#RRGGBB`/`RRGGBB`), RGB (`r,g,b`) | なし |
| `-gradient` | グラデーション効果を有効化 | `false` |

### カラープリセット

| 名前 | 色 |
|---|---|
| `c` | シアン |
| `m` | マゼンタ |
| `y` | イエロー |

### 例

```bash
# 基本
misaki-banner "こんにちは"

# フォント指定
misaki-banner -font misaki_mincho "こんにちは"

# 影付き
misaki-banner -shadow solid "こんにちは"

# カラー + グラデーション + 影
misaki-banner -color c -gradient -shadow solid "こんにちは"

# hex カラー
misaki-banner -color "#ff4444" -gradient "みさきバナー"

# 複数行 (\n で改行)
misaki-banner "Hello\nWorld"
```

## 開発

### ビルド

```bash
go build -o misaki-banner ./cmd/misaki-banner
```

### テスト

```bash
go test ./...
```

### リリース

このプロジェクトは [GoReleaser](https://goreleaser.com/) を使用して自動リリースします。

1. バージョンタグを作成:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. GitHub Actions が自動的に:
   - Linux/macOS/Windows 向けにビルド
   - GitHub Releases にバイナリを公開
   - checksums と changelog を生成

ローカルでテストビルドする場合:
```bash
goreleaser build --snapshot --clean
# → dist/ ディレクトリに成果物が生成されます
```
