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

resource "github_repository_file" "dorkly_workflows" {
  repository          = github_repository.dorkly_repo.name
  branch              = "main"
  file                = ".github/workflows/dorkly.yml"
  content             = file("files/dorkly.yml")
  commit_message      = "Managed by Terraform"
  overwrite_on_create = true
}

resource "github_repository_file" "dorkly_example_flags" {
  repository          = github_repository.dorkly_repo.name
  branch              = "main"
  file                = "flags/test-data.json"
  content             = file("files/flags/test-data.json")
  commit_message      = "Managed by Terraform"
  overwrite_on_create = true
}

resource "github_repository_file" "dorkly_example_env" {
  repository          = github_repository.dorkly_repo.name
  branch              = "main"
  file                = "flags/test.json"
  content             = file("files/flags/test.json")
  commit_message      = "Managed by Terraform"
  overwrite_on_create = true
}

