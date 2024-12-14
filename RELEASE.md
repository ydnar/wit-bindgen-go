# Release

This document describes the steps to release a new version of the `wit-bindgen-go` CLI.

## 1. Prerequisites

If package `cm` has changed, make the neccessary updates to [cm/CHANGELOG.md](./cm/CHANGELOG.md) and create a new release for `go.bytecodealliance.org/cm`. Then update the dependency in this module by running:

```console
go get -u go.bytecodealliance.org/cm@latest
```

Commit those changes prior to tagging a new release of this module.

## 2. Update [CHANGELOG.md](./CHANGELOG.md)

1. Add the latest changes to [CHANGELOG.md](./CHANGELOG.md).
1. Rename the Unreleased section to reflect the new version number.
	1. Update the links to new version tag in the footer of CHANGELOG.md
1. Add today’s date (YYYY-MM-DD) after an em dash (—).
1. Submit a [GitHub Pull Request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests) with these updates.

## 3. Create a new release

Once the PR is merged, tag the new version in Git and push the tag to GitHub.

For example, to tag version `v0.3.0`:

```console
git tag v0.3.0
git push origin v0.3.0
```

After the tag is pushed, GitHub Actions will automatically create a new release with the content of the `CHANGELOG.md` file.
