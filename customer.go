package main

import (
	"fmt"
	"math/rand"
	"time"
)

// This `person` struct type has `name` and `age` fields.
type customer struct {
	customerid       int
	items            int
	checkoutid       int
	queuetime        float32
	maxqueuetime     float32
	maxqueuecustomer int
	purchasecomplete bool
	leftqueue        bool
	checkouttime     float32
}

type weatherclass struct {
	noofcustomer, pausetime int
}

var done = make(chan bool)
var msgs = make(chan customer)
var weather = map[string]weatherclass{
	"quietday":   {30, 8},
	"averageday": {100, 5},
	"busyday":    {200, 1},
}
var timeofday = weather["quietday"]

// `newPerson` constructs a new person struct with the given name.
func createNewCustomer(customerid int) *customer {
	// create customer
	p := customer{customerid: customerid}
	p.items = rand.Intn(200)
	p.maxqueuecustomer = rand.Intn(10)
	p.maxqueuetime = rand.Float32() * 5
	return &p
}

func customerArrivalRate() {
	//simulate the rate of arrival based on the time of the day
	maxwait := timeofday.pausetime
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(maxwait*1000)))
}

func produceCustomer() {
	for i := 0; i < timeofday.noofcustomer; i++ {
		//set a pause between the arrivals of customer
		customerArrivalRate()
		newPerson := createNewCustomer(i)
		msgs <- *newPerson
	}
	done <- true // we are done producing
}

func consume(consumeNo int) {
	for {
		time.Sleep(1)
		msg := <-msgs
		fmt.Println("Customer #", consumeNo, msg) // just delay a little
	}
}
func main() {

	go produceCustomer()
	// I am going to have 2 consumers
	go consume(1)
	go consume(2)
	// just delay a little
	fmt.Println(timeofday.noofcustomer)
	<-done
}
