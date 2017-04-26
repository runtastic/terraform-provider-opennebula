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

type UserVnets struct {
	UserVnet []*UserVnet `xml:"VNET"`
}

type UserVnet struct {
	Name        string       `xml:"NAME"`
	Id          int          `xml:"ID"`
	Uid         int          `xml:"UID"`
	Gid         int          `xml:"GID"`
	Uname       string       `xml:"UNAME"`
	Gname       string       `xml:"GNAME"`
	Permissions *Permissions `xml:"PERMISSIONS"`
	Bridge      string       `xml:"BRIDGE"`
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
			"bridge": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the bridge interface to which the vnet should be associated",
			},
			"ip_start": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Start IP of the range to be allocated",
			},
			"ip_size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Size (in number) of the ip range",
			},
			"reservation_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Carve a network reservation of this size from the reservation starting from `ip-start`",
			},
		},
	}
}

func resourceVnetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	// Create base object
	resp, err := client.Call(
		"one.vn.allocate",
		fmt.Sprintf("NAME = \"%s\"\n", d.Get("name").(string))+d.Get("description").(string)+"\nBRIDGE="+d.Get("bridge").(string),
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
	// add address range and reservations
	var address_range_string = `AR = [
  TYPE = IP4,
  IP = %s,
  SIZE = %d ]`
	_, a_err := client.Call(
		"one.vn.add_ar",
		intId(d.Id()),
		fmt.Sprintf(address_range_string, d.Get("ip_start").(string), d.Get("ip_size").(int)),
	)

	if a_err != nil {
		return a_err
	}

	if d.Get("reservation_size").(int) > 0 {
		// add address range and reservations
		ip := net.ParseIP(d.Get("ip_start").(string))
		ip = ip.To4()

		for i := 0; i < d.Get("reservation_size").(int); i++ {
			var address_reservation_string = `LEASES=[IP=%s]`
			_, r_err := client.Call(
				"one.vn.hold",
				intId(d.Id()),
				fmt.Sprintf(address_reservation_string, ip),
			)

			if r_err != nil {
				return r_err
			}

			ip[3]++
		}

	}

	return resourceVnetRead(d, meta)
}

func resourceVnetRead(d *schema.ResourceData, meta interface{}) error {
	var vn *UserVnet
	var vns *UserVnets

	client := meta.(*Client)
	found := false

	// Try to find the vnet by ID, if specified
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

	// Otherwise, try to find the vnet by (user, name) as the de facto compound primary key
	if d.Id() == "" || !found {
		resp, err := client.Call("one.vnpool.info", -3, -1, -1)
		if err != nil {
			return err
		}

		if err = xml.Unmarshal([]byte(resp), &vns); err != nil {
			return err
		}

		for _, t := range vns.UserVnet {
			if t.Name == d.Get("name").(string) {
				vn = t
				found = true
				break
			}
		}

		if !found || vn == nil {
			d.SetId("")
			log.Printf("Could not find vnet with name %s for user %s", d.Get("name").(string), client.Username)
			return nil
		}
	}

	d.SetId(strconv.Itoa(vn.Id))
	d.Set("name", vn.Name)
	d.Set("uid", vn.Uid)
	d.Set("gid", vn.Gid)
	d.Set("uname", vn.Uname)
	d.Set("gname", vn.Gname)
	d.Set("bridge", vn.Bridge)
	d.Set("permissions", permissionString(vn.Permissions))

	return nil
}

func resourceVnetExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceVnetRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceVnetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	if d.HasChange("description") {
		_, err := client.Call(
			"one.vn.update",
			intId(d.Id()),
			d.Get("description").(string),
			0, // replace the whole vnet instead of merging it with the existing one
		)
		if err != nil {
			return err
		}
	}

	if d.HasChange("name") {
		resp, err := client.Call(
			"one.vn.rename",
			intId(d.Id()),
			d.Get("name").(string),
		)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for Vnet %s\n", resp)
	}

	if d.HasChange("ip_size") {
		var address_range_string = `AR = [
		AR_ID = 0,
		TYPE = IP4,
		IP = %s,
		SIZE = %d ]`
		resp, a_err := client.Call(
			"one.vn.update_ar",
			intId(d.Id()),
			fmt.Sprintf(address_range_string, d.Get("ip_start").(string), d.Get("ip_size").(int)),
		)

		if a_err != nil {
			return a_err
		}
		log.Printf("[INFO] Successfully updated size of address range for Vnet %s\n", resp)
	}

	if d.HasChange("ip_start") {
		log.Printf("[WARNING] Changing the IP address of the Vnet address range is currently not supported")
	}

	if d.HasChange("permissions") {
		resp, err := changePermissions(intId(d.Id()), permission(d.Get("permissions").(string)), client, "one.vn.chmod")
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated Vnet %s\n", resp)
	}

	return nil
}

func resourceVnetDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceVnetRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	client := meta.(*Client)
	resp, err := client.Call("one.vn.delete", intId(d.Id()), false)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Vnet %s\n", resp)
	return nil
}
