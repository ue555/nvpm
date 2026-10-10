-- Match nvpm's install directory on Linux and macOS.
local nvpm_path = vim.fn.expand("~/.local/share/nvim/nvpm")
local plugin_path = nvpm_path .. "/tokyonight.nvim"
assert(vim.fn.isdirectory(plugin_path) == 1, "Install tokyonight.nvim with nvpm first")
vim.opt.runtimepath:prepend(plugin_path)

-- Put plugin setup() calls after adding their directories to runtimepath.
vim.opt.termguicolors = true
require("tokyonight").setup({ style = "night" })
vim.cmd("colorscheme tokyonight-night")
