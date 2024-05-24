## Pending feature tasks for the project
- [x] project.yml file: remove environments field and instead read in directories in environments folder
- [x] project.yml file: change key to name
- [x] Create terraform module in separate repo with example
- [x] Do all S stuff in go code. (with option for local mode testing)
- [x] Logging: switch to zap or similar
- [ ] Logging: more descriptive log messages + clear indication of steps.
- [x] Secrets: terraform: generate and store in aws secrets manager
- [ ] Secrets: dorkly go code: read from aws secrets manager and inject into each environment's config
- [x] Use localstack to test s3
- [ ] Use localstack to do end to end testing (if possible)
- [ ] Create project overview in README.md
- [ ] Create a CONTRIBUTING.md
- [ ] Terraform: github readme: link to docs
- [ ] Terraform: github repo: set up protected branch + pull request process
- [ ] yaml files validation: warn if flag is defined in the project but not an environment
- [ ] yaml files validation: error if flag is defined in an environment but not in the project
- [ ] yaml files validation: error if env flag type does not match project flag type (ie 'true' for a rollout flag)
- [ ] Terraform: validate variables (see TODOs in https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/variables.tf)
- [ ] Terraform: Consider a command line tool to quickly create new flags and maybe turn them off in all envs
- [ ] Terraform: Allow for optional storing of sdk keys in GitHub (for non-prod environments).
- [ ] Terraform: Save Relay URL in GitHub for easy reference (maybe in readme)
- [ ] Terraform: Autogenerate sdk example snippets including the sdk key and backend service url]
- [ ] Terraform: Use freeform workflow with human input to create flags
- [ ] Terraform: Consider moving all environment creation to freeform github actions


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