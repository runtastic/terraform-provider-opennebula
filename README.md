# terraform-provider-opennebula

[OpenNebula](https://opennebula.org/) provider for [Terraform](https://www.terraform.io/).
 
* Leverages [OpenNebula's XML/RPC API](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html) 
* Tested for versions 5.X


The provider tries to impose a lightweight level of abstraction on OpenNebula's resources. This means that only the most fundamental attributes are directly accessible (i.e. names, IDs, permissions and user/group identities). For maximum flexibility and portability, the remaining attributes can be specified using any of the formats natively accepted by OpenNebula (XML and String).



## EXAMPLE

Create a file called `demo_template.txt`.

_Notice how we need to escape native variables `$$USER[SSH_PUBLIC_KEY]` with two dollar signs, as terraform will try to replace all variables with a single dollar sign)_

```
CUSTOM_ATTRIBUTE = "$CUSTOM_ATTRIBUTE_VALUE"
CONTEXT = [
  DNS_HOSTNAME = "yes",
  NETWORK = "YES",
  SSH_PUBLIC_KEY = "$$USER[SSH_PUBLIC_KEY]",
  USERNAME = "root" ]
CPU = "0.5"
VCPU = "4"
MEMORY = "3000"
GRAPHICS = [
  KEYMAP = "en",
  LISTEN = "0.0.0.0",
  TYPE = "VNC" ]
```

And the following `terraform.tf` file:

```
provider "opennebula" {
  endpoint = "api's endpoint"
  username = "user's name"
  password = "user's password"
}

data "template_file" "demo" {
  template = "${file("demo_template.txt")}"
  vars = {
    CUSTOM_ATTRIBUTE_VALUE = "demo-me"
  }
}

resource "opennebula_template" "demo" {
  name = "terraform-demo"
  description = "${data.template_file.demo.rendered}"
  permissions = "600"
}

output "demo_template_id" {
  value = "${opennebula_template.demo.id}"
}

output "demo_template_uname" {
  value = "${opennebula_template.demo.uname}"
}
```


## ROADMAP

The following list represent's all of OpenNebula's resources reachable through their API. The checked items are the ones that are fully functional and tested:

* [X] [onevm](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onevm)
* [X] [onetemplate](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onetemplate)
* [ ] [onehost](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onehost)
* [ ] [onecluster](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onecluster)
* [ ] [onegroup](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onegroup)
* [ ] [onevdc](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onevdc)
* [X] [onevnet](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onevnet)
* [ ] [oneuser](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#oneuser)
* [ ] [onedatastore](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onedatastore)
* [X] [oneimage](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#oneimage)
* [ ] [onemarket](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onemarket)
* [ ] [onemarketapp](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onemarketapp)
* [ ] [onevrouter](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onevrouter)
* [ ] [onezone](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onezone)
* [ ] [onesecgroup](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#onesecgroup)
* [ ] [oneacl](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#oneacl)
* [ ] [oneacct](https://docs.opennebula.org/5.2/integration/system_interfaces/api.html#oneacct)


## Collaborators

- [Lorenzo Arribas](https://github.com/larribas)
- [Jason Tevnan](https://github.com/tnosaj)

## Contributing
Bug reports and pull requests are welcome on GitHub at
https://github.com/runtastic/terraform-provider-opennebula. This project is
intended to be a safe, welcoming space for collaboration, and contributors are
expected to adhere to the
[Contributor Covenant](http://contributor-covenant.org) code of conduct.

## License

The gem is available as open source under the terms of
the [MIT License](http://opensource.org/licenses/MIT).
