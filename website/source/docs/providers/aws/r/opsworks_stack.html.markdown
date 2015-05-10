---
layout: "aws"
page_title: "AWS: aws_opsworks_stack"
sidebar_current: "docs-aws-resource-opsworks-stack"
description: |-
  Provides an OpsWorks stack resource.
---

# aws\_opsworks\_stack

Provides an OpsWorks stack resource.

## Example Usage

```
resource "aws_opsworks_stack" "main" {
    name = "awesome-stack"
    region = "us-west-1"
    service_role_arn = "{$aws_iam_role.opsworks.arn}"
    default_instance_profile_arn = "${aws_iam_instance_profile.opsworks.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the stack.
* `region` - (Required) The name of the region where the stack will exist.
* `service_role_arn` - (Required) The ARN of an IAM role that the OpsWorks service will act as.
* `default_instance_profile_arn` - (Required) The ARN of an IAM Instance Profile that created instances
  will have by default.
* `color` - (Optional) Color to paint next to the stack's resources in the OpsWorks console.
* `default_availability_zone` - (Optional) Name of the availability zone where instances will be created
  by default. This is required unless you set `vpc_id`.
* `chef_configuration` - (Optional) Extra configuration for Chef. This is a sub-object as described below.
* `configuration_manager` - (Optional) Selects which configuration management system is used. This is a
  sub-object as described below.
* `custom_cookbooks_source` - (Optional) When `use_custom_cookbooks` is set, provide this sub-object as
  described below.
* `default_os` - (Optional) Name of OS that will be installed on instances by default.
* `default_root_device_type` - (Optional) Name of the type of root device instances will have by default.
* `default_ssh_key_name` - (Optional) Name of the SSH keypair that instances will have by default.
* `default_subnet_id` - (Optional) Id of the subnet in which instances will be created by default. Mandatory
  if `vpc_id` is set, and forbidden if it isn't.
* `hostname_theme` - (Optional) Keyword representing the naming scheme that will be used for instance hostnames
  within this stack.
* `use_custom_cookbooks` - (Optional) Boolean value controlling whether the custom cookbook settings are
  enabled.
* `use_opsworks_security_groups` - (Optional) Boolean value controlling whether the standard OpsWorks
  security groups apply to created instances.
* `vpc_id` - (Optional) The id of the VPC that this stack belongs to.

The `configuration_manager` block supports the following arguments:

* `name` - (Required) The name of the configuration manager to use. For example, "Chef".
* `version` - (Required) The version number of the given configuration manager.

The `chef_configuration` block supports the following arguments:

* `manage_berkshelf` - (Optional) Boolean value controlling whether OpsWorks will manage Berkshelf.
* `berkshelf_version` - (Optional) The version of Berkshelf to use.

The `custom_cookbooks_source` block supports the following arguments:

* `type` - (Required) The type of source to use. For example, "archive".
* `url` - (Required) The URL where the cookbooks resource can be found.
* `username` - (Optional) Username to use when authenticating to the source.
* `password` - (Optional) Password to use when authenticating to the source.
* `ssh_key` - (Optional) SSH key to use when authenticating to the source.
* `revision` - (Optional) For sources that are version-aware, the revision to use.

## Attributes Reference

The following attributes are exported:

* `id` - The id of the stack.
