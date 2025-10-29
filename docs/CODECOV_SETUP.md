# Codecov Setup

## Problem: Codecov shows "unknown" status

If Codecov badge shows "unknown" or coverage is not being uploaded, it's because the `CODECOV_TOKEN` secret is not configured in GitHub.

## Solution

### 1. Get Codecov Token

1. Go to [codecov.io](https://codecov.io)
2. Log in with your GitHub account
3. Find your repository: `cherrypick-agency/flutter_network_debugger`
4. Go to Settings → Repository Upload Token
5. Copy the token

### 2. Add Token to GitHub Secrets

1. Go to your GitHub repository: https://github.com/cherrypick-agency/flutter_network_debugger
2. Click on **Settings** (repository settings, not your account)
3. In the left sidebar, click **Secrets and variables** → **Actions**
4. Click **New repository secret**
5. Name: `CODECOV_TOKEN`
6. Value: Paste the token from Codecov
7. Click **Add secret**

### 3. Verify

After adding the token, the next push/PR will automatically upload coverage to Codecov.

You can verify by:
1. Pushing a new commit
2. Waiting for the "coverage" workflow to complete
3. Checking the workflow logs for "Process Upload complete" without errors
4. Checking Codecov dashboard for updated coverage

## Alternative: Tokenless Upload (Not Recommended for Protected Branches)

If you don't want to use a token, you can:
1. Make the repository public (not recommended)
2. Or unprotect the `main` branch (not recommended)

For protected branches (which is a good security practice), a token is **required**.

## Workflow Files

The following workflows upload coverage to Codecov:
- `.github/workflows/coverage.yml` - Unit tests coverage
- `.github/workflows/go_build_test.yml` - Go tests with race detector

Both now include `token: ${{ secrets.CODECOV_TOKEN }}` for proper authentication.
