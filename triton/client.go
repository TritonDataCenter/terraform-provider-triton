package triton

import (
	"fmt"
	"sync"

	triton "github.com/TritonDataCenter/triton-go"
	"github.com/TritonDataCenter/triton-go/account"
	"github.com/TritonDataCenter/triton-go/compute"
	"github.com/TritonDataCenter/triton-go/identity"
	"github.com/TritonDataCenter/triton-go/network"
	"github.com/TritonDataCenter/triton-go/services"
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
		return nil, fmt.Errorf("Error Creating Triton Account Client: %s", err)
	}

	if c.insecureSkipTLSVerify {
		accountClient.Client.InsecureSkipTLSVerify()
	}
	return accountClient, nil
}

func (c Client) Compute() (*compute.ComputeClient, error) {
	computeClient, err := compute.NewClient(c.config)
	if err != nil {
		return nil, fmt.Errorf("Error Creating Triton Compute Client: %s", err)
	}
	if c.insecureSkipTLSVerify {
		computeClient.Client.InsecureSkipTLSVerify()
	}
	return computeClient, nil
}

func (c Client) Identity() (*identity.IdentityClient, error) {
	identityClient, err := identity.NewClient(c.config)
	if err != nil {
		return nil, fmt.Errorf("Error Creating Triton Identity Client: %s", err)
	}
	if c.insecureSkipTLSVerify {
		identityClient.Client.InsecureSkipTLSVerify()
	}
	return identityClient, nil
}

func (c Client) Network() (*network.NetworkClient, error) {
	networkClient, err := network.NewClient(c.config)
	if err != nil {
		return nil, fmt.Errorf("Error Creating Triton Network Client: %s", err)
	}
	if c.insecureSkipTLSVerify {
		networkClient.Client.InsecureSkipTLSVerify()
	}
	return networkClient, nil
}

func (c Client) Services() (*services.ServiceGroupClient, error) {
	servicesClient, err := services.NewClient(c.config)
	if err != nil {
		return nil, fmt.Errorf("Error Creating Triton Services Client: %s", err)
	}
	if c.insecureSkipTLSVerify {
		servicesClient.Client.InsecureSkipTLSVerify()
	}
	return servicesClient, nil
}
