# Safe Agent

Fancy your ssh keys all over the interwebs? Home directory `rm -rf`'d? No? Then put on your tin-foil hat and read on!

The potential attack surface for agentic coding is fairly large as they often have access to your whole machine. This means that if Claude suffers a prompt injection attack or perhaps you fat finger a `rm -rf ~/` command accidentally, the repurcussions could be fairly large.

To solve this, we'll put the agent in jail by running it inside Docker.

## Usage

Run `safeagent` from anywhere to start a containerised instance of your agent.

### Commands

#### Start a session

```sh
safeagent
```

#### Change/set the profile

```sh
safeagent use-profile
```

#### Exclude directory or file

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

```sh
safeagent profiles rebuild <profile-id>
```

## Configuration

The following files can be configured:

- `~/.safeagent/.claude/settings.json`
- `~/.safeagent/.claude/CLAUDE.md`
