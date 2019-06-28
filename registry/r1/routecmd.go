package r1

import (
	"encoding/json"
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

	discoveryEnv string

	env map[string]string
}

type ServiceMetadata struct {
	App string
	Environment string
	Label string
	BasePath string
	Scheme string
}

func (r routecmd) build() []string {
	// generate route commands
	var config []string
	var glob = false
	var discoveryKey, env, scheme, basePath string

	discoveryKey = r.svc.ServiceName

	for key, value := range r.svc.ServiceMeta {
		switch key {
		case "Environment":
			env = value
		case "Scheme":
			scheme = value
		case "BasePath":
			basePath = value
		}
	}

	for _, t := range r.svc.ServiceTags {
		if len(env) < 1 || len(scheme) < 1 {
			meta := ServiceMetadata {}

			err := json.Unmarshal([]byte(t), &meta)
			if err == nil {
				env = meta.Environment
				scheme = meta.Scheme
				basePath = meta.BasePath
			}
		}
	}

	if len(env) < 1 || len(scheme) < 1 {
		return config
	}

	if env == r.discoveryEnv {
		glob = true
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

	var ropts []string

	ropts = append(ropts, "strip=" + src)

	ropts = append(ropts, "proto=" + scheme)

	cfg := "route add " + name + " " + src + " " + dst

	if len(ropts) > 0 {
		cfg += " opts " + strconv.Quote(strings.Join(ropts, " "))
	}

	config = append(config, cfg)

	return config
}