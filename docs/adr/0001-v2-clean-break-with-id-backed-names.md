# v2 is a clean break, with Profiles and Stacks backed by hidden IDs

v2 renames the core nouns: v1's "auth identity" becomes the Profile, and v1's "profile" becomes the Stack. Migrating v1 data automatically would mean reinterpreting the same words with swapped meanings, so v2 stores everything under `~/.safeagent/v2/` and doesn't read or migrate v1 data. Users set up again, or move things across by hand.

Profiles and Stacks get a generated ID that's used everywhere internally (Project references, Image tags like `safeagent/<stack-id>:<profile-id>`, on-disk folders like `profiles/<id>/`). Users only ever type and see the lowercase Name. Renaming isn't supported yet, but because nothing outside the Profile or Stack record stores the Name, it can be added later without rewriting references or rebuilding Images.

## Considered Options

- **Automatic migration** (as v1's `MigrateAuthLayout` did): rejected because the noun swap makes it error-prone, and v1 data is easy to recreate.
- **Names as the only identifier**: simpler, but it would put Names in Image tags and folder paths, making a future rename expensive.
