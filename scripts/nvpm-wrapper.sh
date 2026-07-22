#!/bin/bash
# nvpm ラッパースクリプト
# Neovim起動前に自動的にプラグインを管理します

CONFIG_FILE="${NVPM_CONFIG:-$HOME/.config/nvpm/plugins.json}"
NVPM_BIN="${NVPM_BIN:-nvpm}"

# nvpmがインストールされているか確認
if ! command -v "$NVPM_BIN" &> /dev/null; then
    echo "エラー: nvpmが見つかりません"
    echo "インストール: https://github.com/kouji/nvpm"
    exit 1
fi

# 設定ファイルが存在するか確認
if [ ! -f "$CONFIG_FILE" ]; then
    echo "警告: 設定ファイルが見つかりません: $CONFIG_FILE"
    echo "サンプル設定を作成しています..."
    mkdir -p "$(dirname "$CONFIG_FILE")"
    cat > "$CONFIG_FILE" <<'EOF'
{
  "plugins": [
    "folke/tokyonight.nvim",
    "nvim-telescope/telescope.nvim",
    "neovim/nvim-lspconfig"
  ]
}
EOF
    echo "設定ファイルを作成しました: $CONFIG_FILE"
fi

# プラグインディレクトリが存在しない場合は初回インストール
NVPM_DIR="$HOME/.local/share/nvim/nvpm"
if [ ! -d "$NVPM_DIR" ]; then
    echo "プラグインをインストールしています..."
    "$NVPM_BIN" -config "$CONFIG_FILE" -cmd install
fi

# 環境変数 NVPM_AUTO_UPDATE が設定されている場合は更新をチェック
if [ -n "$NVPM_AUTO_UPDATE" ]; then
    # 最終チェック時刻を記録するファイル
    LAST_CHECK_FILE="$HOME/.cache/nvpm/last_check"
    mkdir -p "$(dirname "$LAST_CHECK_FILE")"

    # 前回のチェックから24時間以上経過している場合のみチェック
    if [ ! -f "$LAST_CHECK_FILE" ] || [ $(find "$LAST_CHECK_FILE" -mtime +1 2>/dev/null | wc -l) -gt 0 ]; then
        echo "プラグインの更新をチェックしています..."
        "$NVPM_BIN" -config "$CONFIG_FILE" -cmd check 2>&1 | grep -q "Updates available" && {
            echo "更新が利用可能です。nvpm -config $CONFIG_FILE -cmd update を実行してください。"
        }
        touch "$LAST_CHECK_FILE"
    fi
fi

# Neovimを起動
exec nvim "$@"
