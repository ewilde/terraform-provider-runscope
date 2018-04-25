package runscope

import (
	"log"

	"github.com/ewilde/go-runscope"
)

// Config contains runscope provider settings
type Config struct {
	AccessToken string
	APIURL      string
}

func (c *Config) client() (*runscope.Client, error) {
	client := runscope.NewClient(c.APIURL, c.AccessToken)

	log.Printf("[INFO] runscope client configured for server %s", c.APIURL)

	return client, nil
}
