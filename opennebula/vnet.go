package opennebula

import (
	"strconv"
	"strings"
)

//Adress Ranges
type Ars struct {
  Ar_id                string `xml:"AR_ID"`
  Global_prefix        string `xml:"GLOBAL_PREFIX"`
  Ip                   string `xml:"IP"`
  Mac                  string `xml:"MAC"`
  Parent_network_ar_id string `xml:"PARENT_NETWORK_AR_ID"`
  Size                 int `xml:"SIZE"`
  Type                 string `xml:"TYPE"`
  Ula_prefix           string `xml:"ULA_PREFIX"`
  Vn_mad               string `xml:"VN_MAD"`
  Mac_end              string `xml:"MAC_END"`
  Ip_end               string `xml:"IP_END"`
  Ip6_ula              string `xml:"IP6_ULA"`
  Ip6_ula_end          string `xml:"IP6_ULA_END"`
  Ip6_global           string `xml:"IP6_GLOBAL"`
  Ip6_global_end       string `xml:"IP6_GLOBAL_END"`
  Used_leases          string `xml:"USED_LEASES"`
  // Array of leases
  Leases               [1]Leases `xml:"LEASES"`
}

type Leases struct {
  Ip                   string `xml:"IP"`
  Ip6_global           string `xml:"IP6_GLOBAL"`
  Ip6_link             string `xml:"IP6_LINK"`
  Ip6_ula              string `xml:"IP6_ULA"`
  Mac                  string `xml:"MAC"`
  Vm                   string `xml:"VM"`
  Vnet                 string `xml:"VNET"`
  Vrouter              string `xml:"VROUTER"`
}

