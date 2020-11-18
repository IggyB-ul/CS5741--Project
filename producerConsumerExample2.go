// The producer-consumer pattern with a single producer and a single consumer is one of the most simple patterns
// in parallel computing. In Go, you can implement it like this:
package main

import (
	"fmt"
	"strings"
)

var users = map[int]string{
	1: "Rob",
	2: "Ken",
	3: "Robert",
}

func Search(substring string) <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		for _, name := range users {
			if ok := strings.Contains(name, substring); !ok {
				continue
			}
			c <- name
		}
	}()
	return c
}
func main() {
	for name := range Search("Rob") {
		fmt.Println(name)
	}
}
