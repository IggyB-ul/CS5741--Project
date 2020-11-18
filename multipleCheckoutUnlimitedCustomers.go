// Basic producer/consumer implementation of checkouts (expanded from Chris Exton's example)
// Super basic customer generator implemented as a placeholder for the proper implementation
// Used google UUID package to give each customer a unique ID

// Already noticed one strange output due to concurrency issues.
// Although it might be OK - final version probably won't rely on printing 1000's of customer records to std out?
// Checkout:   2, Customer UUID: 621325f9-4dd0-4e30-b3df-b56171abf74a, Customer item count:    2, Total Customers Served:    47, Total items checked out:   4105, Customer Scan Time:     4, Customer Bag Time:     6
// Checkout:   1, Customer UUID: d161e638-7fc3-4942-b14d-8c1e053876c8, Customer item count:  161, Total Customers Served:    43, Total items checked out:   4432, Customer Scan Time:   322, Customer Bag Time:   483
// Checkout:   4, Customer UUID: 81222770-4be6-430e-93dc-737145e2d1b7, Customer item count:  191, Total Customers Served:    45, Total items checked out:   4344, Customer Scan Time:   191, Customer Bag Time:   286Checkout:   3, Custome
//r UUID: 7532f1e2-6ab5-4870-b893-a0146a742ebb, Customer item count:  170, Total Customers Served:    44, Total items checked out:   4418, Customer Scan Time:   340, Customer Bag Time:   510
//
//Checkout:   1, Customer UUID: 0fe650d8-250b-4fbf-a25a-4059277ea8b6, Customer item count:   34, Total Customers Served:    44, Total items checked out:   4466, Customer Scan Time:    68, Customer Bag Time:   102
package main

import (
	"fmt"
	//"strings"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type customer struct {
	uuid  uuid.UUID
	items int
}

var done = make(chan bool)
var queue = make(chan customer)

func createNewCustomer() *customer {
	p := customer{uuid: uuid.New()}
	p.items = rand.Intn(200) // random value between 0 and 200
	return &p
}
func superBasicCustomerSpawning() {
	for i := 0; i < 200; i++ {
		time.Sleep(10 * time.Millisecond)
		newCustomer := createNewCustomer()
		queue <- *newCustomer
	}
	done <- true
}

func checkout(checkoutNo int,
	itemLimit int,
	operatorMultiplier float64,
	scannerMultiplier float64,
	baggingTimeMultiplier float64,
	paymentTime int) {

	maxItemsLimitActive := false
	var totalProcessedCustomers int64 = 0
	var totalProcessedItems int64 = 0
	var totalScanningTime int64 = 0
	operatorMult := 1.0
	scannerMult := 1.0
	baggingTimeMult := 1.0
	defaultItemScanTime := 2.0

	if itemLimit > 0 {
		// Doesn't do much yet
		maxItemsLimitActive = true
	}
	if operatorMultiplier > 0 {
		// Adjust for faster/slower staff
		operatorMult = operatorMultiplier
	}
	if scannerMultiplier > 0 {
		// Adjust for faster/slower checkout hardware
		scannerMult = scannerMultiplier
	}
	if baggingTimeMultiplier > 0 {
		// simple relationship between total time taken to scan and time taken to bag everything
		baggingTimeMult = baggingTimeMultiplier
	}

	for {
		time.Sleep(1 * time.Millisecond)
		customer := <-queue // pull a customer off the  one common queue ( we need a better way!)
		totalProcessedCustomers++
		totalProcessedItems += int64(customer.items)
		var customerScanTime int64 = 0
		var itemSleepTime int64 = 0
		var bagSleepTimeNS int64 = 0
		var itemScanTime float64 = 1.0
		for i := 0; i < customer.items; i++ {
			itemScanTime = defaultItemScanTime * operatorMult * scannerMult
			// A bit of messing here, work in thousands of nano seconds rather than 1's of milliseconds,
			// due to the truncating of floats when type converting from float to int64
			itemSleepTime = int64(itemScanTime * 1000)
			time.Sleep(time.Duration(itemSleepTime) * time.Nanosecond)
			customerScanTime += (itemSleepTime / 1000)
		}
		totalScanningTime += customerScanTime
		bagSleepTimeNS = customerScanTime * int64(baggingTimeMult*1000)
		// A bit of messing here, work in thousands of nano seconds rather than 1's of milliseconds,
		// due to the truncating of floats when type converting from float to int64
		time.Sleep(time.Duration(bagSleepTimeNS) * time.Nanosecond)
		time.Sleep(time.Duration(paymentTime) * time.Millisecond)
		fmt.Printf("Checkout: %3d, Customer UUID: %s, Customer item count: %4d, ",
			checkoutNo,
			customer.uuid,
			customer.items)
		fmt.Printf("Total Customers Served: %5d, Total items checked out: %6d, ",
			totalProcessedCustomers,
			totalProcessedItems)
		fmt.Printf("Customer Scan Time: %5d, Customer Bag Time: %5d",
			customerScanTime,
			bagSleepTimeNS/1000)
		if maxItemsLimitActive {
			if customer.items > itemLimit {
				// not much use, just a placeholder
				fmt.Printf(", Max Items (%3d) EXCEEDED", itemLimit)
			}
		}
		fmt.Printf("\n")

	}
}

func main() {
	experiencedEmployeeScanMultiplier := 1.0
	newEmployeeScanMultiplier := 0.7
	standardScannerMultiplier := 1.0
	//fastScannerMultiplier := 2.0
	baggingTimeMultiplier := 1.5
	averagePaymentTime := 40

	rand.Seed(time.Now().UnixNano()) // seed random with a new value each time we run

	go superBasicCustomerSpawning()
	go checkout(1, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime)
	go checkout(2, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime)
	go checkout(3, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime)
	go checkout(4, 0, newEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime)
	<-done

}
