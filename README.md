# dorkly

### Process:

1. Load config
2. Load local yaml files
3. Convert to ld structs
4. Fetch/load existing relay archive: either local files (testing), or from S3 (production)
5. Reconcile local files with relay archive. Update versions etc as needed.
6. Marshal json files
7. Create checksum of json files
8. Create ld-relay archive of json + checksum files
9. Save locally (testing) or upload to S3 (production)


### Ideas for later:

1. https://full-stack.blend.com/how-we-write-github-actions-in-go.html#introduction
2. Part of Github actions: connect to relay and await changes.
3. PR check to ensure well-formed yaml files


### Current functionality is limited to a subset of LaunchDarkly's feature flag types.
Here's what is supported:
- One project per git repo. If you need more projects create more repos.
- Boolean flags: either on or off, or a percent rollout based on user id
- Only the user context kind is supported
- server-side flags and client-side flags (can exclude client-side on a per-flag basis)

Feature ideas for later:
1. Segments
2. String variation flags
3. Number variation flags
4. Json variation flags
