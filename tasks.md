## Pending feature tasks for the project
- [x] project.yml file: remove environments field and instead read in directories in environments folder
- [x] project.yml file: change key to name
- [x] Create terraform module in separate repo with example
- [x] Do all S3 stuff in go code. (with option for local mode testing)
- [x] Logging: switch to zap or similar
- [x] Logging: more descriptive log messages + clear indication of steps.
- [x] Secrets: terraform: generate and store in aws secrets manager
- [x] Secrets: dorkly go code: read from aws secrets manager and inject into each environment's config
- [x] Use localstack to test s3
- [x] Terraform: generated env-level readme should contain: general tips on connecting using ld sdk + injected url, sdk keys, etc for each env.
- [...] Terraform: generated repo readme should contain: links to docs, summary of environments.
- [...] Create project overview in README.md
- [x] Terraform: Add option per-env: Disable checking in sdk keys. (for production environments).
- [ ] Good docs on client-side (js) sdk setup.
- [ ] Create a CONTRIBUTING.md
- [ ] Terraform: github readme: link to docs
- [ ] Terraform: github repo: set up protected branch + pull request checks
- [ ] yaml files validation: warn if flag is defined in the project but not an environment
- [ ] yaml files validation: error if flag is defined in an environment but not in the project
- [ ] yaml files validation: error if env flag type does not match project flag type (ie 'true' for a rollout flag)
  MVP ? ==============================================
- [ ] Consider enabling configuring ld-relay client context (aka goals endpoint): https://github.com/launchdarkly/ld-relay/blob/1adf0dde5b11343d3bdf011c86e3f7116c4960fc/internal/relayenv/js_context.go#L7
- [ ] Terraform: validate variables (see TODOs in https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/variables.tf)
- [ ] Terraform: Consider a command line tool to quickly create new flags and maybe turn them off in all envs
- [ ] Terraform: Use freeform workflow with human input to create flags
- [ ] Use localstack to do end to end testing (if possible)
- [ ] Maybe never: Implement mobile key and client-side sdk setup (for now people should just use the client id)


## Pending DX tasks
- [ ] temporary files and archives: keep them in memory avoiding weird filesystem bugs/flaky tests.

## Tasks maybe not required for MVP:
- [ ] Document handling of deleted environments... what happens to the flags?
- [ ] Environment: Specify production or non-production (for displaying secrets or keeping them locked up in aws secrets manager)
- [ ] auto-publish new docker image on push to main.
- [ ] Create dorkly org in docker hub
- [ ] AWS credentials in GitHub actions: Use suggested approach described here: https://github.com/marketplace/actions/configure-aws-credentials-action-for-github-actions#overview
- [ ] Consider using https://pkg.go.dev/github.com/sethvargo/go-githubactions or similar
- [ ] Yaml parsing: smarter handling of DataId field (don't store it as 2 fields)
- [ ] Maybe: Publish binary artifact to Github and consume in Github Action (or publish a github action)
- [ ] Maybe: thin wrappers around SDKs that handle custom url and disabling of events (this is potentially a maintenance headache)