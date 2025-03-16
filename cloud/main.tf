terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_ecr_repository" "heartbeat_ecr_repo" {
  name = "heartbeat"
}

# TODO: ECR for container
resource "aws_lambda_function" "heartbeat_lambda" {
    function_name = "heartbeat"
    role = aws_iam_role.iam_for_lambda.arn
    image_uri = "${aws_ecr_repository.heartbeat_ecr_repo.repository_url}:latest"
    package_type = "Image"
  }

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "publish_to_sns" {
  statement {
    effect = "Allow"

    # principals {
      # type        = "Service"
      # identifiers = ["sns.amazonaws.com"]
    # }
#
    resources = [aws_sns_topic.user_updates.arn]
    actions = ["sns:Publish"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "iam-for-lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "add_cloudwatch_to_lambda" {
  role = aws_iam_role.iam_for_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}


resource "aws_iam_policy" "sns_publish_policy" {
  name = "lambda-publish"
  policy = data.aws_iam_policy_document.publish_to_sns.json
}

resource "aws_iam_role_policy_attachment" "add_sns_to_lambda" {
  role = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.sns_publish_policy.arn
}

resource "aws_cloudwatch_event_rule" "heartbeat_schedule" {
    name = "heartbeat-schedule"
    description = "trigger regular heartbeats for webdrones server"
    schedule_expression = "rate(1 minute)"
}

resource "aws_lambda_permission" "allow_heartbeat_schedule_to_call_lambda" {
    statement_id = "event-bridge-lambda-permission"
    action = "lambda:InvokeFunction"
    function_name = aws_lambda_function.heartbeat_lambda.function_name
    principal = "events.amazonaws.com"
    source_arn = aws_cloudwatch_event_rule.heartbeat_schedule.arn
}

resource "aws_sns_topic" "user_updates" {
  name = "heartbeat-failed"
}

resource "aws_sns_topic_subscription" "email_heartbeat_subscription" {
  protocol = "email"
  endpoint = "ryan.michael.tech@gmail.com"
  topic_arn = aws_sns_topic.user_updates.arn

}


resource "aws_cloudwatch_event_target" "heartbeat_lambda_target" {
    arn = aws_lambda_function.heartbeat_lambda.arn
    rule = aws_cloudwatch_event_rule.heartbeat_schedule.id
    input = jsonencode(
        {"snsArn"=aws_sns_topic.user_updates.arn}
    ) 
}

