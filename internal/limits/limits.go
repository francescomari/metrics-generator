package limits

import (
	"fmt"
	"sync"
)

type Config struct {
	mu               sync.RWMutex
	minDuration      int
	maxDuration      int
	errorsPercentage int
}

func (c *Config) DurationInterval() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.minDuration, c.maxDuration
}

func (c *Config) SetDurationInterval(minDuration, maxDuration int) error {
	if minDuration <= 0 {
		return fmt.Errorf("minimum duration is less than or equal to zero")
	}
	if maxDuration <= 0 {
		return fmt.Errorf("maximum duration is less than or equal to zero")
	}
	if maxDuration < minDuration {
		return fmt.Errorf("maximum duration is less then or equal to minimum duration")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.minDuration = minDuration
	c.maxDuration = maxDuration

	return nil
}

func (c *Config) ErrorsPercentage() int {
	return c.errorsPercentage
}

func (c *Config) SetErrorsPercentage(errorsPercentage int) error {
	if errorsPercentage < 0 || errorsPercentage > 100 {
		return fmt.Errorf("value is not a valid percentage")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.errorsPercentage = errorsPercentage

	return nil
}
