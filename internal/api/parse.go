package api

import (
	"fmt"
	"strconv"
	"strings"
)

func parseDuration(value string) (int, int, error) {
	parts := strings.Split(value, ",")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("not a pair of number")
	}

	min, err := parseInt(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("minimum is not a number")
	}

	max, err := parseInt(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("maximum is not a number")
	}

	return min, max, nil
}

func parseInt(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(string(value)))
	if err != nil {
		return 0, fmt.Errorf("not a number")
	}

	return parsed, nil
}
