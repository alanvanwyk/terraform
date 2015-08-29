---
layout: "chef"
page_title: "Chef: chef_client"
sidebar_current: "docs-chef-resource-client"
description: |-
  Creates and manages an API client in Chef Server.
---

# chef\_client

A *client* object grants Chef Server API access to a particular API client,
establishing its name and issuing it a private key it can use to authenticate.

In the default configuration of Chef server API clients cannot create other
API clients, so this resource requires Terraform itself to authenticate with
Client credentials that grant it adminstrative access.

~> **Important Security Notice** When using this resource, the private key
generated for the client will be stored *unencrypted* in your Terraform state
file. Anyone with access to your state file will have the ability to make API
requests on behalf of the created client. For production use it is better to
create client credentials outside of Terraform to avoid storing sensitive
information in the state file.

## Example Usage

```
resource "chef_client" "example" {
    name = "example-client"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The unique name to assign to the API client. This will
  be the client's identifier when making requests.

* `admin` - (Optional) Boolean that, when set to ``true``, causes the created
  client to have administrative access. Defaults to ``false``.

## Attributes Reference

The following attributes are exported:

* `private_key_pem` - PEM-formatted private key that the client must use to
  authenticate.
* `public_key_pem` - PEM-formatted public key corresponding to the assigned
  public key.

