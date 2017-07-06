package opennebula

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strconv"
	"strings"
)

type Image struct {
	Name        string       `xml:"NAME"`
	Id          int          `xml:"ID"`
	Uid         int          `xml:"UID"`
	Gid         int          `xml:"GID"`
	Uname       string       `xml:"UNAME"`
	Gname       string       `xml:"GNAME"`
	Permissions *Permissions `xml:"PERMISSIONS"`
	RegTime     string       `xml:"REG"`
	Size        int          `xml:"SIZE"`
	State       int          `xml:"STATE"`
	Source      string       `xml:"SOURCE"`
	Path        string       `xml:"PATH"`
	Persistent  string       `xml:"PERSISTENT"`
	DatastoreID int          `xml:"DATASTORE_ID"`
	Datastore   string       `xml:"DATASTORE"`
	FsType      string       `xml:"FSTYPE"`
	RunningVMs  int          `xml:"RUNNING_VMS"`
}

type Images struct {
	Image []*Image `xml:"IMAGE"`
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
				Optional:    true,
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
			"image_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the Image to be cloned from. If Image ID is not set, a new Image will be created",
			},
			"datastore_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the datastore where Image will be stored",
			},
		},
	}
}

func resourceImageCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	// Check if Image ID for cloning is set
	if d.Get("image_id") != nil {
		return resourceImageClone(d, meta)
	}

	// Create base object
	resp, err := client.Call(
		"one.image.allocate",
		fmt.Sprintf("NAME = \"%s\"\n", d.Get("name").(string))+d.Get("description").(string),
		d.Get("datastore_id"),
	)
	if err != nil {
		return err
	}

	d.SetId(resp)
	// update permisions
	if _, err = changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client, "one.image.chmod"); err != nil {
		return err
	}

	return resourceImageRead(d, meta)
}

func resourceImageClone(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	// Create base object
	resp, err := client.Call(
		"one.image.clone",
		d.Get("image_id"),
		d.Get("name"),
		-1,
	)
	if err != nil {
		return err
	}

	d.SetId(resp)
	// update permisions
	if _, err = changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client, "one.image.chmod"); err != nil {
		return err
	}

	return resourceImageRead(d, meta)
}

func resourceImageRead(d *schema.ResourceData, meta interface{}) error {
	var img *Image
	var imgs *Images

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

		for _, t := range imgs.Image {
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

	d.SetId(strconv.Itoa(img.Id))
	d.Set("name", img.Name)
	d.Set("uid", img.Uid)
	d.Set("gid", img.Gid)
	d.Set("uname", img.Uname)
	d.Set("gname", img.Gname)
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
