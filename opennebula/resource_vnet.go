package opennebula

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
	"strings"
)

type UserVnets struct {
	UserVnet []*UserVnet `xml:"VNET"`
}

type UserVnet struct {
	Name                    string       `xml:"NAME"`
	Id                      int          `xml:"ID"`
	Uid                     int          `xml:"UID"`
	Gid                     int          `xml:"GID"`
	Uname                   string       `xml:"UNAME"`
	Gname                   string       `xml:"GNAME"`
	Permissions							*Permissions `xml:"PERMISSIONS"`
  // Clusters is an array of IDs minOccurs="0" maxOccurs="unbounded" - will make it one single one for now
  Clusers									[1]int       `xml:"CLUSTERS"`
	Bridge                  string       `xml:"BRIDGE"`
	Parent_Network_Id       string       `xml:"PARENT_NETWORK_ID"`
	VN_Mad                  string       `xml:"VN_MAD"`
	Phydev                  string       `xml:"PHYDEV"`
	Vlan_id                 string       `xml:"VLAN_ID"`
	Vlan_id_automatic       string       `xml:"VLAN_ID_AUTOMATIC"`
	Used_Leases             int          `xml:"USED_LEASES"`
  // Vrouters is an array of IDs minOccurs="0" maxOccurs="unbounded" - will make it one single one for now
	Vrouters                [1]int       `xml:"VROUTERS"`
	ARpool                  *ARPool      `xml:"AR_POOL"`

//        <xs:element name="TEMPLATE" type="xs:anyType"/>

}

func resourceVnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceVnetCreate,
		Read:   resourceVnetRead,
		Exists: resourceVnetExists,
		Update: resourceVnetUpdate,
		Delete: resourceVnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the vnet",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the vnet, in OpenNebula's XML or String format",
			},
			"permissions": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Permissions for the vnet (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the vnet",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the vnet",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the vnet",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the vnet",
			},
			"clusters": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IDs of the clusters to which the vnet should be associated",
			},
			"Bridge": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the bridge interface to which the vnet should be associated",
			},
			"vn_mad": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vn_mad",
			},
		},
	}
}

func resourceVnetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	resp, err := client.Call(
		"one.vn.allocate",
		fmt.Sprintf("NAME = \"%s\"\n", d.Get("name").(string))+d.Get("description").(string),
	)
	if err != nil {
		return err
	}

	d.SetId(resp)

	if _, err = changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client); err != nil {
		return err
	}

	return resourceVnetRead(d, meta)
}

func resourceVnetRead(d *schema.ResourceData, meta interface{}) error {
	var vn *UserVnet
	var vns *UserVnets

	client := meta.(*Client)
	found := false

	// Try to find the template by ID, if specified
	if d.Id() != "" {
		resp, err := client.Call("one.vn.info", intId(d.Id()), false)
		if err == nil {
			found = true
			if err = xml.Unmarshal([]byte(resp), &vn); err != nil {
				return err
			}
		} else {
			log.Printf("Could not find vnet by ID %s", d.Id())
		}
	}

	// Otherwise, try to find the template by (user, name) as the de facto compound primary key
	if d.Id() == "" || !found {
		resp, err := client.Call("one.vnpool.info", -3, -1, -1)
		if err != nil {
			return err
		}

		if err = xml.Unmarshal([]byte(resp), &vns); err != nil {
			return err
		}

		for _, t := range vns.UserTemplate {
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

func resourceVnetExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceTemplateRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceVnetUpdate(d *schema.ResourceData, meta interface{}) error {
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

func resourceVnetDelete(d *schema.ResourceData, meta interface{}) error {
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
		"one.vn.chmod",
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
