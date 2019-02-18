package v2vvmware

import (
	"context"
	"fmt"
)

/*
  Following code is based on https://github.com/pkliczewski/provider-pod
  modified for the needs of the controller-flow.
*/

func getClient(ctx context.Context, loginCredentials *LoginCredentials) (*Client, error) {
	c, err := NewClient(ctx, loginCredentials)
	if err != nil {
		log.Error(err, "GetVMs: failed to create a client.")
		return nil, err
	}
	return c, nil
}

func GetVMs(c *Client) ([]string, error) {
	vms, err := c.GetVMs()
	if err != nil {
		log.Error(err, "GetVMs: failed to get list of VMs from VMWare.")
		return nil, err
	}

	names := make([]string, len(vms))
	for i, vm := range vms {
		names[i] = vm.Summary.Config.Name
	}

	log.Info(fmt.Sprintf("GetVMs: retrieved list of virtual machines: %s", names))
	return names, nil
}
