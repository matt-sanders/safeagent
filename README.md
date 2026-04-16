# Safe Claude

Fancy your ssh keys all over the interwebs? Home directory `rm -rf`'d? No? Then put on your tin-foil hat and read on!

The potential attack surface for Claude Code is fairly large as it has access to your whole machine. This means that if Claude suffers a prompt injection attack or perhaps you fat finger a `rm -rf ~/` command accidentally, the repurcussions could be fairly large.

To solve this, we'll put Claude in jail by running it inside Docker.

## Usage

Run `safe-claude` from anywhere to start a container with Claude installed.

### Commands

#### Start a session

```sh
safe-claude
```

#### Change/set the profile

```sh
safe-claude use-profile profile-name
```

#### Exclude directory or file

```sh
safe-claude exclude dirname
```

## Profiles

Running `safe-claude` inside a directory creates a new project. Projects can use different profiles ( e.g. JS with Node LTS or Node 22 ).

### Commands

#### List profiles

```sh
safe-claude profiles
```

#### Create a profile

```sh
safe-claude profiles create
```

#### Remove a profile

```sh
safe-claude profiles rm profile-name
```

## Configuration

The following files can be configured:

- `~/.safe-claude/.claude/settings.json`
- `~/.safe-claude/.claude/CLAUDE.md`
