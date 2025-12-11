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

resource "aws_iam_role" "guard_security_scan_role" {
  name = "GuardSecurityScanRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${var.platform_account_id}:root"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "sts:ExternalId" = var.external_id
          }
        }
      }
    ]
  })

  tags = {
    Name        = "GuardSecurityScanRole"
    Purpose     = "CloudSentinel Security Scanning"
    ManagedBy   = "CloudSentinel"
  }
}

resource "aws_iam_role_policy_attachment" "readonly_access" {
  role       = aws_iam_role.guard_security_scan_role.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}