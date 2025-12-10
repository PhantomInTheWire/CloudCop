package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Worker service running...")
	for {
		fmt.Println("Health: ok (worker)")
		time.Sleep(30 * time.Second)
	}
}
