package r1

import (
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
)

// routecmd builds a route command.
type routecmd struct {
	// svc is the consul service instance.
	svc *api.CatalogService

	tag string

	discoveryEnv string

	env map[string]string
}

func (r routecmd) build() []string {
	// generate route commands
	var config, svctags []string
	var discoverable, glob = false, false
	var discoveryKey, env, baseEnv, scheme, basePath string

	for _, t := range r.svc.ServiceTags {
		if t == r.tag {
			discoverable = true
		} else {
			svctags = append(svctags, t)
		}
	}

	if !discoverable {
		return config
	}

	for key, value := range r.svc.ServiceMeta {
		switch key {
		case "DiscoveryKey":
			discoveryKey = value
		case "Env":
			env = value
		case "BaseEnv":
			baseEnv = value
		case "Scheme":
			scheme = value
		case "BasePath":
			basePath = value
		}
	}

	if len(baseEnv) < 1 {
		baseEnv = env
		glob = true
	}

	if baseEnv != r.discoveryEnv {
		return config
	}

	name, addr, port := r.svc.ServiceName, r.svc.ServiceAddress, r.svc.ServicePort

	// use consul node address if service address is not set
	if addr == "" {
		addr = r.svc.Address
	}

	// add .local suffix on OSX for simple host names w/o domain
	if runtime.GOOS == "darwin" && !strings.Contains(addr, ".") && !strings.HasSuffix(addr, ".local") {
		addr += ".local"
	}

	addr = net.JoinHostPort(addr, strconv.Itoa(port))

	e := env
	if glob {
		e = "[^/]+"
	}

	src:= "/" + e + "/" + discoveryKey
	dst := scheme + "://" + addr + "/"
	if len(basePath) > 0 {
		dst = dst + basePath + "/"
	}

	var weight string
	var ropts []string

	ropts = append(ropts, "strip=" + src)

	ropts = append(ropts, "proto=" + scheme)

	cfg := "route add " + name + " " + src + " " + dst
	if weight != "" {
		cfg += " weight " + weight
	}
	if len(svctags) > 0 {
		cfg += " tags " + strconv.Quote(strings.Join(svctags, ","))
	}
	if len(ropts) > 0 {
		cfg += " opts " + strconv.Quote(strings.Join(ropts, " "))
	}

	config = append(config, cfg)

	return config
}