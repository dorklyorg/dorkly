# dorkly
Current status: not ready for public consumption. Stay tuned.

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

### Terraform
Requirements:
1. AWS credentials with permissions to create S3 buckets, Lightsail containers, SQS queues, and IAM roles.
2. Github token with permissions to create repos, secrets, and actions.
3. Terraform installed on your machine.

To run the terraform you need to have both AWS and Github credentials. Here's an example:
```bash
AWS_PROFILE=<aws profile> GITHUB_TOKEN=<github token allowing repo creation etc> terraform apply
```

### Go code
The go code will run in Github Actions so it need to be built for linux. Here's how to update the binary used by terraform:
```bash
GOOS=linux GOARCH=amd64 go build -o ./terraform/files/.github/workflows/dorkly ./cmd/dorkly
```

### Docker
The Dockerfile is used to build the image used in the deployed backend service. It is built on top of the ld-relay image.
To build and publish it: (requires docker login with permissions to push to drichelson)
```bash
docker build --platform=linux/amd64 -t drichelson/dorkly:latest ./docker/ && docker push drichelson/dorkly:latest
```

### Ideas for later:
1. Create Github releases with human-readable diffs of flag changes
2. https://full-stack.blend.com/how-we-write-github-actions-in-go.html#introduction
2. Part of Github actions: connect to relay and await changes.
3. PR check to ensure well-formed yaml files
4. Publish a Github action? then we don't need to check in a binary
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
