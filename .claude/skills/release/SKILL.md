# OpenSkills Release

One-click release workflow: bump version, tag, push, and publish to Homebrew via GoReleaser + GitHub Actions.

## Trigger

- User says "发版", "release", "发布新版本", "bump version"
- User runs `/release`

## Prerequisites

1. Git repo with remote configured (`git remote -v` shows `origin`)
2. GitHub repo secrets configured:
   - `GITHUB_TOKEN` — auto-provided by GitHub Actions, no manual setup needed
   - `HOMEBREW_TAP_TOKEN` — PAT with Contents (Read and write) permission on `lovelyJason/openskills`
3. `lovelyJason/openskills` repo exists on GitHub (it's the same repo as the source code)

## Workflow

### Step 1: Pre-flight checks

```bash
# Verify clean working tree
git status --porcelain

# Verify on main branch
git branch --show-current

# Verify tests pass
go test -race ./...

# Verify build works
go build -o /dev/null ./cmd/openskills
```

If working tree is dirty, ask user to commit first.
If not on main, ask user to switch or confirm.

### Step 2: Determine version

Check existing tags:
```bash
git tag --sort=-v:refname | head -10
```

Ask user what version to release. Suggest based on:
- **patch** (v0.1.0 → v0.1.1): bug fixes
- **minor** (v0.1.0 → v0.2.0): new features
- **major** (v0.1.0 → v1.0.0): breaking changes

If no tags exist, suggest `v0.1.0`.

### Step 3: Tag and push

```bash
# Create annotated tag
git tag -a v<VERSION> -m "Release v<VERSION>"

# Push tag (this triggers the release workflow)
git push origin v<VERSION>
```

### Step 4: Monitor release

Tell user:
- GitHub Actions `release.yaml` is now running
- GoReleaser will:
  1. Run tests (with `-race`)
  2. Build binaries for darwin/linux/windows × amd64/arm64 (5 targets, windows/arm64 excluded)
  3. Create GitHub Release with changelog (`.tar.gz` for macOS/Linux, `.zip` for Windows)
  4. Update Homebrew formula in `lovelyJason/openskills` repo (Formula/ directory)

Direct user to check: `https://github.com/lovelyJason/openskills/actions`

### Step 5: Verify (after Actions complete)

```bash
# Verify GitHub Release exists
gh release view v<VERSION>

# Verify Homebrew formula updated (first time: add tap with explicit URL)
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew update && brew info lovelyJason/openskills/openskills
```

## First-time setup instructions

If this is the first release ever, guide user through:

1. The `lovelyJason/openskills` repo already serves as both source and Homebrew tap.
   Ensure it has a `Formula/` directory (already present in this repo).

2. Create a GitHub PAT (fine-grained) with:
   - Repository access: `openskills` only
   - Permissions: Contents (Read and write)

3. Add the PAT as a secret to the `openskills` repo:
```bash
gh secret set HOMEBREW_TAP_TOKEN --repo lovelyJason/openskills
```

4. Now run the release steps above.

## Homebrew install commands

Since the repo is named `openskills` (not `homebrew-openskills`), users must tap with an explicit URL:

```bash
# First time
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew install openskills

# Or one-liner
brew install lovelyJason/openskills/openskills
```

## Quick reference

```bash
# Full release one-liner (after confirming version):
git tag -a v0.1.0 -m "Release v0.1.0" && git push origin v0.1.0
```
