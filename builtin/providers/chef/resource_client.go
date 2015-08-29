package chef

import (
	"github.com/hashicorp/terraform/helper/schema"

	chefc "github.com/go-chef/chef"
)

func resourceChefClient() *schema.Resource {
	return &schema.Resource{
		Create: CreateClient,
		Read:   ReadClient,
		Delete: DeleteClient,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"private_key_pem": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key_pem": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func CreateClient(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*chefc.Client)

	name := d.Get("name").(string)
	admin := d.Get("admin").(bool)

	result, err := client.Clients.Create(name, admin)
	if err != nil {
		return err
	}

	d.SetId(name)
	d.Set("private_key_pem", result.PrivateKey)
	return ReadClient(d, meta)
}

func ReadClient(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*chefc.Client)

	name := d.Id()
	result, err := client.Clients.Get(name)
	if err != nil {
		if errRes, ok := err.(*chefc.ErrorResponse); ok {
			if errRes.Response.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("public_key_pem", result.PublicKey)
	d.Set("admin", result.Admin)

	return nil
}

func DeleteClient(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*chefc.Client)

	name := d.Id()

	err := client.Clients.Delete(name)
	if err == nil {
		d.SetId("")
	}
	return err
}
