[![CircleCI](https://circleci.com/gh/itzg/entrypoint-demoter.svg?style=svg)](https://circleci.com/gh/itzg/entrypoint-demoter)
[!![GitHub release](https://img.shields.io/github/release/itzg/entrypoint-demoter.svg)](https://github.com/itzg/entrypoint-demoter/releases/latest)

# Usage

The command line arguments remaining after the options below are used to execute a 
sub-command with the demoted user ID (uid) and group ID (gid).

If executed as a non-root user, then this tool skips demoting entirely and just executes
the sub-command with the current uid and gid.

## Environment variables

- `UID` : if set, demotes the sub-command to run with the given user ID
- `GID` : if set, demotes the sub-command to run with the given group ID

## Command-line

- `--match path` : uses the ownership of the given path to determine a user and group ID
- `--debug` : enables debug logging
- `--version` : show version and exit