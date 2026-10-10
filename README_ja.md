# nvpm

**nvpm** (Neovim Package Manager) は、Goで書かれたNeovim用の高速なプラグインマネージャーです。

## 名前の由来

**nvpm**は以下の略称です：
- **nv** = **N**eo**v**im
- **pm** = **P**ackage **M**anager

## ✨ 特徴

- 📦 シンプルなJSON設定ファイルで全てのプラグインを管理
- 🚀 Go製の高速な並行処理によるプラグイン操作
- 💾 Gitパーシャルクローンによる高速インストール
- 🔌 イベント、コマンド、ファイルタイプ、キーマッピングベースの遅延ロードの枠組み（Neovim連携は未実装）
- 🔒 ロックファイル `nvpm-lock.json` でバージョン管理
- 💪 非同期タスク実行による高速処理
- 🛠️ CLIインターフェースによる簡単な操作
- 📊 統計情報とキャッシュシステム

## ⚡ 必要要件

- Neovim >= **0.8.0**
- Git >= **2.19.0** (パーシャルクローンサポートのため)
- Go >= **1.27** (ビルドする場合)

## 🚀 インストール

### 方法1: ソースからビルド

```bash
# リポジトリをクローン
git clone https://github.com/ue555/nvpm.git
cd nvpm

# ビルド
go build -o bin/nvpm ./cmd/nvpm

# バイナリを適切な場所に配置
sudo cp bin/nvpm /usr/local/bin/
# または
mkdir -p ~/.local/bin
cp bin/nvpm ~/.local/bin/
```

### 方法2: Makefileを使用

```bash
git clone https://github.com/ue555/nvpm.git
cd nvpm
# Goの依存関係を準備（実行ファイルは配置しません）
make install
make build

# バイナリをPATHに追加
sudo cp bin/nvpm /usr/local/bin/
```

`make install` の実体は `go mod download` と `go mod tidy` です。
実行ファイルを配置する処理ではありません。`make build` 後は、リポジトリ内で
`./bin/nvpm` を使えます。任意の場所から `nvpm` を使う場合は、上記のように
バイナリを `PATH` 内へ配置してください。プラグインのインストールは別途
`-cmd install` で行います。

`make` がない環境では、方法1の `go build -o bin/nvpm ./cmd/nvpm` で直接ビルドできます。

### インストール確認

```bash
nvpm --help
```

## 最小構成で動作確認

上記の取得・ビルド後、リポジトリのディレクトリ内で実行します。
例には [TokyoNight](https://github.com/folke/tokyonight.nvim) を使います。
Neovim 0.8.0以上が必要で、プラグイン自体の追加ビルドは不要です。

1. [examples/plugins-minimal.json](examples/plugins-minimal.json) でインストールします。

   ```bash
   ./bin/nvpm -config examples/plugins-minimal.json -cmd install
   ```

   JSONの内容は `{"plugins": ["folke/tokyonight.nvim"]}` です。
   `~/.local/share/nvim/nvpm/tokyonight.nvim` に配置されます。

2. 最小設定でNeovimを起動します。

   ```bash
   nvim -u examples/init-minimal.lua
   ```

   [examples/init-minimal.lua](examples/init-minimal.lua) は、次の順序で設定します。

   ```lua
   local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")
   local plugin_path = nvpm_path .. "/tokyonight.nvim"
   assert(vim.fn.isdirectory(plugin_path) == 1, "Install tokyonight.nvim with nvpm first")
   vim.opt.runtimepath:prepend(plugin_path)

   vim.opt.termguicolors = true
   require("tokyonight").setup({ style = "night" })
   vim.cmd("colorscheme tokyonight-night")
   ```

3. Neovimで `:colorscheme` を実行し、`tokyonight-night` と表示されることを確認します。
   画面を開かずに確認する場合は次を実行します。成功すると `nvpm: OK` と表示されます。

   ```bash
   nvim --headless -u examples/init-minimal.lua -i NONE \
     "+lua assert(vim.g.colors_name == 'tokyonight-night'); print('nvpm: OK')" +qa
   ```

4. 通常起動でも使うには、既存の設定を残しながら上記のLuaを自分の `init.lua` に組み込みます。
   配置先は `:echo stdpath('config')` で確認できます（通常は `~/.config/nvim`）。
   `runtimepath` への追加後に各プラグインの `setup()` を書き、Neovimを再起動してください。

## 📁 ディレクトリ構成

nvpmは以下のディレクトリを使用します：

```
~/.local/share/nvim/
├── nvpm/                    # プラグインインストールディレクトリ
│   ├── <plugin-name>/       # 各プラグインディレクトリ
│   └── cache/               # キャッシュディレクトリ
└── nvpm-lock.json           # ロックファイル
```

### macOS・カスタムパスでの注意点

現在のCLIは、Linux・macOSともに上記のホームディレクトリ配下へ配置します。
Neovimの `stdpath("data")`、`XDG_DATA_HOME`、`NVIM_APPNAME` には追従しません。
そのため環境によっては `vim.fn.stdpath("data") .. "/nvpm"` と実際の配置先が一致しません。
設定例では `vim.fn.expand("~/.local/share/nvim/nvpm")` を使って実装に合わせています。
Neovimのデータディレクトリだけを変更しても、nvpmの配置先は変わりません。
また、現在のCLIはJSONの `root` 指定を読み取りません。

## ⚙️ 設定

### 1. 設定ファイルの作成

プロジェクトディレクトリまたはホームディレクトリに設定ファイルを作成します：

```bash
# プロジェクトディレクトリに作成
mkdir -p ~/.config/nvpm
vim ~/.config/nvpm/plugins.json
```

### 2. プラグインの設定

`plugins.json` に使用したいプラグインを記述します：

```json
{
  "plugins": [
    "folke/tokyonight.nvim",
    "nvim-telescope/telescope.nvim",
    "neovim/nvim-lspconfig",
    {
      "url": "hrsh7th/nvim-cmp",
      "event": ["InsertEnter"],
      "dependencies": ["L3MON4D3/LuaSnip", "saadparwaiz1/cmp_luasnip"]
    },
    {
      "url": "nvim-treesitter/nvim-treesitter",
      "build": ":TSUpdate",
      "event": ["BufReadPost", "BufNewFile"]
    },
    {
      "url": "lewis6991/gitsigns.nvim",
      "event": ["BufReadPre"],
      "config": "require('gitsigns').setup()"
    }
  ]
}
```

### プラグイン設定オプション

#### シンプルな書き方

```json
"folke/tokyonight.nvim"
```

#### 詳細な設定

以下は説明用のコメント付き例です。実際のJSONファイルではコメントを除去してください。
遅延ロードや `config` の実行状況は下記「現在の実装状況」を参照してください。

```jsonc
{
  "url": "hrsh7th/nvim-cmp", // GitHubリポジトリ (必須)
  "name": "nvim-cmp", // プラグイン名 (オプション)
  "lazy": true, // 遅延ロードのメタデータ (Neovim連携は未実装)
  "event": ["InsertEnter"], // イベントトリガー
  "cmd": ["CmpStatus"], // コマンドトリガー
  "ft": ["lua", "vim"], // ファイルタイプトリガー
  "keys": ["<leader>c"], // キーマッピングトリガー
  "dependencies": ["L3MON4D3/LuaSnip"], // 依存プラグイン
  "branch": "main", // Gitブランチ
  "tag": "v1.0.0", // Gitタグ
  "commit": "abc123", // Git コミットハッシュ
  "build": "make install", // ビルドコマンド (詳細は下記「ビルドコマンドの実行」参照)
  "config": "require('plugin').setup()", // Lua設定文字列 (通常のロード処理では実行されません)
  "cond": true // 有効条件
}
```

### 現在の実装状況

| 項目 | 動作 |
| --- | --- |
| Gitによる取得・更新、branch/tag/commit固定、ロックファイル | CLIで管理します。Neovim側での読み込み設定は別途必要です。 |
| `build` | install/update時に実行します。`:` コマンドは別プロセスのヘッドレスNeovimで実行します。 |
| `lazy`・`event`・`cmd`・`ft`・`keys` | 枠組みのみです。Neovimに遅延ロード用のイベント・コマンド・マッピングは登録されません。 |
| `config`・`init` | 通常のロード処理ではLua文字列を実行しません。自分の `init.lua` に設定を書いてください。 |
| `dependencies` | 依存先を自動インストールしません。必要なリポジトリをトップレベルの `plugins` にも列挙してください。 |

現在のJSONパーサーは、トリガー項目と `dependencies` の配列の変換にも未対応です。
配列を記述した設定例だけでは、遅延ロードや依存解決は完成しません。
読み込み順序、キーマッピング、必要な遅延ロードはNeovim側で設定してください。
CLIの `loaded` 表示は内部状態であり、起動中のNeovimでの読み込みを表すものではありません。

### ビルドコマンドの実行

`build`フィールドは書き方によって2通りに解釈されます。

- **シェルコマンド**（`:`で始まらない文字列）: プラグインのディレクトリ内で`sh -c`実行されます。例: `"make install_jsregexp"`
- **Neovim Exコマンド**（`:`で始まる文字列）: ヘッドレスかつ独立したNeovimインスタンス（`-u NONE -i NONE`）を起動し、対象プラグイン本体と`dependencies`に指定した依存プラグインだけを`runtimepath`に追加して実行されます。例: `":TSUpdate"`、`":MasonUpdate"`

`-u NONE`は`plugin/`配下スクリプトの自動読み込みも無効化するため、nvpmは`:runtime! plugin/**/*.vim plugin/**/*.lua`を明示的に実行してから対象コマンドを実行します。また、`mason.nvim`の`:MasonUpdate`のように`plugin/`スクリプトではなく`setup()`内でしかコマンドが登録されないプラグインについては、プラグインの`lua/`ディレクトリ構成からモジュール名を推測し、`require(<module>).setup({})`をベストエフォートで先に実行します。

なお、`mason-lspconfig.nvim`の`ensure_installed`のように、プラグイン自身がヘッドレス実行時に自動インストールを意図的にスキップする仕様を持つ場合があります。これはnvpm側で制御できない、プラグイン側の挙動です。

## 🎯 使い方

### 基本コマンド

```bash
# プラグインをインストール
nvpm -config ~/.config/nvpm/plugins.json -cmd install

# プラグインを更新
nvpm -config ~/.config/nvpm/plugins.json -cmd update

# プラグインを同期（クリーン + インストール + 更新）
nvpm -config ~/.config/nvpm/plugins.json -cmd sync

# 更新をチェック
nvpm -config ~/.config/nvpm/plugins.json -cmd check

# インストール済みプラグイン一覧
nvpm -config ~/.config/nvpm/plugins.json -cmd list

# 統計情報を表示
nvpm -config ~/.config/nvpm/plugins.json -cmd stats

# ロックファイルから復元
nvpm -config ~/.config/nvpm/plugins.json -cmd restore

# 未使用プラグインを削除
nvpm -config ~/.config/nvpm/plugins.json -cmd clean
```

`clean`は「インストール先ディレクトリ（`~/.local/share/nvim/nvpm/`）には存在するが、現在の`plugins.json`にはもう記載がないプラグイン」を自動検出し、そのディレクトリを削除します（`cache`ディレクトリは対象外）。プラグインを設定ファイルから削除した後にこのコマンドを実行すると、使われなくなったディレクトリを掃除できます。

### エイリアスの設定（推奨）

毎回長いコマンドを打つのは面倒なので、シェルにエイリアスを設定することをおすすめします：

```bash
# ~/.bashrc または ~/.zshrc に追加
alias nvpm-install="nvpm -config ~/.config/nvpm/plugins.json -cmd install"
alias nvpm-update="nvpm -config ~/.config/nvpm/plugins.json -cmd update"
alias nvpm-sync="nvpm -config ~/.config/nvpm/plugins.json -cmd sync"
alias nvpm-list="nvpm -config ~/.config/nvpm/plugins.json -cmd list"
alias nvpm-stats="nvpm -config ~/.config/nvpm/plugins.json -cmd stats"
```

設定後は以下のように簡単に使えます：

```bash
nvpm-install  # プラグインをインストール
nvpm-update   # プラグインを更新
nvpm-sync     # 同期
```

## 🔧 Neovimとの連携

### 方法1: 手動でruntimepathに追加

`~/.config/nvim/init.lua` に以下を追加：

```lua
-- nvpmでインストールしたプラグインをruntimepathに追加
local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")

-- nvpmディレクトリ内の全プラグインを検索
local plugins = vim.fn.glob(nvpm_path .. "/*", false, true)

for _, plugin in ipairs(plugins) do
  if vim.fn.isdirectory(plugin) == 1 and vim.fn.fnamemodify(plugin, ":t") ~= "cache" then
    vim.opt.rtp:prepend(plugin)
    if vim.fn.isdirectory(plugin .. "/after") == 1 then
      vim.opt.rtp:append(plugin .. "/after")
    end
  end
end

-- プラグインの設定をここに記述
-- 例：
-- require('tokyonight').setup()
-- vim.cmd[[colorscheme tokyonight]]
```

この例は起動時に全プラグインを読み込む構成です。遅延ロードは行いません。
設定変更後はNeovimを再起動し、通常の起動処理で `plugin/` スクリプトを読み込ませます。
[examples/init.lua](examples/init.lua) にも複数プラグイン用の例があります。

### 方法2: 自動起動スクリプト

Neovim起動前にnvpmを実行するスクリプトも使えます。これは取得処理の補助であり、方法1の `init.lua` 設定も必要です：

```bash
#!/bin/bash
# ~/.local/bin/vim-nvpm

# 設定ファイルのパス
CONFIG="$HOME/.config/nvpm/plugins.json"

# プラグインをインストール（初回のみ）
if [ ! -d "$HOME/.local/share/nvim/nvpm" ]; then
  echo "プラグインをインストールしています..."
  nvpm -config "$CONFIG" -cmd install
fi

# Neovimを起動
nvim "$@"
```

実行権限を付与：

```bash
chmod +x ~/.local/bin/vim-nvpm
```

エイリアスを設定：

```bash
# ~/.bashrc または ~/.zshrc
alias vim='vim-nvpm'
alias nvim='vim-nvpm'
```

## 📚 使用例

### 例1: 最小構成

```json
{
  "plugins": [
    "folke/tokyonight.nvim",
    "nvim-telescope/telescope.nvim",
    "neovim/nvim-lspconfig"
  ]
}
```

```bash
# インストール
nvpm -config plugins.json -cmd install

# Neovimで使用
nvim
```

```lua
-- ~/.config/nvim/init.lua
local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")
local plugins = vim.fn.glob(nvpm_path .. "/*", false, true)
for _, plugin in ipairs(plugins) do
  if vim.fn.isdirectory(plugin) == 1 and vim.fn.fnamemodify(plugin, ":t") ~= "cache" then
    vim.opt.rtp:prepend(plugin)
    if vim.fn.isdirectory(plugin .. "/after") == 1 then
      vim.opt.rtp:append(plugin .. "/after")
    end
  end
end

-- カラースキームを適用
vim.cmd[[colorscheme tokyonight]]
```

### 例2: LSP環境の構築

```json
{
  "plugins": [
    "neovim/nvim-lspconfig",
    {
      "url": "hrsh7th/nvim-cmp",
      "dependencies": [
        "hrsh7th/cmp-nvim-lsp",
        "hrsh7th/cmp-buffer",
        "hrsh7th/cmp-path",
        "L3MON4D3/LuaSnip",
        "saadparwaiz1/cmp_luasnip"
      ]
    },
    "williamboman/mason.nvim",
    "williamboman/mason-lspconfig.nvim"
  ]
}
```

```bash
nvpm -config lsp-plugins.json -cmd install
```

```lua
-- ~/.config/nvim/init.lua
local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")
local plugins = vim.fn.glob(nvpm_path .. "/*", false, true)
for _, plugin in ipairs(plugins) do
  if vim.fn.isdirectory(plugin) == 1 and vim.fn.fnamemodify(plugin, ":t") ~= "cache" then
    vim.opt.rtp:prepend(plugin)
    if vim.fn.isdirectory(plugin .. "/after") == 1 then
      vim.opt.rtp:append(plugin .. "/after")
    end
  end
end

-- LSP設定
require("mason").setup()
require("mason-lspconfig").setup()

local lspconfig = require('lspconfig')
lspconfig.lua_ls.setup{}
lspconfig.gopls.setup{}

-- 補完設定
local cmp = require('cmp')
cmp.setup({
  snippet = {
    expand = function(args)
      require('luasnip').lsp_expand(args.body)
    end,
  },
  mapping = cmp.mapping.preset.insert({
    ['<C-Space>'] = cmp.mapping.complete(),
    ['<CR>'] = cmp.mapping.confirm({ select = true }),
  }),
  sources = cmp.config.sources({
    { name = 'nvim_lsp' },
    { name = 'luasnip' },
  }, {
    { name = 'buffer' },
  })
})
```

### 例3: 定期的な更新

cronで定期的にプラグインを更新：

```bash
# crontabを編集
crontab -e

# 毎日午前3時にプラグインを更新
0 3 * * * /usr/local/bin/nvpm -config ~/.config/nvpm/plugins.json -cmd update
```

## 🔒 ロックファイル

nvpmは `nvpm-lock.json` を使用してプラグインのバージョンを管理します。

### ロックファイルの場所

```
~/.local/share/nvim/nvpm-lock.json
```

### ロックファイルの形式

```json
{
  "plugins": {
    "tokyonight.nvim": {
      "commit": "1a1c7942ee0e1a1c7942ee0e1a1c7942ee0e1a1c",
      "branch": "main"
    },
    "telescope.nvim": {
      "commit": "2b2b2c2942ee0e1a1c7942ee0e1a1c7942ee0e2b",
      "branch": "master"
    }
  }
}
```

### ロックファイルの使用

```bash
# 現在のバージョンをロックファイルに記録
nvpm -config plugins.json -cmd update

# ロックファイルからバージョンを復元
nvpm -config plugins.json -cmd restore
```

## 🚀 ヘルパースクリプト

便利に使えるヘルパースクリプトを提供しています。

### インストールスクリプト

自動でnvpmをインストールします：

```bash
curl -fsSL https://raw.githubusercontent.com/ue555/nvpm/main/scripts/install.sh | bash
```

または手動で：

```bash
git clone https://github.com/ue555/nvpm.git
cd nvpm
./scripts/install.sh
```

### CLIヘルパー (nvpm-cli.sh)

コマンドを簡潔に実行できます：

```bash
# インストール
cp scripts/nvpm-cli.sh ~/.local/bin/nvpm-cli
chmod +x ~/.local/bin/nvpm-cli

# 使い方
nvpm-cli install    # プラグインをインストール
nvpm-cli update     # プラグインを更新
nvpm-cli list       # プラグイン一覧
nvpm-cli stats      # 統計情報
nvpm-cli config     # 設定ファイルを編集
```

### Neovimラッパー (nvpm-wrapper.sh)

Neovim起動時に自動的にプラグインを管理します：

```bash
# インストール
cp scripts/nvpm-wrapper.sh ~/.local/bin/vim-nvpm
chmod +x ~/.local/bin/vim-nvpm

# エイリアスを設定
echo 'alias vim="vim-nvpm"' >> ~/.bashrc
echo 'alias nvim="vim-nvpm"' >> ~/.bashrc

# 自動更新チェックを有効化（オプション）
echo 'export NVPM_AUTO_UPDATE=1' >> ~/.bashrc
```

## 💡 ヒントとコツ

### 1. 設定ファイルを複数使い分ける

```bash
# 開発環境用
nvpm -config ~/.config/nvpm/dev.json -cmd install

# 最小構成用
nvpm -config ~/.config/nvpm/minimal.json -cmd install
```

### 2. 特定のプラグインのみ更新

```bash
nvpm -config plugins.json -cmd update -plugin tokyonight.nvim
```

### 3. 統計情報で確認

```bash
nvpm -config plugins.json -cmd stats
```

出力例：

```
NVPM Statistics:
  Total plugins:     10
  Installed plugins: 10
  Loaded plugins:    0
  Cache entries:     5
  Cache size:        1024 bytes
```

### 4. デバッグ

詳細なログを確認したい場合は、標準エラー出力も表示：

```bash
nvpm -config plugins.json -cmd install 2>&1 | tee nvpm.log
```

## 🐛 トラブルシューティング

### プラグインがインストールできない

```bash
# Gitの設定を確認
git --version  # Git 2.19.0以上が必要

# ネットワーク接続を確認
ping github.com

# ディレクトリの権限を確認
ls -la ~/.local/share/nvim/
```

### プラグインがNeovimで認識されない

```lua
-- init.luaで正しくruntimepathに追加されているか確認
print(vim.inspect(vim.opt.rtp:get()))
```

### ロックファイルが壊れた

```bash
# ロックファイルを削除して再生成
rm ~/.local/share/nvim/nvpm-lock.json
nvpm -config plugins.json -cmd update
```

## 📖 よくある質問

### Q: lazy.nvimとの違いは？

A: nvpmはGoで書かれた独立したCLIツールです。lazy.nvimはNeovimプラグインとして動作し、UI機能も持っていますが、nvpmはコマンドラインから操作します。

### Q: 既存のlazy.nvimから移行できますか？

A: プラグインリストをJSON形式に変換する必要がありますが、基本的な設定構造は似ているため、比較的簡単に移行できます。

### Q: Windowsで使えますか？

A: Goがクロスプラットフォーム対応なので、Windows用にビルドすることで使用可能です。

### Q: プラグインの遅延ロードは自動的に機能しますか？

A: 自動では機能しません。ハンドラーとNeovimの連携は未実装で、`config`・`init` のLua実行も未対応です。読み込みと設定は `init.lua` に記述してください。上記「現在の実装状況」を参照してください。

## 🤝 貢献

バグ報告や機能要望は、GitHubのIssueでお願いします。

## 📄 ライセンス

MIT License

## 🙏 謝辞

このプロジェクトは [lazy.nvim](https://github.com/folke/lazy.nvim) by folke にインスパイアされて作成されました。
