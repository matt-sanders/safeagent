# safeagent

safeagent runs a coding agent inside a Docker container so it can only touch what it's explicitly given.

## Language

**Profile**:
The "who" of a session: the agent's identity, including its Claude login, plugins and settings, git credentials and other personal setup.
_Avoid_: Auth, auth identity, identity

**Stack**:
The "what" of a session: the software and tooling installed in the environment the agent runs in.
_Avoid_: Profile (its v1 meaning), environment

**Project**:
The "where" of a session: a host directory safeagent knows about, along with the Profile and Stack it uses.
_Avoid_: Workspace, repo

**Image**:
The runnable combination of one Stack and one Profile. A Project's Stack and Profile together determine which Image a session runs.
_Avoid_: Container, build

**Name**:
The short, lowercase handle a user types to refer to a Profile or Stack. It is unique among Profiles, and among Stacks.
_Avoid_: Slug, label, alias

**Session**:
One running container of an Image, working inside a Project.
_Avoid_: Instance, run

**Mount**:
A host path that a Profile makes available inside every Session using that Profile, or that a Project makes available inside its own Sessions.
_Avoid_: Volume, bind

**Exclusion**:
A path inside a Project that is hidden from Sessions.
_Avoid_: Ignore, exclude rule
