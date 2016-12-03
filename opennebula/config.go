package opennebula

import (
	"github.com/megamsys/opennebula-go/api"
	"log"
)

type Config struct {
	Endpoint string
	User     string
	Password string
}

// Client() returns a new client for accessing OpenNebula.
func (c *Config) Client() (*api.Rpc, error) {
	config := map[string]string{
		api.ENDPOINT: c.Endpoint,
		api.USERID:   c.User,
		api.PASSWORD: c.Password,
	}

	log.Printf("[INFO] OpenNebula Client configured for URL: %s", c.Endpoint)

	return api.NewClient(config)
}
