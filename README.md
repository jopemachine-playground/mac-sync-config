# mac-sync-config

Sync your config files between macs through Github.

I made this for resolving the [keeping the consistent configs issue](https://apple.stackexchange.com/questions/30966/how-can-i-keep-settings-consistent-between-macs) between mac.

## How to set up

1. Create a private repository for `mac-sync-config` in Github.

2. Add `mac-sync-configs.yaml` to the `main` branch of the repository.

3. Run `mac-sync-config push` to upload the configuration files to the repository. You need to enter Github access token when you first try it.

4. In another mac, run `mac-sync-config pull` to download configuration files from the repository.

## Configuration

### `mac-sync-configs.yaml`

Below example upload the specified config files to `mac-sync-configs` folder of the repository.

You can also specify directory path to the below paths.

Example:

```
sync:
  - ~/Library/Preferences/com.apple.dock.plist
```

## Usage

```
NAME:
   mac-sync-config - Sync the config files between macs through Github

USAGE:
   mac-sync-config command [command options] [arguments...]

COMMANDS:
   push                 Push local config files to remote
   pull                 Pull configs from remote
   list, ls             Show the configuration files list
   clear-cache          Clear cache
```

## Example

- [dotfiles-macos](https://github.com/jopemachine/dotfiles-macos) - my config files
