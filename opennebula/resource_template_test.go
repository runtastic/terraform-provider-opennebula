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

func TestAccTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.test", "name", "test-me"),
					resource.TestCheckResourceAttr("opennebula_template.test", "permissions", "642"),
					resource.TestCheckResourceAttrSet("opennebula_template.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_template.test", "reg_time"),
					testAccCheckTemplateAttributes(map[string]string{"FOO": "bar"}),
					testAccCheckTemplatePermissions(&Permissions{
						Owner_U: 1,
						Owner_M: 1,
						Group_U: 1,
						Other_M: 1,
					}),
				),
			},
			{
				Config: testAccTemplateConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.test", "permissions", "600"),
					testAccCheckTemplateAttributes(map[string]string{"BAR": "foo"}),
					testAccCheckTemplatePermissions(&Permissions{
						Owner_U: 1,
						Owner_M: 1,
					}),
				),
			},
		},
	})
}

func testAccCheckTemplateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		_, err := client.Call("one.template.info", intId(rs.Primary.ID), false)
		if err == nil {
			return fmt.Errorf("Expected template %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTemplateAttributes(attrs map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)

		for _, rs := range s.RootModule().Resources {
			resp, err := client.Call("one.template.info", intId(rs.Primary.ID), false)
			if err != nil {
				return fmt.Errorf("Expected template %s to exist", rs.Primary.ID)
			}

			for k, v := range attrs {
				if !strings.Contains(resp, fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k)) {
					return fmt.Errorf("Expected template to contain attribute %s=%s, specified in the description. The template contents were %s", k, v, resp)
				}
			}
		}

		return nil
	}
}

func testAccCheckTemplatePermissions(expected *Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*Client)

		for _, rs := range s.RootModule().Resources {
			resp, err := client.Call("one.template.info", intId(rs.Primary.ID), false)
			if err != nil {
				return fmt.Errorf("Expected template %s to exist", rs.Primary.ID)
			}

			var tmpl UserTemplate
			if err = xml.Unmarshal([]byte(resp), &tmpl); err != nil {
				return err
			}

			if !reflect.DeepEqual(tmpl.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for template %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionString(expected),
					permissionString(tmpl.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccTemplateConfigBasic = `
resource "opennebula_template" "test" {
  name = "test-me"
  description = <<EOF
	FOO = "bar"
  EOF
  permissions = "642"
}
`

var testAccTemplateConfigUpdate = `
resource "opennebula_template" "test" {
  name = "test-me"
  description = <<EOF
	FOO = "bar"
	BAR = "foo"
  EOF
  permissions = "600"
}
`
