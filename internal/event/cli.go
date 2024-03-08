package event

import "os"

type Cli struct {
	Uri string
}

func (c *Cli) Init(cfg string) error {
	// get the argument and set it to the Uri field
	if len(os.Args) > 1 {
		c.Uri = os.Args[len(os.Args)-1]
	}
	return nil
}

func (c *Cli) GetUri() (string, error) {
	return c.Uri, nil
}
