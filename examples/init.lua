-- nvpm で管理されたプラグインを読み込むための設定例
-- このファイルを ~/.config/nvim/init.lua にコピーまたは参照してください

-- nvpmでインストールしたプラグインのパス
local nvpm_path = vim.fn.stdpath("data") .. "/nvpm"

-- nvpmディレクトリが存在するか確認
if vim.fn.isdirectory(nvpm_path) == 1 then
  -- nvpmディレクトリ内の全プラグインを検索
  local plugins = vim.fn.glob(nvpm_path .. "/*", false, true)

  for _, plugin in ipairs(plugins) do
    -- runtimepathに追加
    vim.opt.rtp:prepend(plugin)

    -- after/ディレクトリも追加（プラグインによっては必要）
    local after_dir = plugin .. "/after"
    if vim.fn.isdirectory(after_dir) == 1 then
      vim.opt.rtp:append(after_dir)
    end
  end

  -- プラグインをロード
  vim.cmd("packloadall")
else
  print("nvpm: プラグインディレクトリが見つかりません: " .. nvpm_path)
  print("nvpm -config <設定ファイル> -cmd install を実行してください")
end

-- 基本設定
vim.opt.number = true          -- 行番号表示
vim.opt.relativenumber = true  -- 相対行番号
vim.opt.expandtab = true       -- タブをスペースに展開
vim.opt.tabstop = 2           -- タブ幅
vim.opt.shiftwidth = 2        -- インデント幅
vim.opt.smartindent = true    -- スマートインデント

-- カラースキーム設定（tokyonight.nvimがインストールされている場合）
pcall(function()
  vim.cmd[[colorscheme tokyonight]]
end)

-- プラグイン設定例
-- Telescope（telescope.nvimがインストールされている場合）
pcall(function()
  local telescope = require('telescope')
  telescope.setup{
    defaults = {
      mappings = {
        i = {
          ["<C-j>"] = "move_selection_next",
          ["<C-k>"] = "move_selection_previous",
        },
      },
    },
  }

  -- キーマッピング
  vim.keymap.set('n', '<leader>ff', '<cmd>Telescope find_files<cr>')
  vim.keymap.set('n', '<leader>fg', '<cmd>Telescope live_grep<cr>')
  vim.keymap.set('n', '<leader>fb', '<cmd>Telescope buffers<cr>')
end)

-- Treesitter（nvim-treesitterがインストールされている場合）
pcall(function()
  require('nvim-treesitter.configs').setup {
    highlight = {
      enable = true,
    },
    indent = {
      enable = true,
    },
  }
end)

-- LSP設定（nvim-lspconfigがインストールされている場合）
pcall(function()
  local lspconfig = require('lspconfig')

  -- キーマッピング
  local on_attach = function(client, bufnr)
    local opts = { buffer = bufnr, noremap = true, silent = true }
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, opts)
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, opts)
    vim.keymap.set('n', 'gi', vim.lsp.buf.implementation, opts)
    vim.keymap.set('n', '<C-k>', vim.lsp.buf.signature_help, opts)
    vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, opts)
    vim.keymap.set('n', '<leader>ca', vim.lsp.buf.code_action, opts)
    vim.keymap.set('n', 'gr', vim.lsp.buf.references, opts)
  end

  -- 各種LSPサーバーの設定
  lspconfig.lua_ls.setup{ on_attach = on_attach }
  lspconfig.gopls.setup{ on_attach = on_attach }
  lspconfig.rust_analyzer.setup{ on_attach = on_attach }
end)

-- nvim-cmp（補完設定）
pcall(function()
  local cmp = require('cmp')
  local luasnip = require('luasnip')

  cmp.setup({
    snippet = {
      expand = function(args)
        luasnip.lsp_expand(args.body)
      end,
    },
    mapping = cmp.mapping.preset.insert({
      ['<C-b>'] = cmp.mapping.scroll_docs(-4),
      ['<C-f>'] = cmp.mapping.scroll_docs(4),
      ['<C-Space>'] = cmp.mapping.complete(),
      ['<C-e>'] = cmp.mapping.abort(),
      ['<CR>'] = cmp.mapping.confirm({ select = true }),
      ['<Tab>'] = cmp.mapping(function(fallback)
        if cmp.visible() then
          cmp.select_next_item()
        elseif luasnip.expand_or_jumpable() then
          luasnip.expand_or_jump()
        else
          fallback()
        end
      end, { 'i', 's' }),
    }),
    sources = cmp.config.sources({
      { name = 'nvim_lsp' },
      { name = 'luasnip' },
    }, {
      { name = 'buffer' },
      { name = 'path' },
    })
  })
end)

-- gitsigns（Gitサインカラム表示）
pcall(function()
  require('gitsigns').setup {
    signs = {
      add          = { text = '+' },
      change       = { text = '~' },
      delete       = { text = '_' },
      topdelete    = { text = '‾' },
      changedelete = { text = '~' },
    },
  }
end)

-- which-key（キーバインドヘルプ）
pcall(function()
  require('which-key').setup {}
end)

print("nvpm: 設定ファイルの読み込みが完了しました")
