# mac-sync

Sync the config files and programs between macs through Github.

I made this for resolving the [keeping the consistent configs issue](https://apple.stackexchange.com/questions/30966/how-can-i-keep-settings-consistent-between-macs) between mac.

## How to set up

1. Create a private repository for `mac-sync` in Github.

2. Add `mac-sync-programs.yaml`, `mac-sync-configs.yaml` to the `main` branch of the repository.

3. Run `mac-sync push` to upload configuration files to the repository. You need to enter some information when you first try it.

4. In another mac, run `mac-sync pull` to download configuration files from the repository.

5. Run `mac-sync sync` to download all programs.

## Configuration files

### `mac-sync-programs.yaml`

Below example will run `homebrew install` and `npm install` command with the specified programs when enter `mac-sync sync-programs`.

Example:

```
homebrew:
  install: brew install {program}
  uninstall: brew uninstall {program}
  programs:
    - deno
    - python3

npm:
  install: npm i -g {program}
  uninstall: npm rm -g {program}
  programs:
    - fast-cli
    - ts-node
    - n
```

### `mac-sync-configs.yaml`

Below example upload specified config files to `mac-sync-configs` folder of the repository.

Example:

```
sync:
  - ~/Library/Preferences/com.apple.dock.plist
```

## Usage

```
NAME:
   mac-sync - Sync specified programs and config files between macs using Git.

USAGE:
   mac-sync [global options] command

COMMANDS:
   push                 Push local config files to remote
   pull                 Pull configs from remote
   sync                 Sync programs with remote
   clear-cache          Clear cache

GLOBAL OPTIONS:
   --help, -h           Show help
```

## Example

- [dotfiles-macos](https://github.com/jopemachine/dotfiles-macos) - my config files
