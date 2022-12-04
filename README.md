# mac-sync-config

<p>
	<img src="https://img.shields.io/github/license/jopemachine/mac-sync-config.svg" alt="License">
	<img src="https://goreportcard.com/badge/github.com/jopemachine/mac-sync-config" alt="goreportcard">
</p>

Sync your config files between macs through your Github repository.

I made this for resolving the [keeping the consistent configs issue](https://apple.stackexchange.com/questions/30966/how-can-i-keep-settings-consistent-between-macs) between mac.

⚠️ Note that this project is still Experimental, could include unexpected bug. You can issue the bug with its stacktrace using `debug` flag.

## Why?

- No need to write any shell scripts to sync config files.

- Easy to manage lots of config files using saving those to your Github repository.

- Easy to check diffs between remote configs and local configs.

- Easy to Edit your config files in your local, Github directly.

## Installation

```
$ brew tap jopemachine/mac-sync-config
$ brew install mac-sync-config
```

## How to set up

1. Create a repository for `mac-sync-config` through your Github account.

2. Add `mac-sync-configs.yaml` to the `main` branch of the repository.

3. Run `mac-sync-config push` to upload the configuration files to the repository. Note that you need to enter Github access token when you first try it.

4. Run `mac-sync-config pull` to download configuration files from the repository.

## Configuration

### `mac-sync-configs.yaml`

Below example upload the specified config files to `mac-sync-configs` folder of the repository.

You can also specify directory path to the below paths.

Example:

```yaml
sync:
  # You can use '~'.
  - ~/.tmux.conf
  # You can specify directory.
  - ~/Library/Application Support/PopClip/
```

## Multiple profiles

You can use multiple profiles for your own purpose.

For example, it could be useful when you want to save multiple config files to each profile.

You can use profile name to upload from or to the profile folder.

```
$ mac-sync-config push [profile name]
$ mac-sync-config pull [profile name]
```

## CLI Usage

```
NAME:
   mac-sync-config - Sync the config files between macs through Github

USAGE:
   mac-sync-config command [command options] [arguments...]

COMMANDS:
   push                     Push the local config files to the remote repository
   pull                     Pull the config files from the remote repository
   list, ls                 Show the configuration files list
   switch-profile, profile  Change default profile. This could be useful when you need to the configuration set
   delete-keychain          Delete keychain configurations

GLOBAL OPTIONS:
   --debug        use panic instead of log.Fatal to show stacktrace (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## `mac-sync-configs` Example

- [my-mac-sync-configs](https://github.com/jopemachine/my-mac-sync-configs) - my `mac-sync-configs` config files.
