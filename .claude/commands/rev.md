<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

Rebase onto the latest main, address PR review feedback, and fix failing CI.

## 1. Rebase onto main

1. `git fetch origin main`
2. `git rebase origin/main`
3. If there are conflicts, resolve them, `git add` the resolved files, and
   `git rebase --continue`. Repeat until the rebase completes.

## 2. Address PR review feedback

1. Find the PR for the current branch:
   `gh pr view --json number,url --jq '.number'`
2. Fetch all review comments (use the GraphQL API for threaded context):
   ```
   gh api graphql -f query='
     query($owner:String!, $repo:String!, $pr:Int!) {
       repository(owner:$owner, name:$repo) {
         pullRequest(number:$pr) {
           reviewThreads(first:100) {
             nodes {
               isResolved
               comments(first:50) {
                 nodes { author{login} body path line }
               }
             }
           }
         }
       }
     }' -f owner=cpcloud -f repo=micasa -F pr=<number>
   ```
3. For each **unresolved** thread:
   - Read the referenced file and line to understand the context
   - Make the requested change (or explain in a reply why not)
   - After pushing the fix, reply to the review comment via
     `gh api repos/cpcloud/micasa/pulls/<pr>/comments/<comment_id>/replies`
     explaining how it was addressed (commit hash, what changed)
4. Skip resolved threads -- they need no action.

## 3. Fix failing CI

1. List recent check runs:
   `gh pr checks --json name,state,conclusion,detailsUrl`
2. For each failing job:
   - Fetch the logs: `gh run view <run_id> --log-failed`
   - Diagnose the root cause from the logs
   - Fix the issue in code, config, or dependencies
3. After fixing, run `/pre-commit-check` to verify locally before pushing.

## 4. Push and verify

1. `git push --force-with-lease` (safe force push since we rebased)
2. Wait for CI to start: `gh pr checks --watch --fail-fast`
3. If new failures appear, loop back to step 3.

## 5. Update PR description

After all changes are pushed, re-read the PR title and body
(`gh pr view`) and update them if they no longer match the actual changes.
