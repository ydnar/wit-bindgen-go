# Release

This document describes the steps to release a new version of module `go.bytecodealliance.org/cm`.

## 1. Update [CHANGELOG.md](./CHANGELOG.md)

1. Add the latest changes to [CHANGELOG.md](./CHANGELOG.md).
1. Rename the Unreleased section to reflect the new version number.
	1. Update the links to new version tag in the footer of CHANGELOG.md
1. Add today’s date (YYYY-MM-DD) after an em dash (—).
1. Submit a [GitHub Pull Request](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests) with these updates.

## 2. Create a new release

Once the PR is merged, tag the new version in Git and push the tag to GitHub.

**Note:** the tag **must** start with the prefix `cm/` in order to correctly tag this module.

For example, to tag version `cm/v0.2.0`:

```console
git tag cm/v0.2.0
git push origin cm/v0.2.0
```

## 3. Update the root module

Once the tag is pushed, you can update the root module to depend on the newly created version of package `cm` by running the following:

```console
go get -u go.bytecodealliance.org/cm@latest
```

Then follow the instructions in [RELEASE.md](../RELEASE.md) to release a new version of the root module.
