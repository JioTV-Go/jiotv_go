#!/usr/bin/env bash
set -euo pipefail

# Get the latest tag that looks like a semver tag.
tag=$(git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --abbrev=0 || true)
echo "Latest tag: $tag"
if [[ -z "$tag" || "$tag" == "true" ]]; then
    echo "No semver tag found, use 0.0.0"
    tag="v0.0.0"
fi

# Get the major, minor and patch parts from the tag.
major_minor_patch=$(echo $tag | grep -oE "[0-9]+\.[0-9]+\.[0-9]+")
echo "Major, minor and patch version: $major_minor_patch"

major=$(echo $major_minor_patch | grep -oE "^[0-9]+")
minor=$(echo $major_minor_patch | grep -oE "\.[0-9]+\." | grep -oE "[0-9]+")
patch=$(echo $major_minor_patch | grep -oE "[0-9]+$")
echo "Major version: $major"
echo "Minor version: $minor"
echo "Patch version: $patch"

commits=$(git rev-list $tag.. --count)
echo "Commits since last tag: $commits"

# If any of commit messages contains "BREAKING" string, increment major version.
breaking_changes=$(git log $tag.. --pretty=%B | grep -iE "BREAKING" | wc -l)
echo "Breaking changes: $breaking_changes"

if [[ $breaking_changes -gt 0 ]]; then
    major=$((major + 1))
    minor=0
    patch=0
else
    # Increment minor version.
    features=$(git log $tag.. --pretty=%B | grep -iE "feat|compatibility|integration|upgrade" | wc -l)
    echo "Features: $features"
    if [[ $features -gt 0 ]]; then
        minor=$((minor + 1))
        patch=0
    else
        patch=$((patch + 1))
    fi
fi

# Calculate the new version number.
new_version="v${major}.${minor}.${patch}"
echo "New version: $new_version"
# Update the version in VERSION file
echo $new_version > VERSION
