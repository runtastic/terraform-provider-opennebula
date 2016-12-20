package opennebula

import (
	"strconv"
	"strings"
)

type Permissions struct {
	Owner_U int `xml:"OWNER_U"`
	Owner_M int `xml:"OWNER_M"`
	Owner_A int `xml:"OWNER_A"`
	Group_U int `xml:"GROUP_U"`
	Group_M int `xml:"GROUP_M"`
	Group_A int `xml:"GROUP_A"`
	Other_U int `xml:"OTHER_U"`
	Other_M int `xml:"OTHER_M"`
	Other_A int `xml:"OTHER_A"`
}

func permissionString(p *Permissions) string {
	owner := p.Owner_U<<2 | p.Owner_M<<1 | p.Owner_A
	group := p.Group_U<<2 | p.Group_M<<1 | p.Group_A
	other := p.Other_U<<2 | p.Other_M<<1 | p.Other_A
	return strconv.Itoa(owner*100 + group*10 + other)
}

func permission(p string) *Permissions {
	perms := strings.Split(p, "")
	owner, _ := strconv.Atoi(perms[0])
	group, _ := strconv.Atoi(perms[1])
	other, _ := strconv.Atoi(perms[2])

	return &Permissions{
		Owner_U: owner & 4 >> 2,
		Owner_M: owner & 2 >> 1,
		Owner_A: owner & 1,
		Group_U: group & 4 >> 2,
		Group_M: group & 2 >> 1,
		Group_A: group & 1,
		Other_U: other & 4 >> 2,
		Other_M: other & 2 >> 1,
		Other_A: other & 1,
	}
}

func changePermissions(id int, p *Permissions, client *Client, call string) (string, error) {
  return client.Call(
    call,
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

