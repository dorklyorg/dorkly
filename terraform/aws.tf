locals {
  aws_region = "us-west-2"
  flagArchive = "flags.tar.gz"
}

data "aws_caller_identity" "current" {}

provider "aws" {
  region = local.aws_region
  default_tags {
    tags = {
      terraform = "true"
      dorkly    = "true"
    }
  }
}

resource "aws_lightsail_container_service" "dorkly" {
  name        = "dorkly"
  power       = "nano"
  scale       = 1
  is_disabled = false

}

resource "aws_lightsail_container_service_deployment_version" "dorkly" {
  container {
    container_name = "dorkly"
    image          = "drichelson/dorkly:latest"

    command = []

    environment = {
      AWS_REGION            = local.aws_region
      SQS_QUEUE_URL         = aws_sqs_queue.dorkly_queue.url
      S3_URL                = "s3://${aws_s3_bucket.dorkly_bucket.bucket}/flags.tar.gz"

      # TODO: can we use role permissions instead of access keys?
      AWS_ACCESS_KEY_ID     = aws_iam_access_key.dorkly_read_user_access_key.id
      AWS_SECRET_ACCESS_KEY = aws_iam_access_key.dorkly_read_user_access_key.secret
    }

    ports = {
      8030 = "HTTP"
    }
  }

  public_endpoint {
    container_name = "dorkly"
    container_port = 8030

    health_check {
      healthy_threshold   = 2
      unhealthy_threshold = 2
      timeout_seconds     = 3
      interval_seconds    = 10
      path                = "/status"
      success_codes       = "200"
    }
  }

  service_name = aws_lightsail_container_service.dorkly.name
}

# SQS Queue
resource "aws_sqs_queue" "dorkly_queue" {
  name   = "dorkly"
  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = [
          "SQS:SendMessage"
        ]
        Resource = [
          "arn:aws:sqs:*:*:dorkly"
        ],
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_s3_bucket.dorkly_bucket.arn
          }
        }
      }
    ]
  })
}

# S3 Bucket
resource "aws_s3_bucket" "dorkly_bucket" {
  bucket = "dorkly"
}

data "aws_s3_object" "existing_flag_archive" {
  bucket = aws_s3_bucket.dorkly_bucket.bucket
  key    = local.flagArchive
}

# We only want to upload the flags archive if it doesn't already exist.
resource "aws_s3_object" "maybe_upload_flag_archive" {
  count  = data.aws_s3_object.existing_flag_archive.body == null ? 1 : 0
  bucket = aws_s3_bucket.dorkly_bucket.bucket
  key    = local.flagArchive
  source = "flags.tar.gz"
  acl    = "private"
}

# S3 Bucket Notification
resource "aws_s3_bucket_notification" "dorkly_bucket_notification" {
  bucket = aws_s3_bucket.dorkly_bucket.id

  queue {
    queue_arn = aws_sqs_queue.dorkly_queue.arn
    events    = ["s3:ObjectCreated:*"]
  }

  depends_on = [
    aws_sqs_queue.dorkly_queue
  ]
}

# IAM User for reading SQS queue and reading S3 bucket
resource "aws_iam_user" "dorkly_read_user" {
  name = "dorkly-read"
}

resource "aws_iam_access_key" "dorkly_read_user_access_key" {
  user = aws_iam_user.dorkly_read_user.name
}

resource "aws_iam_user_policy" "dorkly_read_user_policy" {
  name = "dorkly-read-policy"
  user = aws_iam_user.dorkly_read_user.name

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          "${aws_s3_bucket.dorkly_bucket.arn}/*",
          aws_sqs_queue.dorkly_queue.arn
        ]
      }
    ]
  })
}

# IAM User for writing S3 bucket
resource "aws_iam_user" "dorkly_write_user" {
  name = "dorkly-write"
}

resource "aws_iam_access_key" "dorkly_write_user_access_key" {
  user = aws_iam_user.dorkly_write_user.name
}

resource "aws_iam_user_policy" "dorkly_write_user_policy" {
  name = "dorkly-write-user-policy"
  user = aws_iam_user.dorkly_write_user.name

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject"
        ]
        Resource = [
          "${aws_s3_bucket.dorkly_bucket.arn}/*"
        ]
      }
    ]
  })
}