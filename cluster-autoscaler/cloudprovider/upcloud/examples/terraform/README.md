# Cluster Autoscaler for UpCloud

## Deploy autoscaler using Terraform

### Required environment variables
- `UPCLOUD_USERNAME` - UpCloud's API username
- `UPCLOUD_PASSWORD` - UpCloud's API user's password

### Apply Terraform plan
Init Terraform if needed
```shell
$ terraform init
```

This example uses `autoscaler_username` and `autoscaler_password` input variables to set cluster autoscaler credentials.  
For demonstration purposes, we can use same account that we use with Terraform provider:
```shell
$ TF_VAR_autoscaler_username=$UPCLOUD_USERNAME TF_VAR_autoscaler_password=$UPCLOUD_PASSWORD terraform apply
```