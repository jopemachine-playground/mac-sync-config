# mac-sync-config

<img src="https://img.shields.io/github/license/jopemachine/mac-sync-config.svg" alt="License">

Sync your config files between macs through your Github repository.

I made this for resolving the [keeping the consistent configs issue](https://apple.stackexchange.com/questions/30966/how-can-i-keep-settings-consistent-between-macs) between mac.

## Why?

- No need to write any shell scripts to sync config files.

- Easy to manage lots of config files using saving those to your Github repository.

- Easy to check diffs between remote configs and local configs.

- Easy to Edit your config files in your local, Github directly.

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
   push                     Push the local config files to the remote repository
   pull                     Pull the config files from the remote repository
   list, ls                 Show the configuration files list
   switch-profile, profile  Switch the profile. This could be useful when you need to the configuration set
```

## Example

- [mac-sync-configs](https://github.com/jopemachine/mac-sync-configs) - my config files
