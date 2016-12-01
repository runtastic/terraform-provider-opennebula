package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/megamsys/opennebula-go/api"
	"github.com/megamsys/opennebula-go/template"
	"errors"
	"strconv"
	"log"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Update:   resourceTemplateUpdate,
		Delete: resourceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// User Template
			"uid": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "ID of the user that will own the template",
			},
			"gid": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "ID of the group that will own the template",
			},
			"uname": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Name of the user that will own the template",
			},
			"gname": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Name of the group that will own the template",
			},
			"reg_time": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "Registration time",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the template",
			},
			"permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Description: "Granular permissions for the template",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner_u": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"owner_m": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"owner_a": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"group_u": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"group_m": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"group_a": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"other_u": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"other_m": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"other_a": {
							Type: schema.TypeInt,
							Optional: true,
						},
					},
				},
			},


			// Template
			"context": {
				Type:     schema.TypeSet,
				Optional: true,
				Description: "Contextualized settings to be passed to a Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type: schema.TypeString,
							Optional: true,
						},
						"files": {
							Type: schema.TypeString,
							Optional: true,
						},
						"ssh_public_key": {
							Type: schema.TypeString,
							Optional: true,
						},
						"set_hostname": {
							Type: schema.TypeString,
							Optional: true,
						},
						"node_name": {
							Type: schema.TypeString,
							Optional: true,
						},
						"accounts_id": {
							Type: schema.TypeString,
							Optional: true,
						},
						"platform_id": {
							Type: schema.TypeString,
							Optional: true,
						},
						"assembly_id": {
							Type: schema.TypeString,
							Optional: true,
						},
						"assemblies_id": {
							Type: schema.TypeString,
							Optional: true,
						},
						"org_id": {
							Type: schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Number of physical cores to allocate",
			},
			"cpu_cost": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"logo": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory_cost": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sunstone_capacity_select": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sunstone_network_select": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vcpu": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"graphics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"listen": {
							Type: schema.TypeString,
							Optional: true,
						},
						"port": {
							Type: schema.TypeString,
							Optional: true,
						},
						"type": {
							Type: schema.TypeString,
							Optional: true,
						},
						"random_passwd": {
							Type: schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"disk": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"driver": {
							Type: schema.TypeString,
							Optional: true,
						},
						"image": {
							Type: schema.TypeString,
							Optional: true,
						},
						"image_uname": {
							Type: schema.TypeString,
							Optional: true,
						},
						"size": {
							Type: schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"disk_cost": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"from_app": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"from_app_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"nic": {
				Type: schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type: schema.TypeString,
							Optional: true,
						},
						"network_uname": {
							Type: schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"os": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arch": {
							Type: schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"sched_requirements": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sched_ds_requirements": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getNestedSet(d *schema.ResourceData, k string) map[string]interface{} {
	setList := d.Get(k).(*schema.Set).List()
	if len(setList) == 0 {
		return nil
	}

	return setList[0].(map[string]interface{})
}

func resourceTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("jason says: creating...")

	userTemplate := &template.UserTemplate{
		T: meta.(*api.Rpc),
		Template: &template.Template{
			Name: d.Get("name").(string),
			Cpu: d.Get("cpu").(string),
			Cpu_cost: d.Get("cpu_cost").(string),
			Description: d.Get("description").(string),
			Hypervisor: d.Get("hypervisor").(string),
			Logo: d.Get("logo").(string),
			Memory: d.Get("memory_cost").(string),
			Sunstone_capacity_select: d.Get("sunstone_capacity_select").(string),
			Sunstone_Network_select: d.Get("sunstone_network_select").(string),
			VCpu: d.Get("vcpu").(string),
			Disk_cost: d.Get("disk_cost").(string),
			From_app: d.Get("from_app").(string),
			From_app_name: d.Get("from_app_name").(string),
			Sched_requirments: d.Get("sched_requirements").(string),
			Sched_ds_requirments: d.Get("sched_ds_requirements").(string),
		},
	}

	// TODO: Clean up a bit
	permissions := getNestedSet(d, "permissions")
	if permissions != nil {
		userTemplate.Permissions = &template.Permissions{
			Owner_U: permissions["owner_u"].(int),
			Owner_M: permissions["owner_m"].(int),
			Owner_A: permissions["owner_a"].(int),
			Group_U: permissions["group_u"].(int),
			Group_M: permissions["group_m"].(int),
			Group_A: permissions["group_a"].(int),
			Other_U: permissions["other_u"].(int),
			Other_M: permissions["other_m"].(int),
			Other_A: permissions["other_a"].(int),
		}
	}

	context := getNestedSet(d, "context")
	if context != nil {
		userTemplate.Template.Context = &template.Context{
			Network: context["network"].(string),
			SSH_Public_key: context["ssh_public_key"].(string),
			Files: context["files"].(string),
			Set_Hostname: context["set_hostname"].(string),
			Node_Name: context["node_name"].(string),
			Accounts_id: context["accounts_id"].(string),
			Platform_id: context["platform_id"].(string),
			Assembly_id: context["assembly_id"].(string),
			Assemblies_id: context["assemblies_id"].(string),
			Org_id: context["org_id"].(string),
		}
	}

	graphics := getNestedSet(d, "graphics")
	if graphics != nil {
		userTemplate.Template.Graphics = &template.Graphics{
			Listen: graphics["listen"].(string),
			Port: graphics["port"].(string),
			Type: graphics["type"].(string),
			RandomPassWD: graphics["random_passwd"].(string),
		}
	}


	disk := getNestedSet(d, "disk")
	if disk != nil {
		userTemplate.Template.Disk = &template.Disk{
			Driver: disk["driver"].(string),
			Image: disk["image"].(string),
			Image_Uname: disk["image_uname"].(string),
			Size: disk["size"].(string),
		}
	}


	os := getNestedSet(d, "os")
	if os != nil {
		userTemplate.Template.Os = &template.OS{
			Arch: os["arch"].(string),
		}
	}

	nicList := d.Get("nic").(*schema.Set).List()
	nics := make([]*template.NIC, len(nicList))
	for i, nic := range nicList {
		mapN := nic.(map[string]interface{})
		nics[i] = &template.NIC{
			Network: mapN["network"].(string),
			Network_uname: mapN["network_uname"].(string),
		}
	}
	userTemplate.Template.Nic = nics

	xmlRpc, err := userTemplate.AllocateTemplate()
	if err != nil {
		log.Println("[FATAL] Could not allocate template", err)
		return err
	}

	log.Println("[INFO] Creating a template returned response", xmlRpc)
	d.SetId(xmlRpc.(string))

	return resourceTemplateRead(d, meta)
}

func resourceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	templateReq := &template.TemplateReqs{
		T: meta.(*api.Rpc),
		TemplateName: name,
	}
	userTemplates, err := templateReq.Get()

	if err != nil {
		log.Println("[FATAL] Could not read template", err)
		return err
	}

	if len(userTemplates) == 0 {
		log.Printf("[FATAL] No template with name %s\n", name)
		return errors.New("No template with such name")
	}

	// TODO: PK (user_id, name)
	if len(userTemplates) != 1 {
		log.Printf("[WARN] Reading template %s returned %d different templates\n", name, len(userTemplates))
	}

	t := userTemplates[0]
	d.SetId(strconv.Itoa(t.Id))

	d.Set("uid", t.Uid)
	d.Set("gid", t.Gid)
	d.Set("uname", t.Uname)
	d.Set("gname", t.Gname)
	d.Set("reg_time", t.RegTime)
	d.Set("name", t.Name)

	d.Set("cpu", t.Template.Cpu)
	d.Set("cpu_cost", t.Template.Cpu_cost)
	d.Set("description", t.Template.Description)
	d.Set("hypervisor", t.Template.Hypervisor)
	d.Set("logo", t.Template.Logo)
	d.Set("memory", t.Template.Memory)
	d.Set("memory_cost", t.Template.Memory_cost)
	d.Set("sunstone_capacity_select", t.Template.Sunstone_capacity_select)
	d.Set("sunstone_network_select", t.Template.Sunstone_Network_select)
	d.Set("vcpu", t.Template.VCpu)
	d.Set("sched_requirements", t.Template.Sched_requirments)
	d.Set("sched_ds_requirements", t.Template.Sched_ds_requirments)
	d.Set("disk_cost", t.Template.Disk_cost)
	d.Set("from_app", t.Template.From_app)
	d.Set("from_app_name", t.Template.From_app_name)

	if t.Permissions != nil {
		d.Set("permissions", []map[string]interface{}{
			{
				"owner_u": t.Permissions.Owner_U,
				"owner_m": t.Permissions.Owner_M,
				"owner_a": t.Permissions.Owner_A,
				"group_u": t.Permissions.Group_U,
				"group_m": t.Permissions.Group_M,
				"group_a": t.Permissions.Group_A,
				"other_u": t.Permissions.Other_U,
				"other_m": t.Permissions.Other_M,
				"other_a": t.Permissions.Other_A,
			},
		})
	}

	if t.Template.Context != nil {
		d.Set("context", []map[string]interface{}{
			{
				"network": t.Template.Context.Network,
				"files": t.Template.Context.Files,
				"ssh_public_key": t.Template.Context.SSH_Public_key,
				"set_hostname": t.Template.Context.Set_Hostname,
				"node_name": t.Template.Context.Node_Name,
				"accounts_id": t.Template.Context.Accounts_id,
				"platform_id": t.Template.Context.Platform_id,
				"assembly_id": t.Template.Context.Assembly_id,
				"assemblies_id": t.Template.Context.Assemblies_id,
				"org_id": t.Template.Context.Org_id,
			},
		})
	}


	if t.Template.Graphics != nil {
		d.Set("graphics", []map[string]interface{}{
			{
				"listen": t.Template.Graphics.Listen,
				"port": t.Template.Graphics.Port,
				"type": t.Template.Graphics.Type,
				"random_passwd": t.Template.Graphics.RandomPassWD,
			},
		})
	}

	if t.Template.Disk != nil {
		d.Set("disk", []map[string]interface{}{
			{
				"driver": t.Template.Disk.Driver,
				"image": t.Template.Disk.Image,
				"image_uname": t.Template.Disk.Image_Uname,
				"size": t.Template.Disk.Size,
			},
		})
	}

	nics := make([]map[string]interface{}, len(t.Template.Nic))
	for i, n := range t.Template.Nic {
		nics[i] = map[string]interface{}{
			"network": n.Network,
			"network_uname": n.Network_uname,
		}
	}
	d.Set("nic", nics)

	if t.Template.Os != nil {
		d.Set("os", []map[string]interface{}{
			{
				"arch": t.Template.Os.Arch,
			},
		})
	}

	return nil
}

func resourceTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Rpc)
	err := resourceTemplateRead(d, meta)
	if err != nil {
		return nil
	}

	intId, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	args := []interface{}{client.Key, intId, false, 0}
	if _, err := client.Call("one.template.delete", args); err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted template %s\n", d.Get("name").(string))
	return nil
}
