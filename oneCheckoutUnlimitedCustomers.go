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
	p.items = rand.Intn(200)
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
		maxItemsLimitActive = true
	}
	if operatorMultiplier > 0 {
		operatorMult = operatorMultiplier
	}
	if scannerMultiplier > 0 {
		scannerMult = scannerMultiplier
	}
	if baggingTimeMultiplier > 0 {
		baggingTimeMult = baggingTimeMultiplier
	}

	for {
		time.Sleep(1 * time.Millisecond)
		customer := <-queue
		totalProcessedCustomers++
		totalProcessedItems += int64(customer.items)
		var customerScanTime int64 = 0
		var itemSleepTime int64 = 0
		var bagSleepTimeNS int64 = 0
		var itemScanTime float64 = 1.0
		for i := 0; i < customer.items; i++ {
			itemScanTime = defaultItemScanTime * operatorMult * scannerMult
			itemSleepTime = int64(itemScanTime * 1000)
			time.Sleep(time.Duration(itemSleepTime) * time.Nanosecond)
			customerScanTime += (itemSleepTime / 1000)
		}
		totalScanningTime += customerScanTime
		bagSleepTimeNS = customerScanTime * int64(baggingTimeMult*1000)
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
				fmt.Printf(", Max Items (%3d) EXCEEDED", itemLimit)
			}
		}
		fmt.Printf("\n")

	}
}

func main() {
	experiencedEmployeeScanMultiplier := 1.0
	//newEmployeeScanMultiplier := 0.7
	standardScannerMultiplier := 1.0
	//fastScannerMultiplier := 2.0
	baggingTimeMultiplier := 1.5
	averagePaymentTime := 40

	rand.Seed(time.Now().UnixNano())

	go superBasicCustomerSpawning()
	//func checkout (checkoutNo int,
	//	itemLimit int,
	//	operatorMultiplier float32,
	//	scannerMultiplier float32,
	//	baggingTimeMultiplier float32,
	//	paymentTime int)
	go checkout(1, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime)
	//go checkout(2, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime )
	//go checkout(3, 0, experiencedEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime )
	//go checkout(4, 0, newEmployeeScanMultiplier, standardScannerMultiplier, baggingTimeMultiplier, averagePaymentTime )
	<-done

}
