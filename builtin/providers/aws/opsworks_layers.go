package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
)

// OpsWorks has a single concept of "layer" which represents several different
// layer types. The differences between these are in some extra properties that
// get packed into an "Attributes" map, but in the OpsWorks UI these are presented
// as first-class options, and so Terraform prefers to expose them this way and
// hide the implementation detail that they are all packed into a single type
// in the underlying API.
//
// This file contains utilities that are shared between all of the concrete
// layer resource types, which have names matching aws_opsworks_*_layer .

type opsworksLayerTypeAttribute struct {
	AttrName  string
	Type      schema.ValueType
	Default   interface{}
	Required  bool
	WriteOnly bool
}

type opsworksLayerType struct {
	TypeName         string
	DefaultLayerName string
	Attributes       map[string]*opsworksLayerTypeAttribute
	CustomShortName  bool
}

var (
	opsworksTrueString  = "1"
	opsworksFalseString = "0"
)

func (lt *opsworksLayerType) SchemaResource() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},

		"auto_assign_elastic_ips": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"auto_assign_public_ips": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"custom_instance_profile_arn": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},

		"custom_setup_recipes": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_configure_recipes": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_deploy_recipes": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_undeploy_recipes": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_shutdown_recipes": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_security_group_ids": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Set:      schema.HashString,
		},

		"auto_healing": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"install_updates_on_boot": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"instance_shutdown_timeout": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Default:  120,
		},

		"drain_elb_on_shutdown": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"system_packages": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Set:      schema.HashString,
		},

		"stack_id": &schema.Schema{
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},

		"use_ebs_optimized_instances": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		// TODO: ebs_volume block
		/*"ebs_volume": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"iops": &schema.Schema{
						Type:     schema.TypeInt,
						Optional: true,
					},

					"mount_point": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},

					"number_of_disks": &schema.Schema{
						Type:     schema.TypeInt,
						Required: true,
					},

					"raid_level": &schema.Schema{
						Type:     schema.TypeInt,
						Required: true,
					},

					"size": &schema.Schema{
						Type:     schema.TypeInt,
						Required: true,
					},

					"volume_type": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Default:  "standard",
					},
				},
			},
			Set: func(v interface{}) int {
				m := v.(map[string]interface{})
				return hashcode.String(m["mount_point"].(string))
			},
		},*/
	}

	if lt.CustomShortName {
		resourceSchema["short_name"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		}
	}

	if lt.DefaultLayerName != "" {
		resourceSchema["name"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  lt.DefaultLayerName,
		}
	} else {
		resourceSchema["name"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		}
	}

	for key, def := range lt.Attributes {
		resourceSchema[key] = &schema.Schema{
			Type:     def.Type,
			Default:  def.Default,
			Required: def.Required,
			Optional: !def.Required,
		}
	}

	return &schema.Resource{
		Read: func(nd *schema.ResourceData, meta interface{}) error {
			client := meta.(*AWSClient).opsworksconn
			d := makeAwsResourceData(nd)
			return lt.Read(d, client)
		},
		Create: func(nd *schema.ResourceData, meta interface{}) error {
			client := meta.(*AWSClient).opsworksconn
			d := makeAwsResourceData(nd)
			return lt.Create(d, client)
		},
		Update: func(nd *schema.ResourceData, meta interface{}) error {
			client := meta.(*AWSClient).opsworksconn
			d := makeAwsResourceData(nd)
			return lt.Update(d, client)
		},
		Delete: func(nd *schema.ResourceData, meta interface{}) error {
			client := meta.(*AWSClient).opsworksconn
			d := makeAwsResourceData(nd)
			return lt.Delete(d, client)
		},

		Schema: resourceSchema,
	}
}

func (lt *opsworksLayerType) Read(d *awsResourceData, client *opsworks.OpsWorks) error {

	req := &opsworks.DescribeLayersInput{
		LayerIDs: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks layer: %s", d.Id())

	resp, err := client.DescribeLayers(req)
	if err != nil {
		if awsErr, ok := err.(aws.APIError); ok && awsErr.Code == "ResourceNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	layer := resp.Layers[0]
	d.Set("id", layer.LayerID)
	d.Set("auto_assign_elastic_ips", layer.AutoAssignElasticIPs)
	d.Set("auto_assign_public_ips", layer.AutoAssignPublicIPs)
	d.Set("custom_instance_profile_arn", layer.CustomInstanceProfileARN)
	d.setAwsStringList("custom_security_group_ids", layer.CustomSecurityGroupIDs)
	d.Set("auto_healing", layer.EnableAutoHealing)
	d.Set("install_updates_on_boot", layer.InstallUpdatesOnBoot)
	d.Set("name", layer.Name)
	d.setAwsStringList("system_packages", layer.Packages)
	d.Set("stack_id", layer.StackID)
	d.Set("use_ebs_optimized_instances", layer.UseEBSOptimizedInstances)
	// TODO: d.Set("ebs_volume", ...)

	if lt.CustomShortName {
		d.Set("short_name", layer.Shortname)
	}

	lt.SetAttributeMap(d, layer.Attributes)
	lt.SetLifecycleEventConfiguration(d, layer.LifecycleEventConfiguration)
	lt.SetCustomRecipes(d, layer.CustomRecipes)

	return nil
}

func (lt *opsworksLayerType) Create(d *awsResourceData, client *opsworks.OpsWorks) error {

	req := &opsworks.CreateLayerInput{
		AutoAssignElasticIPs:        d.getAwsBool("auto_assign_elastic_ips"),
		AutoAssignPublicIPs:         d.getAwsBool("auto_assign_public_ips"),
		CustomInstanceProfileARN:    d.getAwsString("custom_instance_profile_arn"),
		CustomRecipes:               lt.CustomRecipes(d),
		CustomSecurityGroupIDs:      d.getAwsStringSet("custom_security_group_ids"),
		EnableAutoHealing:           d.getAwsBool("auto_healing"),
		InstallUpdatesOnBoot:        d.getAwsBool("install_updates_on_boot"),
		LifecycleEventConfiguration: lt.LifecycleEventConfiguration(d),
		Name:                     d.getAwsString("name"),
		Packages:                 d.getAwsStringSet("system_packages"),
		Type:                     aws.String(lt.TypeName),
		StackID:                  d.getAwsString("stack_id"),
		UseEBSOptimizedInstances: d.getAwsBool("use_ebs_optimized_instances"),
		Attributes:               lt.AttributeMap(d),
		// TODO: VolumeConfigurations: ...,
	}

	if lt.CustomShortName {
		req.Shortname = d.getAwsString("short_name")
	} else {
		req.Shortname = aws.String(lt.TypeName)
	}

	log.Printf("[DEBUG] Creating OpsWorks layer: %s", d.Id())

	resp, err := client.CreateLayer(req)
	if err != nil {
		return err
	}

	layerId := *resp.LayerID
	d.SetId(layerId)
	d.Set("id", layerId)

	return lt.Read(d, client)
}

func (lt *opsworksLayerType) Update(d *awsResourceData, client *opsworks.OpsWorks) error {

	req := &opsworks.UpdateLayerInput{
		LayerID:                     aws.String(d.Id()),
		AutoAssignElasticIPs:        d.getAwsBool("auto_assign_elastic_ips"),
		AutoAssignPublicIPs:         d.getAwsBool("auto_assign_public_ips"),
		CustomInstanceProfileARN:    d.getAwsString("custom_instance_profile_arn"),
		CustomRecipes:               lt.CustomRecipes(d),
		CustomSecurityGroupIDs:      d.getAwsStringSet("custom_security_group_ids"),
		EnableAutoHealing:           d.getAwsBool("auto_healing"),
		InstallUpdatesOnBoot:        d.getAwsBool("install_updates_on_boot"),
		LifecycleEventConfiguration: lt.LifecycleEventConfiguration(d),
		Name:                     d.getAwsString("name"),
		Packages:                 d.getAwsStringSet("system_packages"),
		UseEBSOptimizedInstances: d.getAwsBool("use_ebs_optimized_instances"),
		Attributes:               lt.AttributeMap(d),
		// TODO: VolumeConfigurations: ...,
	}

	if lt.CustomShortName {
		req.Shortname = d.getAwsString("short_name")
	} else {
		req.Shortname = aws.String(lt.TypeName)
	}

	log.Printf("[DEBUG] Updating OpsWorks layer: %s", d.Id())

	_, err := client.UpdateLayer(req)
	if err != nil {
		return err
	}

	return lt.Read(d, client)
}

func (lt *opsworksLayerType) Delete(d *awsResourceData, client *opsworks.OpsWorks) error {
	req := &opsworks.DeleteLayerInput{
		LayerID: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks layer: %s", d.Id())

	_, err := client.DeleteLayer(req)
	return err
}

func (lt *opsworksLayerType) AttributeMap(d *awsResourceData) *map[string]*string {
	attrs := map[string]*string{}

	for key, def := range lt.Attributes {
		value := d.Get(key)
		switch def.Type {
		case schema.TypeString:
			strValue := value.(string)
			attrs[def.AttrName] = &strValue
		case schema.TypeInt:
			intValue := value.(int)
			strValue := strconv.Itoa(intValue)
			attrs[def.AttrName] = &strValue
		case schema.TypeBool:
			boolValue := value.(bool)
			if boolValue {
				attrs[def.AttrName] = &opsworksTrueString
			} else {
				attrs[def.AttrName] = &opsworksFalseString
			}
		default:
			// should never happen
			panic(fmt.Errorf("Unsupported OpsWorks layer attribute type"))
		}
	}

	return &attrs
}

func (lt *opsworksLayerType) SetAttributeMap(d *awsResourceData, attrsPtr *map[string]*string) {
	var attrs map[string]*string
	if attrsPtr != nil {
		attrs = *attrsPtr
	} else {
		attrs = map[string]*string{}
	}

	for key, def := range lt.Attributes {
		// Ignore write-only attributes; we'll just keep what we already have stored.
		// (The AWS API returns garbage placeholder values for these.)
		if def.WriteOnly {
			continue
		}

		if strPtr, ok := attrs[def.AttrName]; ok && strPtr != nil {
			strValue := *strPtr

			switch def.Type {
			case schema.TypeString:
				d.Set(key, strValue)
			case schema.TypeInt:
				intValue, err := strconv.Atoi(strValue)
				if err == nil {
					d.Set(key, intValue)
				} else {
					// Got garbage from the AWS API
					d.Set(key, nil)
				}
			case schema.TypeBool:
				boolValue := true
				if strValue == opsworksFalseString {
					boolValue = false
				}
				d.Set(key, boolValue)
			default:
				// should never happen
				panic(fmt.Errorf("Unsupported OpsWorks layer attribute type"))
			}
			return

		} else {
			d.Set(key, nil)
		}
	}
}

func (lt *opsworksLayerType) LifecycleEventConfiguration(d *awsResourceData) *opsworks.LifecycleEventConfiguration {
	return &opsworks.LifecycleEventConfiguration{
		Shutdown: &opsworks.ShutdownEventConfiguration{
			DelayUntilELBConnectionsDrained: d.getAwsBool("drain_elb_on_shutdown"),
			ExecutionTimeout:                d.getAwsLong("instance_shutdown_timeout"),
		},
	}
}

func (lt *opsworksLayerType) SetLifecycleEventConfiguration(d *awsResourceData, v *opsworks.LifecycleEventConfiguration) {
	if v == nil || v.Shutdown == nil {
		d.Set("drain_elb_on_shutdown", nil)
		d.Set("instance_shutdown_timeout", nil)
	} else {
		d.Set("drain_elb_on_shutdown", v.Shutdown.DelayUntilELBConnectionsDrained)
		d.Set("instance_shutdown_timeout", v.Shutdown.ExecutionTimeout)
	}
}

func (lt *opsworksLayerType) CustomRecipes(d *awsResourceData) *opsworks.Recipes {
	return &opsworks.Recipes{
		Configure: d.getAwsStringList("custom_configure_recipes"),
		Deploy:    d.getAwsStringList("custom_deploy_recipes"),
		Setup:     d.getAwsStringList("custom_setup_recipes"),
		Shutdown:  d.getAwsStringList("custom_shutdown_recipes"),
		Undeploy:  d.getAwsStringList("custom_undeploy_recipes"),
	}
}

func (lt *opsworksLayerType) SetCustomRecipes(d *awsResourceData, v *opsworks.Recipes) {
	// Null out everything first, and then we'll consider what to put back.
	d.Set("custom_configure_recipes", nil)
	d.Set("custom_deploy_recipes", nil)
	d.Set("custom_setup_recipes", nil)
	d.Set("custom_shutdown_recipes", nil)
	d.Set("custom_undeploy_recipes", nil)

	if v == nil {
		return
	}

	d.setAwsStringList("custom_configure_recipes", v.Configure)
	d.setAwsStringList("custom_deploy_recipes", v.Deploy)
	d.setAwsStringList("custom_setup_recipes", v.Setup)
	d.setAwsStringList("custom_shutdown_recipes", v.Shutdown)
	d.setAwsStringList("custom_undeploy_recipes", v.Undeploy)
}
