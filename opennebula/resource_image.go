package opennebula

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net"
	"strconv"
	"strings"
)

type UserImages struct {
	UserImage []*UserImages `xml:"IMAGE"`
}

type UserImage struct {
	Name        string       `xml:"NAME"`
	Id          int          `xml:"ID"`
	Uid         int          `xml:"UID"`
	Gid         int          `xml:"GID"`
	Uname       string       `xml:"UNAME"`
	Gname       string       `xml:"GNAME"`
	Permissions *Permissions `xml:"PERMISSIONS"`
	DatastoreID int          `xml:"DATASTORE_ID"`
	Persistent  string       `xml:"PERSISTENT"`
}

func resourceImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceImageCreate,
		Read:   resourceImageRead,
		Exists: resourceImageExists,
		Update: resourceImageUpdate,
		Delete: resourceImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Image",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the Image, in OpenNebula's XML or String format",
			},
			"permissions": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Permissions for the Image (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the Image",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the Image",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the Image",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the Image",
			},
			"imageid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the Image to be cloned from",
			},
		},
	}
}

func resourceImageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	// Create base object
	resp, err := client.Call(
		"one.image.clone",
		d.Get("imageid"),
		d.Get("name"),
		-1,
	)
	if err != nil {
		return err
	}

	d.SetId(resp)
	// update permisions
	if _, err = changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client, "one.vn.chmod"); err != nil {
		return err
	}

	return resourceImageRead(d, meta)
}

func resourceImageRead(d *schema.ResourceData, meta interface{}) error {
	var img *UserImage
	var imgs *UserImages

	client := meta.(*Client)
	found := false

	// Try to find the Image by ID, if specified
	if d.Id() != "" {
		resp, err := client.Call("one.image.info", intId(d.Id()), false)
		if err == nil {
			found = true
			if err = xml.Unmarshal([]byte(resp), &img); err != nil {
				return err
			}
		} else {
			log.Printf("Could not find Image by ID %s", d.Id())
		}
	}

	// Otherwise, try to find the Image by (user, name) as the de facto compound primary key
	if d.Id() == "" || !found {
		resp, err := client.Call("one.imagepool.info", -3, -1, -1)
		if err != nil {
			return err
		}

		if err = xml.Unmarshal([]byte(resp), &imgs); err != nil {
			return err
		}

		for _, t := range imgs.UserImage {
			if t.Name == d.Get("name").(string) {
				img = t
				found = true
				break
			}
		}

		if !found || img == nil {
			d.SetId("")
			log.Printf("Could not find Image with name %s for user %s", d.Get("name").(string), client.Username)
			return nil
		}
	}

	d.SetId(strconv.Itoa(vn.Id))
	d.Set("name", img.Name)
	d.Set("uid", img.Uid)
	d.Set("gid", img.Gid)
	d.Set("uname", img.Uname)
	d.Set("gname", img.Gname)
	d.Set("imageid", img.Imageid)
	d.Set("permissions", permissionString(img.Permissions))

	return nil
}

func resourceImageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceImageRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceImageUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	if d.HasChange("description") {
		_, err := client.Call(
			"one.image.update",
			intId(d.Id()),
			d.Get("description").(string),
			0, // replace the whole image instead of merging it with the existing one
		)
		if err != nil {
			return err
		}
	}

	if d.HasChange("name") {
		resp, err := client.Call(
			"one.image.rename",
			intId(d.Id()),
			d.Get("name").(string),
		)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for Image %s\n", resp)
	}

	if d.HasChange("permissions") {
		resp, err := changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client, "one.image.chmod")
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated Image %s\n", resp)
	}

	return nil
}

func resourceImageDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceImageRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	client := meta.(*Client)

	resp, err := client.Call("one.image.delete", intId(d.Id()), false)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Image %s\n", resp)
	return nil
}
