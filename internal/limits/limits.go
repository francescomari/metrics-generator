package limits

import (
	"fmt"
	"sync/atomic"
)

type Config struct {
	maxDuration      int64
	errorsPercentage int64
}

func (c *Config) MaxDuration() int {
	return int(atomic.LoadInt64(&c.maxDuration))
}

func (c *Config) SetMaxDuration(maxDuration int) error {
	if maxDuration < 0 {
		return fmt.Errorf("value is less than zero")
	}

	atomic.StoreInt64(&c.maxDuration, int64(maxDuration))

	return nil
}

func (c *Config) ErrorsPercentage() int {
	return int(atomic.LoadInt64(&c.errorsPercentage))
}

func (c *Config) SetErrorsPercentage(errorsPercentage int) error {
	if errorsPercentage < 0 || errorsPercentage > 100 {
		return fmt.Errorf("value is not a valid percentage")
	}

	atomic.StoreInt64(&c.errorsPercentage, int64(errorsPercentage))

	return nil
}
