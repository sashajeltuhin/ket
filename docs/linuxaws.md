# Using Linux & AWS Free Tier
## One time environment set up

1. Make a new [AWS account](https://aws.amazon.com/free/) (you may re-use an existing one, but it's more likely you will run in to IAM issues)
2. You will need an `Access Key Id` and `Secret Access Key`. You can either generate one for your root user (https://console.aws.amazon.com/iam/home?region=us-east-1#/security_credential, click "Access Keys") or build a more secure user in IAM with full access to VPCs and EC2
3. `mkdir /opt/kismatic` Make a new directory for Kismatic
   `cd /opt/kismatic` And make it the working directory
4. `curl -L https://kismatic-installer.s3-accelerate.amazonaws.com/latest/kismatic.tar.gz | tar -zx` Download & unpack Kismatic
5. `export AWS_ACCESS_KEY_ID=<from #2>` Export your AWS credentials for use by the provision tool
   `export AWS_SECRET_ACCESS_KEY=<from #2>`

## Make a new cluster

6. `./provision aws create -i="micro" -f` This will create 3 t2.micro instances in AWS inside a Kismatic VPC and create a Kismatic keypair.
7. `./kismatic install apply -f kismatic-cluster.yaml` This will install Kubernetes

## Tear it down when you're done with it

8. <table><tr><td>`./provision aws delete-all`</td><td> *remove any EC2 instances tagged as created by kismatic on this machine.*</td></tr></table>