[![CircleCI](https://circleci.com/gh/itzg/entrypoint-demoter.svg?style=svg)](https://circleci.com/gh/itzg/entrypoint-demoter)
[![GitHub release](https://img.shields.io/github/release/itzg/entrypoint-demoter.svg)](https://github.com/itzg/entrypoint-demoter/releases/latest)

# Motivation

It is  highly recommended that processes within containers do not run as root in order to 
[prevent container breakouts](https://developer.okta.com/blog/2019/07/18/container-security-a-developer-guide#prevent-container-breakouts); however, that
goal is hard to maintain when attached volumes come into play. For example, 
[Synology's DiskStation](https://www.synology.com/en-us/dsm)
manages ownership of attachable volumes as regular UNIX users; however, its Docker integration
doesn't allow for aligning the container's user with what is normally `-u` on the `docker run`
command-line.

`entrypoint-demoter` solves this types of scenarios by allowing the container to momentarily run as 
root, but demote the initial sub-command to either match the ownership of a given path or be 
explicitly defined by `UID`/`GID` environment variables.

# Usage

The command line arguments remaining after the options below are used to execute a 
sub-command with the demoted user ID (uid) and group ID (gid).

If executed as a non-root user, then this tool skips demoting entirely and just executes
the sub-command with the current uid and gid.

## Environment variables

- `UID` : if set, demotes the sub-command to run with the given user ID
- `GID` : if set, demotes the sub-command to run with the given group ID

## Command-line

- `--match PATH` : uses the ownership of the given path to determine a user and group ID
- `--stdin-on-term MESSAGE` : some applications prefer to be gracefully stopped with a command on
   stdin rather than handling SIGTERM, such as Minecraft servers. 
   The `MESSAGE` and a newline will be written to the sub-command's stdin when a `TERM` signal is received.
- `--debug` : enables debug logging
- `--version` : show version and exit

# Example

The following shows how to use `entrypoint-demoter` in a Debian based Dockerfile. The `entry.sh`
would be where your own entry point script/application would be specified.

```Dockerfile
ARG DEMOTER_VERSION=0.1.0
ARG DEMOTER_ARCH=amd64

ADD https://github.com/itzg/entrypoint-demoter/releases/download/${DEMOTER_VERSION}/entrypoint-demoter_${DEMOTER_VERSION}_linux_${DEMOTER_ARCH}.deb /usr/src

RUN dpkg -i /usr/src/entrypoint-demoter_${DEMOTER_VERSION}_linux_${DEMOTER_ARCH}.deb

ENTRYPOINT ["/usr/local/bin/entrypoint-demoter", "--match", "/data", "/entry.sh"]
```

> The full context of this example can be seen [in this Dockerfile](https://github.com/itzg/docker-minecraft-bedrock-server/blob/master/Dockerfile)
