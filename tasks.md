## Pending feature tasks for the project
- [x] Create terraform module in separate repo with example
- [x] Do all S3 stuff in go code. (with option for local mode testing)
- [ ] Logging: switch to zap or similar
- [ ] Secrets management.. including technique for creating envs + their secrets.
- [x] Use localstack to test s3
- [ ] Use localstack to do end to end testing (if possible)
- [ ] Create project overview in README.md
- [ ] Create a CONTRIBUTING.md
- [ ] github repo: better default description
- [ ] github readme: link to docs
- [ ] github repo: set up protected branch + pull request process
- [ ] yaml files: aggressive validation (and run it in PR branches)


## Pending DX tasks
- [ ] temporary files and archives: keep them in memory avoiding weird filesystem bugs/flaky tests.

## Tasks for later not required for MVP:
- [ ] auto-publish new docker image on push to main.
- [ ] Create dorkly org in docker hub
- [ ] AWS credentials in GitHub actions: Use suggested approach described here: https://github.com/marketplace/actions/configure-aws-credentials-action-for-github-actions#overview
- [ ] Consider using https://pkg.go.dev/github.com/sethvargo/go-githubactions or similar
- [ ] Yaml parsing: smarter handling of DataId field (don't store it as 2 fields)
- [ ] Publish binary artifact to Github and consume in Github Action (or publish a github action)