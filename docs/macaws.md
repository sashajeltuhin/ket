# Using a Mac & AWS Free Tier

## One time environment set up

1. Make a new [AWS account](https://aws.amazon.com/free/) (you may re-use an existing one, but it's more likely you will run in to IAM issues)
2. You will need an `Access Key Id` and `Secret Access Key`. 
   * You can either [generate one for your root user](https://console.aws.amazon.com/iam/home?region=us-east-1#/security_credential) or build a user in IAM with full access to VPCs and EC2
3. Open a Terminal session
4. <table><tr><td>`mkdir ~/kismatic` <br/>
   `cd ~/kismatic`</td> 
   <td>*Make a new directory for Kismatic (~/kismatic would work) and make it the working directory*</td></tr></table>
5. <table><tr><td>`wget -O - https://kismatic-installer.s3-accelerate.amazonaws.com/latest-darwin/kismatic.tar.gz | tar -zx`</td> 
   <td> *Download & unpack Kismatic*</td></tr></table>
6. <table><tr><td>`export AWS_ACCESS_KEY_ID=YOURACCESSKEYID`<br/>
`export AWS_SECRET_ACCESS_KEY=YOURSECRETACCESSKEY`</td><td> *Export your AWS credentials for use by the provision tool*</td></tr></table>

## Make a new cluster

7. <table><tr><td>`./provision aws create -i="micro" -f`</td><td> *create 3 t2.micro instances in AWS inside a Kismatic VPC and create a Kismatic keypair.*</td></tr></table>
8. <table><tr><td>`./kismatic install apply -f kismatic-cluster.yaml`</td><td> *install Kubernetes*</td></tr></table>

## Tear it down when you're done with it

9. <table><tr><td>`./provision aws delete-all`</td><td> *remove any EC2 instances tagged as created by kismatic on this machine.*</td></tr></table>