package opennebula

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"strings"
	"testing"
)

func TestAccVnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVnetConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_vnet.test", "name", "test-vnet"),
					resource.TestCheckResourceAttr("opennebula_vnet.test", "bridge", "br-test"),
					resource.TestCheckResourceAttr("opennebula_vnet.test", "ip_start", "192.168.0.1"),
					resource.TestCheckResourceAttr("opennebula_vnet.test", "ip_size", "10"),
					resource.TestCheckResourceAttr("opennebula_vnet.test", "permissions", "642"),
					resource.TestCheckResourceAttrSet("opennebula_vnet.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_vnet.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_vnet.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_vnet.test", "gname"),
					testAccCheckVnetAttributes(map[string]string{"FOO": "bar"}),
					testAccCheckVnetPermissions(&Permissions{
						Owner_U: 1,
						Owner_M: 1,
						Group_U: 1,
						Other_M: 1,
					}),
				),
			},
			{
				Config: testAccVnetConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_vnet.test", "permissions", "700"),
					testAccCheckVnetAttributes(map[string]string{"FOO": "bar2"}),
					testAccCheckVnetPermissions(&Permissions{
						Owner_U: 1,
						Owner_M: 1,
						Owner_A: 1,
					}),
				),
			},
		},
	})
}

func testAccCheckVnetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		_, err := client.Call("one.vn.info", intId(rs.Primary.ID), false)
		if err == nil {
			return fmt.Errorf("Expected vnet %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckVnetAttributes(attrs map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)

		for _, rs := range s.RootModule().Resources {
			resp, err := client.Call("one.vn.info", intId(rs.Primary.ID), false)
			if err != nil {
				return fmt.Errorf("Expected vnet %s to exist when checking attributes", rs.Primary.ID)
			}

			for k, v := range attrs {
				if !strings.Contains(resp, fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k)) {
					return fmt.Errorf("Expected vnet to contain attribute %s=%s, specified in the description. The vnet contents were %s", k, v, resp)
				}
			}
		}

		return nil
	}
}

func testAccCheckVnetPermissions(expected *Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)

		for _, rs := range s.RootModule().Resources {
			resp, err := client.Call("one.vn.info", intId(rs.Primary.ID), false)
			if err != nil {
				return fmt.Errorf("Expected vnet %s to exist when checking permissions", rs.Primary.ID)
			}

			var vnet UserVnet
			if err = xml.Unmarshal([]byte(resp), &vnet); err != nil {
				return err
			}

			if !reflect.DeepEqual(vnet.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for vnet %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionString(expected),
					permissionString(vnet.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccVnetConfigBasic = `
resource "opennebula_vnet" "test" {
  name = "test-vnet"
  description = <<EOF
	FOO = "bar"
  PHYDEV=""
  SECURITY_GROUPS="0"
  VN_MAD="dummy"
  EOF
  bridge = "br-test"
  ip_start = "192.168.0.1"
  ip_size = 10
  permissions = "642"
}
`

var testAccVnetConfigUpdate = `
resource "opennebula_vnet" "test" {
  name = "test-vnet"
  description = <<EOF
	FOO = "bar2"
  PHYDEV=""
  SECURITY_GROUPS="0"
  VN_MAD="dummy"
  EOF
  bridge = "br-test"
  ip_start = "192.168.0.10"
  ip_size = 20
  permissions = "700"
}
`
