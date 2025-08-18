#!/usr/bin/env bash
set -euo pipefail

# Get the latest tag that looks like a semver tag.
tag=$(git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --abbrev=0 || true)
echo "Latest tag: $tag"
if [[ -z "$tag" ]]; then
    echo "No semver tag found, use 0.0.0"
    tag="v0.0.0"
fi

# Get the major, minor and patch parts from the tag.
major_minor_patch=$(echo "$tag" | grep -oE "[0-9]+\.[0-9]+\.[0-9]+")
echo "Major, minor and patch version: $major_minor_patch"

IFS='.' read -r major minor patch <<< "$major_minor_patch"
echo "Major version: $major"
echo "Minor version: $minor"
echo "Patch version: $patch"

commits=$(git rev-list $tag.. --count)
echo "Commits since last tag: $commits"

# If any of commit messages contains "BREAKING" string, increment major version.
commit_messages=$(git log --oneline "$tag..")
breaking_changes=$(echo "$commit_messages" | grep -i "BREAKING" | wc -l)
echo "Breaking changes (commits): $breaking_changes"

if [[ $breaking_changes -gt 0 ]]; then
    major=$((major + 1))
    minor=0
    patch=0
else
    # Increment minor version.
    features=$(echo "$commit_messages" | grep -i "feat" | wc -l)
    echo "Features (commits): $features"
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
