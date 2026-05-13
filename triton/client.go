/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2021 Joyent, Inc.
 * Copyright 2022 MNX Cloud, Inc.
 * Copyright 2026 Edgecast Cloud LLC.
 */

package triton

import (
	"context"
	"fmt"
	"sync"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
	"github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang/typed"
)

// Client represents the Triton CloudAPI client and the configuration
// necessary to make authenticated requests.
type Client struct {
	api     *cloudapi.ClientWithResponses
	typed   *typed.Client
	account string
	url     string

	// affinityLock serializes CreateMachine calls that carry affinity
	// rules.  Terraform creates resources in parallel, but affinity
	// constraints are only visible to DAPI after a machine is placed —
	// without serialization two concurrent creates could both be
	// scheduled before either placement is recorded.
	affinityLock *sync.RWMutex

	cnsOnce    sync.Once
	cnsEnabled bool
	cnsErr     error
}

// API returns the underlying CloudAPI client.
func (c *Client) API() *cloudapi.ClientWithResponses { return c.api }

// Typed returns a wrapper client for CloudAPI's action-dispatch endpoints
// (StartMachine, StopMachine, RenameMachine, ResizeMachine, firewall/
// deletion-protection toggles, UpdateVolume, ResizeDisk, ...).
func (c *Client) Typed() *typed.Client { return c.typed }

// Account returns the Triton account name used for API calls.
func (c *Client) Account() string { return c.account }

// URL returns the Triton CloudAPI endpoint URL.
func (c *Client) URL() string { return c.url }

// CNSEnabled returns whether the Triton Container Name Service is
// enabled for this account.  The result is fetched once from CloudAPI
// and cached for the lifetime of the client.
func (c *Client) CNSEnabled() (bool, error) {
	c.cnsOnce.Do(func() {
		resp, err := c.api.GetAccountWithResponse(
			context.Background(), c.account)
		if err != nil {
			c.cnsErr = fmt.Errorf("error checking CNS status: %s", err)
			return
		}
		if resp.JSON200 == nil {
			c.cnsErr = fmt.Errorf("error checking CNS status: %s",
				formatAPIError(resp.StatusCode(), resp.Body))
			return
		}
		c.cnsEnabled = derefBool(resp.JSON200.TritonCnsEnabled)
	})
	return c.cnsEnabled, c.cnsErr
}
