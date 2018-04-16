Deploy a Production-Ready Docker Swarm Cluster on AWS with Alexa.

# How it works

Schema

# Lambda Functions

## Infrastructure Lambda Function

Description

### IAM Role

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "1",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole",
                "dynamodb:PutItem",
                "sqs:SendMessage",
                "sqs:SetQueueAttributes"
            ],
            "Resource": [
                "arn:aws:sqs:AWS_REGION:ACCOUNT_ID:QUEUE_NAME",
                "arn:aws:iam::ACCOUNT_ID:role/SSM_ROLE_NAME",
                "arn:aws:dynamodb:AWS_REGION:ACCOUNT_ID:table/TABLE_NAME"
            ]
        },
        {
            "Sid": "2",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "ec2:CreateTags",
                "ec2:RunInstances",
                "logs:CreateLogGroup",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

### Environment Variables

| Name | Description |
| ---- | ----------- |
| AMI  | Amazon Machine Image ID with Docker CE pre-installed |
| KEYPAIR | AWS SSH KeyPair |
| SSM_ROLE_NAME | IAM Role with SSM permissions for EC2 instances |
| SECURITY_GROUP | Security Group ID with allowed inbound traffic on 2377/tcp |
| SQS_URL | SQS URL |
| TABLE_NAME | DynamoDB Table name |

## Swarm Lambda Function

Description

### IAM Role

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "1",
            "Effect": "Allow",
            "Action": [
                "sqs:DeleteMessage",
                "sqs:ReceiveMessage",
                "dynamodb:UpdateItem"
            ],
            "Resource": [
                "arn:aws:sqs:AWS_REGION:ACCOUNT_ID:QUEUE_NAME",
                "arn:aws:dynamodb:AWS_REGION:ACCOUNT_ID:table/TABLE_NAME"
            ]
        },
        {
            "Sid": "2",
            "Effect": "Allow",
            "Action": [
                "ssm:SendCommand",
                "logs:CreateLogStream",
                "logs:CreateLogGroup",
                "logs:PutLogEvents",
                "ssm:GetCommandInvocation"
            ],
            "Resource": "*"
        }
    ]
}
```

### Environment Variables

| Name | Description |
| ---- | ----------- |
| SQS_URL | SQS URL |
| TABLE_NAME | DynamoDB Table name |

# Going further

* Cleanup a Swarm Cluster
* Deploy Docker Containers
* Deploy Swarm in speicifc VPC
* Change Instance Type
* Deploy with multiple managers
* etc

# Licence

MIT

# Maintainers

* Mohamed Labouardy <mohamed@labouardy.com>

# Tutorial
