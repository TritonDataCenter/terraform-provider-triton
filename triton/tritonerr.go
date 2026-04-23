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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// retryOnError uses resource.Retry from Terraform core to retry a function when
// specific errors are thrown. The first argument is a predicate that determines
// whether the error is retryable.
func retryOnError(isRetry func(err error) bool, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := retry.Retry(2*time.Minute, func() *retry.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			if isRetry(err) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	return resp, err
}
