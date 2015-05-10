package aws

import (
	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
)

type awsResourceData struct {
	*schema.ResourceData
}

func makeAwsResourceData(d *schema.ResourceData) *awsResourceData {
	return &awsResourceData{
		ResourceData: d,
	}
}

// Convert a string value from ResourceData to an AWS string, nor nil if not set.
func (d *awsResourceData) getAwsString(key string) *string {
	if val, ok := d.GetOk(key); ok {
		return aws.String(val.(string))
	} else {
		return nil
	}
}

// Convert a string value from ResourceData to an AWS string, nor nil if not set.
func (d *awsResourceData) getAwsBool(key string) *bool {
	if val, ok := d.GetOk(key); ok {
		return aws.Boolean(val.(bool))
	} else {
		return nil
	}
}
