package command

import (
	"fmt"
	"github.com/devlibx/gox-base/errors"
	"github.com/devlibx/gox-base/util"
)

func (c *Config) SetupDefaults() {

	// Fill defaults in servers
	if c.Servers != nil {
		for k, v := range c.Servers {
			v.Name = k
			if v.ConnectTimeout <= 0 {
				v.ConnectTimeout = 50
			}
			if v.ConnectionRequestTimeout <= 0 {
				v.ConnectionRequestTimeout = 50
			}
			if v.Port == 0 {
				v.Port = 80
			}
			if util.IsStringEmpty(v.Host) {
				v.Host = "localhost"
			}
		}
	}

	// Fill defaults in apis
	if c.Apis != nil {
		for k, v := range c.Apis {
			v.Name = k
			if v.Timeout <= 0 {
				v.Timeout = 1
			}
			if v.Concurrency <= 0 {
				v.Concurrency = 1
			}
			if v.QueueSize <= 0 {
				v.QueueSize = 1
			}
			if util.IsStringEmpty(v.Method) {
				v.Method = "GET"
			}
		}
	}
}

func (c *Config) FindServerByName(toFind string) (*Server, error) {
	for name, server := range c.Servers {
		if name == toFind {
			return server, nil
		}
	}
	return nil, errors.New("server not found with %s name", toFind)
}

func (c *Config) FindApiByName(toFind string) (*Api, error) {
	for name, api := range c.Apis {
		if name == toFind {
			return api, nil
		}
	}
	return nil, errors.New("api not found with %s name", toFind)
}

func (a *Api) GetPath(server *Server) string {
	if server.Https {
		return fmt.Sprintf("https://%s:%d%s", server.Host, server.Port, a.Path)
	} else {
		return fmt.Sprintf("http://%s:%d%s", server.Host, server.Port, a.Path)
	}
}
