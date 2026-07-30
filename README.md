# Safe Agent

Fancy your ssh keys all over the interwebs? Home directory `rm -rf`'d? No? Then put on your tin-foil hat and read on!

The potential attack surface for agentic coding is fairly large as they often have access to your whole machine. This means that if Claude suffers a prompt injection attack or perhaps you fat finger a `rm -rf ~/` command accidentally, the repurcussions could be fairly large.

To solve this, we'll put the agent in jail by running it inside Docker.

> [!NOTE]
> This was built entirely with Claude. It's a personal project and I just need it to work. Code quality is likely poor, as I just needed something up and running quickly.

## Core concepts

**Profile**: A profile is essentially a docker image built with certain configurations or tools built into it. This allows the same image to be used across multiple projects. For example, you may have a Node22 profile or a Node24 profile. Currently safeagent comes installed with python3, uv, and nvm. When creating a profile, you'll be asked to select a node version.

**Project**: A project is a set of configurations for a particular directory. These are things like the profile safeagent uses, any directory exclusions, etc. Whenever you start safeagent for the first time in a directory, it will ask you to create a project and link it to a profile.

**Auth identity**: An auth identity is a separate Claude Code login (its own credentials, settings, plugins, and history), kept in its own directory. Projects pick an identity so you can switch between different Claude logins (e.g. personal vs work) without logging in and out. The `default` identity always exists.

## Build & Installation

You'll need to build safeagent before you can use it.

```
mkdir bin
go build -o bin/safeagent
```

You can then run it my ensuring the path to safeagent is in your `PATH`. Options are:

- Adding the binary to your `PATH` ( e.g. in your `.zshrc` or equivalent )
- Move the binary to somewhere already in your path ( e.g. `mv bin/safeagent ~/.local/bin/` )
- Symlink it

## Usage

Run `safeagent` from anywhere to start a containerised instance of your agent inside the cwd. The cwd will be mounted to a docker container.

> [!NOTE]
> See the section on [git worktrees](#git-worktrees) below if you use them.

### Commands

#### Start a session

```sh
safeagent
```

#### Create a project in the cwd without starting a session

```sh
safeagent project create
```

#### Change/set the profile

```sh
safeagent use-profile
```

#### Exclude directory or file

Sometimes you will want to exclude certain directories from being mounted into the container. e.g. you might want to exclude your `build` directory so Claude doesn't pollute local builds, or `node_modules` so that you can install modules specific to the container.

NOTE: Exclusions work slightly differently with git worktrees. See the [git worktrees](#git-worktrees) documentation below.

```sh
safeagent exclude
```

## Profiles

Running `safeagent` inside a directory creates a new project. Projects can use different profiles ( e.g. JS with Node LTS or Node 22 ).

### Commands

#### List profiles

```sh
safeagent profiles
```

#### Create a profile

```sh
safeagent profiles create
```

#### Remove a profile

```sh
safeagent profiles rm <profile-name>
```

#### Rebuild a profile

This rebuilds the image used for the profile. At the moment, updating safeagent won't update your images, so you may want to run this after updating safeagent.

```sh
safeagent profiles rebuild <profile-id>
```

## Authentication

Each project uses an **auth identity** — an isolated Claude Code login. You pick one when a project is created (defaulting to `default`), and can change it any time.

Identities live in `~/.safeagent/auth/<name>/.claude` and are mounted into the container at `/claude`. Because each identity is a full, independent config directory, settings, plugins, and `CLAUDE.md` are **not** shared between identities (yet).

The first session that uses an identity with no saved login runs Claude Code's normal `/login` flow; the credentials then persist for that identity.

### Commands

#### List identities

```sh
safeagent auth
```

#### Create an identity

```sh
safeagent auth create
```

#### Remove an identity

```sh
safeagent auth rm <name>
```

#### Change the identity for the current project

```sh
safeagent use-auth
```

## Configuration

The auth identity's directory (`~/.safeagent/auth/<name>/.claude`) is mounted into each container at `/claude`. This is where per-identity plugins, settings, and login state are persisted across sessions. Upgrading from an older safeagent automatically moves your existing `~/.safeagent/.claude` to the `default` identity.

Common files you may want to configure (per identity):

- `~/.safeagent/auth/<name>/.claude/settings.json`
- `~/.safeagent/auth/<name>/.claude/CLAUDE.md`

## Git Worktrees

Git worktrees can get kinda crazy. If you want safeagent to work properly with worktrees, create the project **at the worktree root**. Every subdirectory will use that worktree. But note that the whole worktree will be mounted.

For example, if you have the following structure:

```
my-repo/
 - branch-1/
 - branch-2/
```

You would do the following to start a safeagent session in `branch-1`:

```sh
cd my-repo
safeagent project create
cd branch-1
safeagent
```

### Exclusions

If you add exclusions to the project, these will be exluded for the current worktree only ( better support for this coming soon ).
