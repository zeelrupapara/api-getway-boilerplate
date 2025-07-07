package utils

import (
	"fmt"
	"time"
)

// use it before closing a function to calculate the time took to excute this function
func Timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start).Milliseconds())
	}
}
