package triton

import (
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

// retryOnError uses resource.Retry from Terraform core to retry a function when
// specific Triton errors are thrown. The first argument is a function from
// `triton-go` which checks the error returned by the function of the second
// argument. Error functions can be found in `triton-go`.
func retryOnError(isRetry func(err error) bool, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			if isRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	return resp, err
}
