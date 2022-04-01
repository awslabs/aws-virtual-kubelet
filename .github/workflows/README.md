# GitHub Actions Workflows
This directory contains workflows ran via [GitHub Actions](https://docs.github.com/en/actions).

## Overview
Workflows and their purpose are described in the sections below, followed by setup instructions.

### Validation
This workflow ensures a basic level of sanity exists in the code being pushed.

### Coverage
This workflow ensures that pull requests meet minimum coverage requirements.

## Setup
Some setup is required for these workflows to run properly.

### Gist Storage
The workflows use a secret [gist](https://gist.github.com/) for lightweight file storage.  This gist must be created manually via the following steps.

1. Browse to https://gist.github.com/ (the _new gist_ interface should appear)
2. Enter a description such as `vk-badges`
3. Enter a filename (e.g. `README.md`)
4. Enter content for the file (e.g. `gist for storing vk badge data files`)
5. Click _Create secret gist_
6. Copy the gist id from the URL (we'll need it in the next section)

### Gist Access Token
Github requires an authentication token to access the gist.  Follow the steps below to create the token.
**`NOTE`** The token must belong to a user account

1. Browse to https://github.com/settings/tokens
2. Click _Generate new token_ (entering your password if prompted)
3. Enter a note (e.g. `vk-badges gist`)
4. Set an expiration (and create a calendar invite for yourself and those affected by this token's expiration)
5. Check the _gist_ scope
6. Click _Generate token_ and save a copy of the token shown

### GitHub Secrets
The following secret values must be set.

#### BADGE_GIST_ID
Set this to the _gist id_ copied above

#### BADGE_GIST_TOKEN
Set this to the _gist access token_ created above

