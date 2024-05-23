## Pending feature tasks for the project
- [x] project.yml file: remove environments field and instead read in directories in environments folder
- [x] project.yml file: change key to name
- [x] Create terraform module in separate repo with example
- [x] Do all S stuff in go code. (with option for local mode testing)
- [x] Logging: switch to zap or similar
- [x] Logging: more descriptive log messages + clear indication of steps.
- [x] Secrets: terraform: generate and store in aws secrets manager
- [ ] Secrets: dorkly go code: read from aws secrets manager
- [x] Use localstack to test s3
- [ ] Use localstack to do end to end testing (if possible)
- [ ] Create project overview in README.md
- [ ] Create a CONTRIBUTING.md
- [ ] github repo: better default description
- [ ] github readme: link to docs
- [ ] github repo: set up protected branch + pull request process
- [ ] yaml files validation: warn if flag is defined in the project but not an environment
- [ ] yaml files validation: error if flag is defined in an environment but not in the project
- [ ] yaml files validation: error if env flag type does not match project flag type (ie 'true' for a rollout flag)
- [ ] terraform: validate variables (see TODOs in https://github.com/dorklyorg/terraform-aws-dorkly-flags/blob/main/variables.tf)
- 


## Pending DX tasks
- [ ] temporary files and archives: keep them in memory avoiding weird filesystem bugs/flaky tests.

## Tasks for later not required for MVP:
- [ ] auto-publish new docker image on push to main.
- [ ] Create dorkly org in docker hub
- [ ] AWS credentials in GitHub actions: Use suggested approach described here: https://github.com/marketplace/actions/configure-aws-credentials-action-for-github-actions#overview
- [ ] Consider using https://pkg.go.dev/github.com/sethvargo/go-githubactions or similar
- [ ] Yaml parsing: smarter handling of DataId field (don't store it as 2 fields)
- [ ] Publish binary artifact to Github and consume in Github Action (or publish a github action)