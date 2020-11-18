// The producer-consumer pattern with a single producer and a single consumer is one of the most simple patterns
// in parallel computing. In Go, you can implement it like this:
package main

/* producer-consumer problem in Go */

import (
	"fmt"
	"math/rand"
	"time"
)

// This `person` struct type has `name` and `age` fields.
type person struct {
	name string
	age  int
}

var done = make(chan bool)
var msgs = make(chan person)

// `newPerson` constructs a new person struct with the given name.
func createNewPerson(name string) *person {
	// You can safely return a pointer to local variable
	// as a local variable will survive the scope of the function.
	p := person{name: name}
	p.age = rand.Intn(100)
	return &p
}
func produce() {
	for i := 0; i < 10; i++ {
		// It's idiomatic to encapsulate new struct creation in constructor functions
		newPerson := createNewPerson("Jon")
		msgs <- *newPerson
	}
	done <- true // we are done producing
}
func consume(consumeNo int) {
	for {
		time.Sleep(1)
		msg := <-msgs
		fmt.Println("Consumer #", consumeNo, msg) // just delay a little
	}
}
func main() {
	go produce()
	// I am going to have 2 consumers
	go consume(1)
	go consume(2)
	<-done
}
