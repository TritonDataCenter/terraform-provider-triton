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
	"sync"

	cloudapi "github.com/TritonDataCenter/monitor-reef/clients/external/cloudapi-client/golang"
)

// Client represents the Triton CloudAPI client and the configuration
// necessary to make authenticated requests.
type Client struct {
	api          *cloudapi.ClientWithResponses
	account      string
	url          string
	affinityLock *sync.RWMutex
}

// API returns the underlying CloudAPI client.
func (c *Client) API() *cloudapi.ClientWithResponses { return c.api }

// Account returns the Triton account name used for API calls.
func (c *Client) Account() string { return c.account }

// URL returns the Triton CloudAPI endpoint URL.
func (c *Client) URL() string { return c.url }
