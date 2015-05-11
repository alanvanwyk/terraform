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

// Convert a string value from ResourceData to an AWS string, or nil if not set.
func (d *awsResourceData) getAwsString(key string) *string {
	if val, ok := d.GetOk(key); ok {
		return aws.String(val.(string))
	} else {
		return nil
	}
}

// Convert a bool value from ResourceData to an AWS boolean, or nil if not set.
func (d *awsResourceData) getAwsBool(key string) *bool {
	if val, ok := d.GetOk(key); ok {
		return aws.Boolean(val.(bool))
	} else {
		return nil
	}
}

// Convert an int value from ResourceData to an AWS long, or nil if not set.
func (d *awsResourceData) getAwsLong(key string) *int64 {
	if val, ok := d.GetOk(key); ok {
		return aws.Long(int64(val.(int)))
	} else {
		return nil
	}
}

// Convert a string list value from ResourceData to an AWS string list, empty if not set.
func (d *awsResourceData) getAwsStringList(key string) []*string {
	if v, ok := d.GetOk(key); ok && v != nil {
		in := v.([]interface {})
		ret := make([]*string, len(in), len(in))
		for i := 0; i < len(in); i++ {
			ret[i] = aws.String(in[i].(string))
		}
		return ret
	} else {
		return make([]*string, 0, 0)
	}
}

// Convert a string set value from ResourceData to an AWS string list, empty if not set.
func (d *awsResourceData) getAwsStringSet(key string) []*string {
	if v, ok := d.GetOk(key); ok && v != nil {
		set := v.(*schema.Set)
		in := set.List()
		ret := make([]*string, len(in), len(in))
		for i := 0; i < len(in); i++ {
			ret[i] = aws.String(in[i].(string))
		}
		return ret
	} else {
		return make([]*string, 0, 0)
	}
}

// Write an AWS string list into a string list value in the ResourceData, empty if not set
func (d *awsResourceData) setAwsStringList(key string, v []*string) {
	nv := make([]string, len(v), len(v))
	for i := 0; i < len(v); i++ {
		if v[i] != nil {
			nv[i] = *(v[i])
		}
	}
	d.Set(key, nv)
}
