provider "github" {
  #   token = "" Alternately specified using GITHUB_TOKEN env var
  owner = "dorklyorg"
}

resource "github_repository" "dorkly_repo" {
  name        = "dorkly-flags-test-1"
  description = "feature flags for dorkly"
  visibility  = "private"
  auto_init   = true
}

resource "github_actions_variable" "aws_region" {
  repository    = github_repository.dorkly_repo.name
  variable_name = "AWS_REGION"
  value         = local.aws_region
}

resource "github_actions_secret" "aws_access_key_id" {
  repository      = github_repository.dorkly_repo.name
  secret_name     = "AWS_ACCESS_KEY_ID"
  plaintext_value = aws_iam_access_key.dorkly_write_user_access_key.id
}

resource "github_actions_secret" "aws_secret_key_secret" {
  repository      = github_repository.dorkly_repo.name
  secret_name     = "AWS_SECRET_ACCESS_KEY"
  plaintext_value = aws_iam_access_key.dorkly_write_user_access_key.secret
}

# Upload each file to the repo using the same relative path minus 'files/'
resource "github_repository_file" "dorkly_flags_project" {
  for_each = toset(
    [
      "files/project/project.yml",
      "files/project/flags/boolean1.yml",
      "files/project/flags/rollout1.yml",
      "files/project/environments/production/boolean1.yml",
      "files/project/environments/staging/boolean1.yml",
      "files/project/environments/production/rollout1.yml",
      "files/project/environments/staging/rollout1.yml",
    ])
  repository          = github_repository.dorkly_repo.name
  branch              = "main"
  file                = trimprefix(each.key, "files/")
  content             = file(each.key)
  commit_message      = "Managed by Terraform"
  overwrite_on_create = true
}


