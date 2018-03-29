/*
- ami env variable
- keypair env variable

conversation:
	welcome to swarm: orchestration bla bla
		- swarm status
			how many stack
			how many service per stack
			how many nodes
		- deploy new swarm
			create ec2 instances
			convert to swarm
*/

## Lambda Functions:

### Setup Infrasturcture

* Setup an AMI with preinstalled Docker CE
* Setup a Security group with SSH & allow inbound traffic on port 2377
* Setup an IAM role with SSM permissions
* Setup a SSH KeyPair
* Setup an SQS Queue
* Setup DynamoDB Table

```
export AMI="ami-38f42c45"
export KEYPAIR="vpc"
export SSM_ROLE_NAME="SSMRole"
export SECURITY_GROUP_ID="sg-72caf704"
export SQS_URL="https://sqs.us-east-1.amazonaws.com/ID/SwarmQueue"
export TABLE_NAME="clusters"
```

```
func encodeUserData() string {
	userData := `#/bin/sh
	yum update -y
	yum install -y docker
	service docker start
	usermod -aG docker ec2-user
	`
	return b64.StdEncoding.EncodeToString([]byte(userData))
}
```

### Setup Swarm Cluster


```
export SQS_URL="https://sqs.us-east-1.amazonaws.com/305929695733/SwarmQueue"
```


- 3 nodes
- cluster name OK
- Lambda 1: create infra OK
- Lambda 1: push to sqs OK
- Lambda 1: push to dynamodb: name, how many nodes, status OK
- Lambda 2: get sqs OK
- Lambda 2: swarm setup Ok
- Lambda 1: update dynamodb with new status OK
- Lambda 3: Swarm status OK
- Lambda 3: get from dynamodb status OK
- Lambda 3: get how many stacks on swarm manager


setup a dynamodb table
insert to dynamodb
get from dynamodb and update


dynamodb table:
	id
	name
	nodes
	status
	manager_ip

## Going further

* Cleanup a Swarm Cluster
* Deploy Docker Containers
* Deploy Swarm in speicifc VPC
* Change Instance Type
* Deploy with multiple managers
* etc


## Schenario

* Alexa, open docker swarm
* Hello Mohamed, welcome to Containers universe, heres is what you can do
	- deploy a swarm cluster A W S
	- get state of your cluster
	- know how many services are deploying in your cluster
* Deploy a swarm cluster
* sure, how many nodes do you want ?
* 3
* what do you want to name it ?
* staging
* sure, you cluster is been deploying, it will be created in virgin region
* tell me the state of my cluster 


$ /c/Users/Mohamed/go/bin/build-lambda-zip.exe -o main.zip main
$ GOOS=linux GOARCH=amd64 go build -o main
