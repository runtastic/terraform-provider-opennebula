package opennebula

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
	"strings"
)

type UserTemplates struct {
	UserTemplate []*UserTemplate `xml:"VMTEMPLATE"`
}

type UserTemplate struct {
	Name        string       `xml:"NAME"`
	Id          int          `xml:"ID"`
	Uid         int          `xml:"UID"`
	Gid         int          `xml:"GID"`
	Uname       string       `xml:"UNAME"`
	Gname       string       `xml:"GNAME"`
	RegTime     int          `xml:"REGTIME"`
	Permissions *Permissions `xml:"PERMISSIONS"`
}

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Exists: resourceTemplateExists,
		Update: resourceTemplateUpdate,
		Delete: resourceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the template",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the template, in OpenNebula's XML or String format",
			},
			"permissions": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Permissions for the template (in Unix format, owner-group-other, use-manage-admin)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) != 3 {
						errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
					}

					all := true
					for _, c := range strings.Split(value, "") {
						if c < "0" || c > "7" {
							all = false
						}
					}
					if !all {
						errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
					}

					return
				},
			},

			"uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the user that will own the template",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the template",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the template",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the template",
			},
			"reg_time": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Registration time",
			},
		},
	}
}

func resourceTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	resp, err := client.Call(
		"one.template.allocate",
		fmt.Sprintf("NAME = \"%s\"\n", d.Get("name").(string))+d.Get("description").(string),
	)
	if err != nil {
		return err
	}

	d.SetId(resp)

	if _, err = changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client); err != nil {
		return err
	}

	return resourceTemplateRead(d, meta)
}

func resourceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	var tmpl *UserTemplate
	var tmpls *UserTemplates

	client := meta.(*Client)
	found := false

	// Try to find the template by ID, if specified
	if d.Id() != "" {
		resp, err := client.Call("one.template.info", intId(d.Id()), false)
		if err == nil {
			found = true
			if err = xml.Unmarshal([]byte(resp), &tmpl); err != nil {
				return err
			}
		} else {
			log.Printf("Could not find template by ID %s", d.Id())
		}
	}

	// Otherwise, try to find the template by (user, name) as the de facto compound primary key
	if d.Id() == "" || !found {
		resp, err := client.Call("one.templatepool.info", -3, -1, -1)
		if err != nil {
			return err
		}

		if err = xml.Unmarshal([]byte(resp), &tmpls); err != nil {
			return err
		}

		for _, t := range tmpls.UserTemplate {
			if t.Name == d.Get("name").(string) {
				tmpl = t
				found = true
				break
			}
		}

		if !found || tmpl == nil {
			d.SetId("")
			log.Printf("Could not find template with name %s for user %s", d.Get("name").(string), client.Username)
			return nil
		}
	}

	d.SetId(strconv.Itoa(tmpl.Id))
	d.Set("name", tmpl.Name)
	d.Set("uid", tmpl.Uid)
	d.Set("gid", tmpl.Gid)
	d.Set("uname", tmpl.Uname)
	d.Set("gname", tmpl.Gname)
	d.Set("reg_time", tmpl.RegTime)
	d.Set("permissions", permissionString(tmpl.Permissions))

	return nil
}

func resourceTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceTemplateRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	if d.HasChange("description") {
		_, err := client.Call(
			"one.template.update",
			intId(d.Id()),
			d.Get("description").(string),
			0, // replace the whole template instead of merging it with the existing one
		)
		if err != nil {
			return err
		}
	}

	if d.HasChange("permissions") {
		resp, err := changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated template %s\n", resp)
	}

	return nil
}

func resourceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceTemplateRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	client := meta.(*Client)
	resp, err := client.Call("one.template.delete", intId(d.Id()), false)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted template %s\n", resp)
	return nil
}

func changePermissions(id int, p *Permissions, client *Client) (string, error) {
	return client.Call(
		"one.template.chmod",
		id,
		p.Owner_U,
		p.Owner_M,
		p.Owner_A,
		p.Group_U,
		p.Group_M,
		p.Group_A,
		p.Other_U,
		p.Other_M,
		p.Other_A,
		false, // recursive (do not change the associated images' permissions)
	)
}
