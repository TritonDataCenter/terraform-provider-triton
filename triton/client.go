package triton

import (
	"sync"

	"github.com/hashicorp/errwrap"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/account"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/identity"
	"github.com/joyent/triton-go/network"
)

// Client represents all internally accessible Triton APIs utilized by this
// provider and the configuration necessary to connect to them.
type Client struct {
	config                *triton.ClientConfig
	insecureSkipTLSVerify bool
	affinityLock          *sync.RWMutex
}

func (c Client) Account() (*account.AccountClient, error) {
	accountClient, err := account.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Account Client: {{err}}", err)
	}

	if c.insecureSkipTLSVerify {
		accountClient.Client.InsecureSkipTLSVerify()
	}
	return accountClient, nil
}

func (c Client) Compute() (*compute.ComputeClient, error) {
	computeClient, err := compute.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Compute Client: {{err}}", err)
	}
	if c.insecureSkipTLSVerify {
		computeClient.Client.InsecureSkipTLSVerify()
	}
	return computeClient, nil
}

func (c Client) Identity() (*identity.IdentityClient, error) {
	identityClient, err := identity.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Identity Client: {{err}}", err)
	}
	if c.insecureSkipTLSVerify {
		identityClient.Client.InsecureSkipTLSVerify()
	}
	return identityClient, nil
}

func (c Client) Network() (*network.NetworkClient, error) {
	networkClient, err := network.NewClient(c.config)
	if err != nil {
		return nil, errwrap.Wrapf("Error Creating Triton Network Client: {{err}}", err)
	}
	if c.insecureSkipTLSVerify {
		networkClient.Client.InsecureSkipTLSVerify()
	}
	return networkClient, nil
}
