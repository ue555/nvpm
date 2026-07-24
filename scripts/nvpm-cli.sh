#!/bin/bash
# nvpm CLIヘルパースクリプト
# コマンドを簡潔に実行するためのラッパー

CONFIG_FILE="${NVPM_CONFIG:-$HOME/.config/nvpm/plugins.json}"
NVPM_BIN="${NVPM_BIN:-nvpm}"

# 色の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ヘルプメッセージ
show_help() {
    echo "nvpm - Neovim Package Manager CLI"
    echo ""
    echo "使い方: $(basename "$0") <コマンド> [オプション]"
    echo ""
    echo "コマンド:"
    echo "  install, i      - プラグインをインストール"
    echo "  update, u       - プラグインを更新"
    echo "  sync, s         - 同期（クリーン + インストール + 更新）"
    echo "  clean, c        - 未使用プラグインを削除"
    echo "  list, l         - プラグイン一覧を表示"
    echo "  stats, st       - 統計情報を表示"
    echo "  check, ch       - 更新を確認"
    echo "  restore, r      - ロックファイルから復元"
    echo "  config, cf      - 設定ファイルを編集"
    echo "  help, h         - このヘルプを表示"
    echo ""
    echo "環境変数:"
    echo "  NVPM_CONFIG     - 設定ファイルのパス (デフォルト: ~/.config/nvpm/plugins.json)"
    echo "  NVPM_BIN        - nvpmバイナリのパス (デフォルト: nvpm)"
    echo ""
    echo "例:"
    echo "  $(basename "$0") install          # プラグインをインストール"
    echo "  $(basename "$0") update           # プラグインを更新"
    echo "  $(basename "$0") list             # プラグイン一覧"
    echo ""
}

# エラーメッセージ
error() {
    echo -e "${RED}エラー:${NC} $1" >&2
    exit 1
}

# 成功メッセージ
success() {
    echo -e "${GREEN}✓${NC} $1"
}

# 情報メッセージ
info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# 警告メッセージ
warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# nvpmの存在確認
check_nvpm() {
    if ! command -v "$NVPM_BIN" &> /dev/null; then
        error "nvpmが見つかりません。インストールしてください: https://github.com/ue555/nvpm"
    fi
}

# 設定ファイルの存在確認
check_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        warning "設定ファイルが見つかりません: $CONFIG_FILE"
        return 1
    fi
    return 0
}

# nvpmコマンドを実行
run_nvpm() {
    local cmd="$1"
    info "実行中: nvpm -config $CONFIG_FILE -cmd $cmd"
    "$NVPM_BIN" -config "$CONFIG_FILE" -cmd "$cmd"
}

# メインロジック
main() {
    check_nvpm

    local command="${1:-help}"

    case "$command" in
        install|i)
            check_config || exit 1
            run_nvpm "install"
            success "インストールが完了しました"
            ;;
        update|u)
            check_config || exit 1
            run_nvpm "update"
            success "更新が完了しました"
            ;;
        sync|s)
            check_config || exit 1
            run_nvpm "sync"
            success "同期が完了しました"
            ;;
        clean|c)
            check_config || exit 1
            run_nvpm "clean"
            success "クリーンアップが完了しました"
            ;;
        list|l)
            check_config || exit 1
            run_nvpm "list"
            ;;
        stats|st)
            check_config || exit 1
            run_nvpm "stats"
            ;;
        check|ch)
            check_config || exit 1
            run_nvpm "check"
            ;;
        restore|r)
            check_config || exit 1
            run_nvpm "restore"
            success "復元が完了しました"
            ;;
        config|cf)
            if [ ! -f "$CONFIG_FILE" ]; then
                warning "設定ファイルが存在しません。新規作成します。"
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
                success "設定ファイルを作成しました: $CONFIG_FILE"
            fi
            ${EDITOR:-vim} "$CONFIG_FILE"
            ;;
        help|h|--help|-h)
            show_help
            ;;
        *)
            error "不明なコマンド: $command\nヘルプを表示するには '$(basename "$0") help' を実行してください"
            ;;
    esac
}

main "$@"
