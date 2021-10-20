# Scripthub
A CLI tool to manage all your scripts scattered around on your computer.

## Usage
- Prefix all commands with `scripthub` (Recommended: Set up `sch` as alias for `scripthub`)

### Commands
- `list, ls` - List all available scripts
- `edit, e` - Edit the given script
	- `{name of script to edit}`
- `add, a` - Add a script to the library
	- `--name, -n` - Name of script to add
	- `--executable, -x` - Relative path to executable script
	- `--editable, -e` - Relative path to editable script (Optional. Defaults to `--executable`)
- `run, r` - Run the given script
	- `{name of script to edit}`
- `path, p` - Get paths for given script
	- `--specifier, -s` - `x` || `executable` for only executable path, `e` || `editable` for only editable path
	- `{name of script to get paths for}`
- `setup` - Set up scripthub
- `help, h` - Show help overview

#### Global options
- `--help, -h` - Show help for command

### Configuration
- Change editor to open files in when running `edit`.
	- Add this line to your shell config file: `export EDITOR="{YOUR FAVORITE EDITOR}"`
