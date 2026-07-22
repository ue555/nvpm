#!/bin/bash
# nvpm インストールスクリプト

set -e

NVPM_VERSION="latest"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.config/nvpm"
DATA_DIR="$HOME/.local/share/nvim"

echo "======================================"
echo "   nvpm インストールスクリプト"
echo "======================================"
echo ""

# ディレクトリ作成
echo "📁 ディレクトリを作成しています..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR/nvpm"

# Goがインストールされているか確認
if ! command -v go &> /dev/null; then
    echo "❌ エラー: Goがインストールされていません"
    echo "   https://golang.org/dl/ からGoをインストールしてください"
    exit 1
fi

# Gitがインストールされているか確認
if ! command -v git &> /dev/null; then
    echo "❌ エラー: Gitがインストールされていません"
    exit 1
fi

# リポジトリをクローン
echo "📥 nvpmをダウンロードしています..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

git clone https://github.com/kouji/nvpm.git
cd nvpm

# ビルド
echo "🔨 nvpmをビルドしています..."
go build -o nvpm ./cmd/nvpm

# インストール
echo "📦 nvpmをインストールしています..."
cp nvpm "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/nvpm"

# サンプル設定をコピー
if [ ! -f "$CONFIG_DIR/plugins.json" ]; then
    echo "📄 サンプル設定ファイルをコピーしています..."
    cp examples/config.json "$CONFIG_DIR/plugins.json"
fi

# クリーンアップ
cd ~
rm -rf "$TEMP_DIR"

echo ""
echo "✅ インストールが完了しました！"
echo ""
echo "📍 インストール先: $INSTALL_DIR/nvpm"
echo "📍 設定ファイル: $CONFIG_DIR/plugins.json"
echo ""

# PATHに追加されているか確認
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  警告: $INSTALL_DIR がPATHに含まれていません"
    echo ""
    echo "以下のコマンドを実行してPATHに追加してください："
    echo ""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    echo ""
    echo "または zsh を使用している場合："
    echo ""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
    echo "  source ~/.zshrc"
    echo ""
fi

echo "次のステップ："
echo "1. プラグインをインストール:"
echo "   nvpm -config $CONFIG_DIR/plugins.json -cmd install"
echo ""
echo "2. Neovimの設定ファイルを編集:"
echo "   vim ~/.config/nvim/init.lua"
echo ""
echo "詳細は README_ja.md を参照してください。"
