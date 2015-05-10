package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
)

func resourceAwsOpsworksStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsworksStackCreate,
		Read:   resourceAwsOpsworksStackRead,
		Update: resourceAwsOpsworksStackUpdate,
		Delete: resourceAwsOpsworksStackDelete,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"region": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"service_role_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"default_instance_profile_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"color": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"chef_configuration": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"berkshelf_version": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "3.2.0",
						},

						"manage_berkshelf": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"configuration_manager": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"version": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"custom_cookbooks_source": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"username": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"password": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"revision": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"ssh_key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"default_availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_os": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Ubuntu 12.04 LTS",
			},

			"default_root_device_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "instance-store",
			},

			"default_ssh_key_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"default_subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"hostname_theme": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Layer_Dependent",
			},

			"use_custom_cookbooks": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"use_opsworks_security_groups": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAwsOpsworksStackValidate(d *awsResourceData) error {
	chefConfigCount := d.Get("chef_configuration.#").(int)
	if chefConfigCount > 1 {
		return fmt.Errorf("Only one chef_configuration is permitted")
	}

	configManagerCount := d.Get("configuration_manager.#").(int)
	if configManagerCount > 1 {
		return fmt.Errorf("Only one configuration_manager is permitted")
	}

	cookbooksSourceCount := d.Get("custom_cookbooks_source.#").(int)
	if cookbooksSourceCount > 1 {
		return fmt.Errorf("Only one custom_cookbooks_source is permitted")
	}

	return nil
}

func (d *awsResourceData) opsworksStackConfigurationManager() *opsworks.StackConfigurationManager {
	count := d.Get("configuration_manager.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.StackConfigurationManager{
		Name:    d.getAwsString("configuration_manager.0.name"),
		Version: d.getAwsString("configuration_manager.0.version"),
	}
}

func (d *awsResourceData) opsworksSetStackConfigurationManager(v *opsworks.StackConfigurationManager) {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.Name != nil {
			m["name"] = *v.Name
		}
		if v.Version != nil {
			m["version"] = *v.Version
		}
		nv = append(nv, m)
	}

	err := d.Set("configuration_manager", nv)
	if err != nil {
		// should never happen
		panic(err)
	}
}

func (d *awsResourceData) opsworksStackChefConfiguration() *opsworks.ChefConfiguration {
	count := d.Get("chef_configuration.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.ChefConfiguration{
		BerkshelfVersion: d.getAwsString("chef_configuration.0.berkshelf_version"),
		ManageBerkshelf:  d.getAwsBool("chef_configuration.0.manage_berkshelf"),
	}
}

func (d *awsResourceData) opsworksSetStackChefConfiguration(v *opsworks.ChefConfiguration) {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.BerkshelfVersion != nil {
			m["berkshelf_version"] = *v.BerkshelfVersion
		}
		if v.ManageBerkshelf != nil {
			m["manage_berkshelf"] = *v.ManageBerkshelf
		}
		nv = append(nv, m)
	}

	err := d.Set("chef_configuration", nv)
	if err != nil {
		// should never happen
		panic(err)
	}
}

func (d *awsResourceData) opsworksStackCustomCookbooksSource() *opsworks.Source {
	count := d.Get("custom_cookbooks_source.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.Source{
		Type:     d.getAwsString("custom_cookbooks_source.0.type"),
		URL:      d.getAwsString("custom_cookbooks_source.0.url"),
		Username: d.getAwsString("custom_cookbooks_source.0.username"),
		Password: d.getAwsString("custom_cookbooks_source.0.password"),
		Revision: d.getAwsString("custom_cookbooks_source.0.revision"),
		SSHKey:   d.getAwsString("custom_cookbooks_source.0.ssh_key"),
	}
}

func (d *awsResourceData) opsworksSetStackCustomCookbooksSource(v *opsworks.Source) {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.Type != nil {
			m["type"] = *v.Type
		}
		if v.URL != nil {
			m["url"] = *v.URL
		}
		if v.Username != nil {
			m["username"] = *v.Username
		}
		if v.Password != nil {
			m["password"] = *v.Password
		}
		if v.Revision != nil {
			m["revision"] = *v.Revision
		}
		if v.SSHKey != nil {
			m["ssh_key"] = *v.SSHKey
		}
		nv = append(nv, m)
	}

	err := d.Set("custom_cookbooks_source", nv)
	if err != nil {
		// should never happen
		panic(err)
	}
}

func (d *awsResourceData) opsworksStackAttributes() *map[string]*string {
	attr := make(map[string]*string)
	color := d.getAwsString("color")
	if color != nil {
		attr["Color"] = color
	}
	return &attr
}

func (d *awsResourceData) opsworksSetStackAttributes(attr *map[string]*string) {
	d.Set("color", (*attr)["Color"])
}

func resourceAwsOpsworksStackRead(nd *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn
	d := makeAwsResourceData(nd)

	req := &opsworks.DescribeStacksInput{
		StackIDs: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks stack: %s", d.Id())

	resp, err := client.DescribeStacks(req)
	if err != nil {
		if awsErr, ok := err.(aws.APIError); ok && awsErr.Code == "ResourceNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	stack := resp.Stacks[0]
	d.Set("name", stack.Name)
	d.Set("region", stack.Region)
	d.Set("default_instance_profile_arn", stack.DefaultInstanceProfileARN)
	d.Set("service_role_arn", stack.ServiceRoleARN)
	d.Set("default_availability_zone", stack.DefaultAvailabilityZone)
	d.Set("default_os", stack.DefaultOs)
	d.Set("default_root_device_type", stack.DefaultRootDeviceType)
	d.Set("default_ssh_key_name", stack.DefaultSSHKeyName)
	d.Set("default_subnet_id", stack.DefaultSubnetID)
	d.Set("hostname_theme", stack.HostnameTheme)
	d.Set("use_custom_cookbooks", stack.UseCustomCookbooks)
	d.Set("use_opsworks_security_groups", stack.UseOpsWorksSecurityGroups)
	d.Set("vpc_id", stack.VPCID)
	d.opsworksSetStackAttributes(stack.Attributes)
	d.opsworksSetStackChefConfiguration(stack.ChefConfiguration)
	d.opsworksSetStackConfigurationManager(stack.ConfigurationManager)
	d.opsworksSetStackCustomCookbooksSource(stack.CustomCookbooksSource)

	return nil
}

func resourceAwsOpsworksStackCreate(nd *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn
	d := makeAwsResourceData(nd)

	err := resourceAwsOpsworksStackValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.CreateStackInput{
		Name: d.getAwsString("name"),
		DefaultInstanceProfileARN: d.getAwsString("default_instance_profile_arn"),
		Region:                    d.getAwsString("region"),
		ServiceRoleARN:            d.getAwsString("service_role_arn"),
		DefaultAvailabilityZone:   d.getAwsString("default_availability_zone"),
		DefaultOs:                 d.getAwsString("default_os"),
		DefaultRootDeviceType:     d.getAwsString("default_root_device_type"),
		DefaultSSHKeyName:         d.getAwsString("default_ssh_key_name"),
		DefaultSubnetID:           d.getAwsString("default_subnet_id"),
		HostnameTheme:             d.getAwsString("hostname_theme"),
		VPCID:                     d.getAwsString("vpc_id"),
		UseCustomCookbooks:        d.getAwsBool("use_custom_cookbooks"),
		UseOpsWorksSecurityGroups: d.getAwsBool("use_opsworks_security_groups"),
		ConfigurationManager:      d.opsworksStackConfigurationManager(),
		ChefConfiguration:         d.opsworksStackChefConfiguration(),
		CustomCookbooksSource:     d.opsworksStackCustomCookbooksSource(),
		Attributes:                d.opsworksStackAttributes(),
	}

	log.Printf("[DEBUG] Creating OpsWorks stack: %s", *req.Name)

	resp, err := client.CreateStack(req)
	if err != nil {
		return err
	}

	stackId := *resp.StackID
	d.SetId(stackId)
	d.Set("id", stackId)

	return resourceAwsOpsworksStackRead(nd, meta)
}

func resourceAwsOpsworksStackUpdate(nd *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn
	d := makeAwsResourceData(nd)

	err := resourceAwsOpsworksStackValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.UpdateStackInput{
		StackID: aws.String(d.Id()),
		Name:    d.getAwsString("name"),
		DefaultInstanceProfileARN: d.getAwsString("default_instance_profile_arn"),
		ServiceRoleARN:            d.getAwsString("service_role_arn"),
		DefaultAvailabilityZone:   d.getAwsString("default_availability_zone"),
		DefaultOs:                 d.getAwsString("default_os"),
		DefaultRootDeviceType:     d.getAwsString("default_root_device_type"),
		DefaultSSHKeyName:         d.getAwsString("default_ssh_key_name"),
		DefaultSubnetID:           d.getAwsString("default_subnet_id"),
		HostnameTheme:             d.getAwsString("hostname_theme"),
		UseCustomCookbooks:        d.getAwsBool("use_custom_cookbooks"),
		UseOpsWorksSecurityGroups: d.getAwsBool("use_opsworks_security_groups"),
		ConfigurationManager:      d.opsworksStackConfigurationManager(),
		ChefConfiguration:         d.opsworksStackChefConfiguration(),
		CustomCookbooksSource:     d.opsworksStackCustomCookbooksSource(),
		Attributes:                d.opsworksStackAttributes(),
	}

	log.Printf("[DEBUG] Updating OpsWorks stack: %s", d.Id())

	_, err = client.UpdateStack(req)
	if err != nil {
		return err
	}

	return resourceAwsOpsworksStackRead(nd, meta)
}

func resourceAwsOpsworksStackDelete(nd *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn
	d := makeAwsResourceData(nd)

	req := &opsworks.DeleteStackInput{
		StackID: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks stack: %s", d.Id())

	_, err := client.DeleteStack(req)
	return err
}
