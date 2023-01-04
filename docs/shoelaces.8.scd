shoelaces(8)

# NAME

shoelaces - automated server bootstrapping

# SYNOPSIS

*shoelaces* [options...]

# OPTIONS

*-base-url* <string>
	Optional parameter. Specifies the base address that will be used when
	generating URLs.
	If it's not specified, the value of "-bind-addr" will be used.

*-bind-addr* <host:port>
	The address where Shoelaces will listen for requests. Defaults to
	"localhost:8081".

*-config* <config>
	Specifies a config file. All the following options can be specified in
	the config.

*-data-dir* <directory>
	Specifies a directory with mappings, configs, templates, etc.

*-debug*
	Enables debug mode.

*-env-dir* <directory>
	Specifies a directory with environment overrides. Refer to the README of
	the project for more information about environment overrides.

*-mappings-file* <file>
	Specifies a mappings YAML file. Defaults to "mappings.yaml". Refer to the
	README of the project for more information about mappings.

*-static-dir* <directory>
	Specifies a custom web directory with static files. Defaults to "web".

*-template-extension* <extension>
	Shoelaces template extension. Defaults to ".slc".

# DESCRIPTION

Shoelaces serves over HTTP iPXE boot scripts, cloud-init configuration, and
any other configuration files to baremetal or virtual machines booting
using iPXE. It also does a few other things to make easier managing server
deployments:

- Has a simple but nice UI to show the current configuration, and history
  of servers that booted.
- Uses Go templates to handle complex configurations.
- Allows specifying the boot entry point for a given server based on its IP
  address or DNS PTR record.
- Supports the notion of environments for Development and Production
  environment configurations, while trying to minimize template
  duplication.
- Puts unknown servers into iPXE script boot retry loop, while at the same
  time showing them in the UI allowing the user to select a specific boot
  configuration.

Shoelaces is better when used along a DHCP server. Refer to the
*CONFIGURATION* section for examples on how to configure a DHCP server for
hooking it with Shoelaces.

# CONFIGURATION

Shoelaces searches for a config file whenever the *-config* flag is
specified.

Here is example config file:

```
bind-addr=localhost:8081
data-dir=/etc/shoelaces/data-dir/
template-extension=.slc
mappings-file=mappings.yaml
debug=true
```

*dhcpd*(8) can be hooked with Shoelaces by adding the following snippet in
*dhcpd.conf*(5):

```
# dhcp.conf
next-server <your-tftp-server>;
if exists user-class and option user-class = "iPXE" {
  filename "http://<shoelaces-server>/start";
} else {
  filename "undionly.kpxe";
}
```

Similarly, *dnsmasq*(8) can be configured for working with Shoelaces with
the following snippet:

```
dhcp-match=set:ipxe,175 # iPXE sends a 175 option.
dhcp-boot=tag:!ipxe,undionly.kpxe
dhcp-boot=http://<shoelaces-server>/start
```

A TFTP server such as *tftpd*(8) must be configured to serve the iPXE executable,
*undionly.kpxe*.

# SEE ALSO

*dhcpd*(8) *dhcpd.conf*(5) *dnsmasq*(8) *tftpd*(8)

# AUTHORS

Maintained by ThousandEyes Inc. <opensource@thousandeyes.com>.

Up-to-date sources can be found at https://github.com/thousandeyes/shoelaces.

Bugs reports and patches can be submitted via GitHub's interface or
contacting the maintainers directly via email.
