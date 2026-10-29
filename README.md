# Safe Agent

Fancy your ssh keys all over the interwebs? Home directory `rm -rf`'d? No? Then put on your tin-foil hat and read on!

The potential attack surface for agentic coding is fairly large as they often have access to your whole machine. This means that if Claude suffers a prompt injection attack or perhaps you fat finger a `rm -rf ~/` command accidentally, the repurcussions could be fairly large.

To solve this, we'll put the agent in jail by running it inside Docker. It can only touch what it's explicitly given.

> [!NOTE]
> This was built entirely with Claude. It's a personal project and I just need it to work. Code quality is likely poor, as I just needed something up and running quickly.

## Core concepts

A session is made up of three things: **who** the agent is (a Profile), **what** software it has (a Stack), and **where** it works (a Project).

**Profile**: The "who" of a session. Who the agent is: its Claude login, plugins and settings, git credentials and any other personal setup. Each Profile has its own Claude config folder, so your personal and work setups never mix. Create one Profile per Claude login (e.g. `personal` and `work`).

**Stack**: The "what" of a session. The software and tooling installed in the environment the agent runs in. Every Stack has a node version (default `lts`), fixed once the Stack is created. On top of that, every Stack comes with python3, uv, nvm, yarn, pnpm, gh and Claude Code.

**Project**: The "where" of a session. A host directory safeagent knows about, along with the Profile and Stack it uses. Running `safeagent` anywhere inside a Project uses that Project. If Projects are nested, the innermost one wins.

**Image**: The runnable combination of one Stack and one Profile. A Project's Stack and Profile together decide which Image a session runs. You never have to build Images by hand: safeagent builds one the first time a session needs it. On the command line an Image is written `<stack>:<profile>`, e.g. `node22:work`.

**Name**: The short, lowercase handle you type to refer to a Profile or Stack. Names can only use lowercase letters, digits, `-` and `_`. A Name is unique among Profiles, and among Stacks, but a Stack and a Profile can share a Name.

**Session**: One running container of an Image, working inside a Project.

**Mount**: A host path that a Profile makes available inside every Session using that Profile. Handy for things like your gh config or SSH keys.

**Exclusion**: A path inside a Project that's hidden from Sessions, e.g. `node_modules` or `build`.

## v2 is a clean break

v2 renamed the core nouns, so it doesn't read or migrate any data from older versions of safeagent. Everything v2 knows about lives under `~/.safeagent/v2/`:

```
~/.safeagent/v2/
  profiles.json
  stacks.json
  projects.json
  profiles/<id>/claude/       # the Profile's Claude config, mounted at /claude
  profiles/<id>/Dockerfile    # the Profile's Dockerfile fragment
  stacks/<id>/Dockerfile      # the Stack's Dockerfile fragment
```

If you used an older version, set things up again (or copy files across by hand). Old data and images are left alone. See [ADR 0001](docs/adr/0001-v2-clean-break-with-id-backed-names.md) for the reasoning.

Profiles and Stacks also have a generated ID, shown in `profile ls` and `stack ls` and used in folder names and Image tags (`safeagent/<stack-id>:<profile-id>`). You only ever type Names.

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

Run `safeagent` from anywhere inside a Project to start a Session. If there's no Project for the current directory, safeagent creates one there and asks you to pick a Profile and a Stack. Both pickers offer "Create new…", so a first run can set everything up in one go. safeagent never picks a Profile or Stack for you, even when only one exists.

The Session mounts:

- the Project directory, at the same path inside the container
- the Profile's Claude config folder, at `/claude`
- the Profile's Mounts
- the Project's Exclusions, hidden as empty volumes (relative to the directory you started in)

The first Session that uses a new Profile runs Claude Code's normal `/login` flow. The login is saved in that Profile, so you only log in once per Profile. Credentials are always mounted, never built into an Image.

> [!NOTE]
> See the section on [git worktrees](#git-worktrees) below if you use them.

Every listing is a plain table, like `docker ps`, so it pipes cleanly into other tools. Anything you're prompted for can also be passed as a flag or argument, so every command can run without interaction.

### Start a session

```sh
safeagent [--shell] [--profile <name>] [--stack <name>]
```

- `--shell`: open a zsh shell in the container instead of Claude.
- `--profile`, `--stack`: the Profile and Stack for a **new** Project, skipping the pickers. Passing them when the Project already exists is an error. Use `project use-profile` / `project use-stack` to switch instead.

If the Project's Profile or Stack has since been removed, safeagent tells you which one and how to fix it.

## Profiles

```sh
safeagent profile ls                          # NAME, ID, MOUNTS, PROJECTS
safeagent profile create [--name <name>]      # prompts for the Name if omitted
safeagent profile rm <name>
safeagent profile dockerfile <name>
safeagent profile rebuild <name>
```

`profile rm` refuses while any Project uses the Profile, and lists those Projects. Otherwise it removes the Profile, its folder (including its Claude login) and every Image built with it. There's no confirmation, so be sure.

`profile dockerfile` opens the Profile's Dockerfile fragment (`~/.safeagent/v2/profiles/<id>/Dockerfile`) in `$EDITOR` (falls back to `vim`). Use it for Profile-specific setup such as global git config or plugin installs. Its instructions are added to the Image build after the Stack's fragment and before `ENTRYPOINT`. When you're done, you'll be offered a rebuild (in a terminal). Otherwise, run `profile rebuild`.

`profile rebuild` rebuilds only the Images that already exist for that Profile. It doesn't build pairings you've never used.

### Mounts

```sh
safeagent profile mounts ls <name>                                     # HOST PATH, CONTAINER PATH
safeagent profile mounts add <name> [--host <path>] [--container <path>]
safeagent profile mounts rm <name> <container-path>
```

`mounts add` prints the Profile's folder on the host for reference (a good place to keep Profile-owned files you want to mount), then prompts for whichever of `--host` and `--container` you left out. The host path must exist. The container path must be absolute, can't be `/claude`, and can't already be used by another of the Profile's Mounts.

A Mount whose container path is the same as the Project's path stops the Session from starting, with an error explaining why.

For example, to give a `work` Profile your gh login:

```sh
safeagent profile mounts add work --host ~/.config/gh --container /home/developer/.config/gh
```

## Stacks

```sh
safeagent stack ls                                      # NAME, ID, NODE, PROJECTS
safeagent stack create [--name <name>] [--node <version>]
safeagent stack rm <name>
safeagent stack dockerfile <name>
safeagent stack rebuild <name>
```

`stack create` prompts for anything you leave out, with the node version prefilled as `lts`. The node version can't be changed later, so a Stack always means the same environment. Create a new Stack instead.

`stack rm` refuses while any Project uses the Stack, and lists those Projects. Otherwise it removes the Stack, its folder and every Image built with it.

`stack dockerfile` opens the Stack's Dockerfile fragment (`~/.safeagent/v2/stacks/<id>/Dockerfile`) in `$EDITOR` (falls back to `vim`), so you can install extra software. Its instructions are added to the Image build before the Profile's fragment. As with Profiles, you'll be offered a rebuild afterwards.

`stack rebuild` rebuilds only the Images that already exist for that Stack.

## Projects

Every Project command works on the Project for the current directory (the closest one above it). Pass `-p <path>` to work on any Project from anywhere. The path is resolved relative to the current directory.

```sh
safeagent project ls                                                  # PATH, PROFILE, STACK, IMAGE
safeagent project inspect [-p <path>]
safeagent project create [-p <path>] [--profile <name>] [--stack <name>]
safeagent project rm [-p <path>]
safeagent project use-profile <name> [-p <path>]
safeagent project use-stack <name> [-p <path>]
```

`project inspect` shows the Project's path, Profile, Stack, whether its Image is built, the Mounts a Session will get and its Exclusions, so you can check a setup without starting a Session.

`project create` creates a Project in the current directory (or at `-p`) without starting a Session. It prompts for the Profile and Stack if you don't pass them, and refuses if a Project already exists at exactly that path. A Project nested inside another is fine.

`project rm` only makes safeagent forget the Project. Nothing on disk is touched.

`use-profile` and `use-stack` switch the Project's Profile or Stack. The new Image is built at the next Session if it doesn't exist yet.

### Exclusions

Sometimes you will want to hide certain directories from the container. e.g. you might want to exclude your `build` directory so Claude doesn't pollute local builds, or `node_modules` so that you can install modules specific to the container.

```sh
safeagent project exclusions ls [-p <path>]
safeagent project exclusions add [<path>] [-p <path>]     # prompts for the path if omitted
safeagent project exclusions rm <path> [-p <path>]
```

Exclusion paths must be relative and stay inside the directory (no absolute paths or `..`). They're cleaned when stored, so `./node_modules` and `node_modules/` are both stored as `node_modules`.

> [!IMPORTANT]
> Exclusions are relative to the directory the Session **starts** in, not the Project root. See [git worktrees](#git-worktrees) for why.

## Images

Images are named `<stack>:<profile>` using Names.

```sh
safeagent image ls                                    # IMAGE, STACK, PROFILE, PROJECTS
safeagent image rebuild <stack>:<profile>
safeagent image rebuild --all
safeagent image rm <stack>:<profile>
```

`image ls` lists only the Images that have actually been built. An Image whose Stack or Profile has gone is shown with the raw ID instead of a Name.

`image rebuild` rebuilds one Image (building it even if it doesn't exist yet), or every built Image with `--all`. Pass exactly one of the two. Updating safeagent doesn't update your Images, so run `safeagent image rebuild --all` after updating to pick up the new base image. `--all` skips (and reports) Images whose Stack or Profile no longer exists.

`image rm` removes a built Image to reclaim disk space, and lists the Projects that will rebuild it at their next Session. It errors if the Image isn't built.

## Sessions

```sh
safeagent session ls          # CONTAINER, PROJECT, PROFILE, STACK, DIRECTORY, RUNNING FOR
```

Lists running safeagent Sessions only, so other Docker containers don't clutter it. DIRECTORY is the directory the Session started in.

## Configuration

Each Profile's Claude config folder (`~/.safeagent/v2/profiles/<id>/claude`) is mounted at `/claude` in every Session using that Profile. This is where the Profile's login, plugins and settings persist across Sessions. Use `safeagent profile ls` to find a Profile's ID.

Common files you may want to configure (per Profile):

- `~/.safeagent/v2/profiles/<id>/claude/settings.json`
- `~/.safeagent/v2/profiles/<id>/claude/CLAUDE.md`

Settings, plugins and `CLAUDE.md` are **not** shared between Profiles (yet).

## Git Worktrees

Git worktrees can get kinda crazy. If you want safeagent to work properly with worktrees, create the Project **at the repo root that holds your worktrees**. Starting a Session from any worktree below it then uses that Project. But note that the whole Project directory (every worktree) will be mounted.

For example, if you have the following structure:

```
my-repo/
 - branch-1/
 - branch-2/
```

You would do the following to start a Session in `branch-1`:

```sh
cd my-repo
safeagent project create --profile work --stack node22
cd branch-1
safeagent
```

### Exclusions

Exclusions are applied relative to the directory the Session starts in, not the Project root. So with the layout above, adding a `node_modules` Exclusion:

```sh
cd my-repo
safeagent project exclusions add node_modules
```

hides `branch-1/node_modules` when you start a Session from `my-repo/branch-1`, and `branch-2/node_modules` when you start from `my-repo/branch-2`. One Exclusion list works for every worktree, as long as you start each Session from the worktree's root. Start a Session from somewhere else (e.g. `my-repo/branch-1/src`) and the Exclusion points at `src/node_modules` instead.
