# Dorkly Flags 'Backend' (in quotes because it's mostly just a GitHub Action)
Not sure what's going on here? Check out the project overview [here](https://github.com/dorklyorg)

This repo contains the backend bits for the Dorkly Flags project:
1. Go code that runs in GitHub Actions to convert human-friendly yaml files to a format that can be consumed by the ld-relay appliance.
2. Dockerfile to build the image used in the deployed backend service.

## Everything below here is a work in progress and is probably not accurate.

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



### Go code
The go code in this repo runs in GitHub Action provisioned by the [dorklyorg/dorkly-flags Terraform module](https://registry.terraform.io/modules/dorklyorg/dorkly-flags/aws/latest).

### Docker
The Dockerfile is used to build the image used in the deployed backend service. It is built on top of the ld-relay image.
To build and publish it: (requires docker login with permissions to push to drichelson)
```bash
TAG=0.0.1 docker build --platform=linux/amd64 -t drichelson/dorkly:$TAG ./docker/ && docker push drichelson/dorkly:$TAG
```

### Ideas for later:
1. Create GitHub releases with human-readable diffs of flag changes
2. https://full-stack.blend.com/how-we-write-github-actions-in-go.html#introduction
2. Part of GitHub actions: connect to relay and await changes.
3. PR check to ensure well-formed yaml files
4. Publish a GitHub action? then we don't need to check in a binary
5. Send Slack notifications on changes.


### Current functionality is limited to a subset of LaunchDarkly's feature flag types.
Here's what is supported:
- One project per git repo. If you need more projects create more repos.
- Boolean flags: either on or off, or a percent rollout based on user id
- Only the user context kind is supported
- server-side flags and client-side flags (can exclude client-side on a per-flag basis)

Ideas for later later
1. Segments
2. String variation flags
3. Number variation flags
4. Json variation flags
