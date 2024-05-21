## Pending feature tasks for the project
- [x] Create terraform module in separate repo with example
- [ ] Do all S3 stuff in go code. (with option for local mode testing)
- [ ] Secrets management.. including technique for creating envs + their secrets.
- [ ] Use localstack to test s3 and sqs.
- [ ] auto-publish new docker image on push to main. 
- [ ] Create project overview in README.md
- [ ] Create a CONTRIBUTING.md
- [ ] github repo: better default description
- [ ] github readme: link to docs
- [ ] github repo: set up protected branch + pull request process
- [ ] yaml files: aggressive validation (and run it in PR branches)
- [ ] Logging: switch to zap or similar
- [ ] Config: switch to viper or similar
- [ ] Yaml parsing: smarter handling of DataId field (don't store it as 2 fields)

## Pending DX tasks
- [ ] temporary files and archives: keep them in memory avoiding weird filesystem bugs/flaky tests.

## Tasks for later not required for MVP:
[ ] Create dorkly org in docker hub