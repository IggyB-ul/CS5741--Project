package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

//TODO:
// - Implement the busy hours.
// - Implement weather
// - Implement time of the day so that it can be printed what is the "Real time".
// - Implement all counters, as total items processed, etc.
// - Make output as much information as possible
// - Comment code more so that others can understand (only when the full code is finished)
// - Implement customer giving up because of time
// - Implement customer giving up because of current deep.

type store struct {
	storeId            int
	checkouts          map[string]checkout
	busyRanges         map[string]busyRange
	weather            optionFactor
	openingHours       string
	totalCustomers     int
	customers          map[string]customer
	processedCustomers int
	hasFloorManager    bool
}

type busyRange struct {
	fromHour         int
	toHour           int
	busyOptionFactor optionFactor
}

type checkout struct {
	checkoutId           int
	cashierEfficiency    float64
	maxItems             int
	checkoutDesirability int
	currentDeep          int
	status               string
	totalCustomersServed int
	totalItemsCheckedOut int
}

func (c checkout) scanProduct(product product) {
	//productProcessTime := gClock.convertFromSeconds(product.processTimeSecond)
	// timeToProcess := time.Duration(productProcessTime*c.cashierEfficiency) * time.Second
	timeToProcess := time.Duration(1 * time.Second)
	time.Sleep(timeToProcess)
	fmt.Println("Checkout" + "Scanning: " + strconv.Itoa(product.productId))
}

type customer struct {
	customerId          int
	items               int
	checkoutId          int
	queueTimeSeconds    int
	maxQueueTimeSeconds int
	maxQueueCustomers   int
	purchaseComplete    bool
	leftQueue           bool
	checkoutTime        float32
	products            map[string]product
}

type clock struct {
	secondsAreOneHour int
}

func (c clock) convertFromSeconds(seconds int) float64 {
	return float64(seconds) / 60 / 60 * float64(c.secondsAreOneHour)
}

type optionFactor struct {
	name   string
	factor float32
}

var weatherOptions = map[string]optionFactor{
	"B": {"Bad", 0.8},
	"G": {"Good", 1},
	"E": {"Excellent", 0.85},
}

var busyRangeOptions = map[string]optionFactor{
	"Q":  {"Quiet", 0.8},
	"LB": {"Little-busy", 1},
	"B":  {"Busy", 1.2},
}

type product struct {
	productId         int
	processTimeSecond int
}

func readFromConsole(label string, convertToUpper bool, defaultValue string, useDefaultSettings bool) string {

	fmt.Print(label + "\n")

	if useDefaultSettings {
		fmt.Print(defaultValue + "\n")
		return defaultValue
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if convertToUpper {
		text = strings.ToUpper(text)
	}

	if text == "" {
		text = defaultValue
	}

	fmt.Print(text + "\n")
	return text
}

func generateRandomNumber(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func openCheckout(store store, checkoutName string, checkout checkout) {

	fmt.Println("Opening: " + checkoutName)

	for {
		queueIndex := getQueueIndex(store, checkout)
		customer := <-queues[queueIndex]
		fmt.Println("Customer: " + strconv.Itoa(customer.customerId) + ", Arrived at checkout: " + strconv.Itoa(checkout.checkoutId))
		for _, eProduct := range customer.products {
			checkout.scanProduct(eProduct)
		}
	}

}

var queues = make(map[string]chan customer)

//
//var done = make(chan bool)
//var done map[string]chan bool

//var queue = make(chan customer)
var gClock clock

func customerSpawning(eStore store) {

	i := 0

	for _, eCustomer := range eStore.customers {
		rangeEnds := len(eStore.checkouts) - 1
		queueIndex := getQueueIndex(eStore, eStore.checkouts["checkout"+strconv.Itoa(generateRandomNumber(1, rangeEnds))])
		if eStore.hasFloorManager {
			queues[queueIndex] <- eCustomer
		} else {
			queues[queueIndex] <- eCustomer
		}
		i++
	}

	// Close queues after finishing.
	if len(eStore.customers)-1 == i {
		for _, eCheckout := range eStore.checkouts {
			queueIndex := getQueueIndex(eStore, eCheckout)
			close(queues[queueIndex])
		}
	}

	//done <- true
}

func getQueueIndex(eStore store, eCheckout checkout) string {
	return "store_" + strconv.Itoa(eStore.storeId) + "_checkout_" + strconv.Itoa(eCheckout.checkoutId)
}

func main() {

	var lastStringReader string
	var stores = map[string]store{}

	lastStringReader = readFromConsole(
		"Do you want to use all defaults settings? [Y/n]:",
		true,
		"Y",
		false)

	useDefaultSettings := false
	if lastStringReader == "Y" {
		useDefaultSettings = true
	}

	////Value of One hour In seconds
	lastStringReader = readFromConsole(
		"How many seconds in the simulation will be one hour in real life? [1] means: 1 second is 1 hour in real life.",
		true,
		"8-20",
		useDefaultSettings)
	oneHourIsInSeconds, _ := strconv.Atoi(lastStringReader)

	gClock = clock{secondsAreOneHour: oneHourIsInSeconds}

	if gClock.secondsAreOneHour > 60 {
		fmt.Println("Warning simulation may be slow..")
	}

	//// Number of stores
	lastStringReader = readFromConsole(
		"How many stores do you want to simulate?",
		true,
		"1",
		useDefaultSettings)

	numberOfStores, _ := strconv.Atoi(lastStringReader)

	//// Define settings by each store
	for iStore := 1; iStore <= numberOfStores; iStore++ {

		//// Opening Hours
		openingHours := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] Enter opening hours from-to, [8-20]:",
			true,
			"8-20",
			useDefaultSettings)
		//// busy ranges, ask based on opening times.
		openingHoursParts := strings.Split(openingHours, "-")
		openingHoursFrom, _ := strconv.Atoi(openingHoursParts[0])
		openingHoursTo, _ := strconv.Atoi(openingHoursParts[1])

		var busyRanges = map[string]busyRange{}

		for iBusyRange := openingHoursFrom; iBusyRange <= openingHoursTo; iBusyRange++ {
			lastStringReader := readFromConsole(
				"[Store "+strconv.Itoa(iStore)+"] How busy will this store be at: ["+strconv.Itoa(iBusyRange)+":00]",
				true,
				"lb",
				useDefaultSettings)
			selectedBusyRange := busyRangeOptions[lastStringReader]

			busyRanges["busyRange_"+strconv.Itoa(iBusyRange)] = busyRange{
				fromHour:         iBusyRange,
				toHour:           iBusyRange + 1,
				busyOptionFactor: selectedBusyRange,
			}
		}

		//// Weather
		lastStringReader := readFromConsole(
			"Set weather conditions: type: B or G or E. Where B means bad, G means good and E means excellent:",
			true,
			"G",
			useDefaultSettings)

		weather := weatherOptions[lastStringReader]
		//// Floor manager
		lastStringReader = readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] Do you want to enable a Floor Manager for this store? [Y/n]:",
			true,
			"Y",
			useDefaultSettings)
		isFloorManager := false
		if lastStringReader == "Y" {
			isFloorManager = true
		}
		//// number of customers
		numberOfCustomers := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many customers do you want to generate? Range response [100-200] "+
				"means from 100 to 200 customers a day.",
			true,
			"100-200",
			useDefaultSettings)
		//// number of products
		numberOfProducts := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many products do you want to generate per customer? Range "+
				"response [1-50] means from 1 to 50 customers a day.",
			true,
			"1-50",
			useDefaultSettings)
		//// number of products
		productProcessTime := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How much should it take a product to be scanned? Range response in "+
				"seconds [1-30] means from 1 second to 30 seconds per product.",
			true,
			"1-30",
			useDefaultSettings)

		//// max queue time
		maxQueueTime := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many minutes will usually a customer be in queue before giving up? "+
				"Range response in minutes [15-30] means from 15 to 30 minute a person will usually give up",
			true,
			"15-30",
			useDefaultSettings)

		//// max queue customers
		maxQueueCustomers := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How deep should usually a queue be for customer to give up? "+
				"Range response in customer numbers [10-15] means from 10 to 15 customers in queue will make a customer "+
				"to give up.",
			true,
			"10-15",
			useDefaultSettings)

		//// number of checkouts
		lastStringReader = readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many checkouts will this store have? [10] ",
			true,
			"10",
			useDefaultSettings)

		numberOfCheckouts, _ := strconv.Atoi(lastStringReader)

		var checkouts = map[string]checkout{}

		//// Define settings by each checkout
		for iCheckout := 1; iCheckout <= numberOfCheckouts; iCheckout++ {
			//// Cashier Efficiency
			lastStringReader = readFromConsole(
				"[Store "+strconv.Itoa(iStore)+"][Checkout "+strconv.Itoa(iCheckout)+"] How efficient is this cashier? [1] Recommended value from 0.1 (Really Slow) to 1.9 (Really Fast) ",
				true,
				"1",
				useDefaultSettings)

			cashierEfficiency, _ := strconv.ParseFloat(lastStringReader, 64)
			//// Max Items
			lastStringReader = readFromConsole(
				"[Store "+strconv.Itoa(iStore)+"][Checkout "+strconv.Itoa(iCheckout)+"] Maximum items for this checkout? 0 means unlimited [0] ",
				true,
				"0",
				useDefaultSettings)

			maxItems, _ := strconv.Atoi(lastStringReader)
			//// Checkout desirability
			lastStringReader = readFromConsole(
				"[Store "+strconv.Itoa(iStore)+"][Checkout "+strconv.Itoa(iCheckout)+"] How desirable will be this checkout in respect to the others "+
					"based on its location? ",
				true,
				strconv.Itoa(iCheckout),
				useDefaultSettings)

			checkoutDesirability, _ := strconv.Atoi(lastStringReader)

			checkouts["checkout"+strconv.Itoa(iCheckout)] = checkout{
				checkoutId:           iCheckout,
				cashierEfficiency:    cashierEfficiency,
				maxItems:             maxItems,
				checkoutDesirability: checkoutDesirability,
				currentDeep:          0,
				status:               "IDLE",
				totalCustomersServed: 0,
				totalItemsCheckedOut: 0,
			}
		}

		numberOfCustomersParts := strings.Split(numberOfCustomers, "-")
		numberOfCustomersFrom, _ := strconv.Atoi(numberOfCustomersParts[0])
		numberOfCustomersTo, _ := strconv.Atoi(numberOfCustomersParts[1])

		var customers = map[string]customer{}

		for iCustomer := numberOfCustomersFrom; iCustomer <= numberOfCustomersTo; iCustomer++ {

			numberOfProductsParts := strings.Split(numberOfProducts, "-")
			numberOfProductsFrom, _ := strconv.Atoi(numberOfProductsParts[0])
			numberOfProductsTo, _ := strconv.Atoi(numberOfProductsParts[1])

			var products = map[string]product{}

			for iProduct := numberOfProductsFrom; iProduct <= numberOfProductsTo; iProduct++ {

				productProcessTimeParts := strings.Split(productProcessTime, "-")
				productProcessTimeFrom, _ := strconv.Atoi(productProcessTimeParts[0])
				productProcessTimeTo, _ := strconv.Atoi(productProcessTimeParts[1])

				products["product"+strconv.Itoa(iProduct)] = product{
					productId:         iProduct,
					processTimeSecond: generateRandomNumber(productProcessTimeFrom, productProcessTimeTo),
				}
			}

			var maxQueueTimeSeconds int

			maxQueueTimeParts := strings.Split(maxQueueTime, "-")
			maxQueueTimeFrom, _ := strconv.Atoi(maxQueueTimeParts[0])
			maxQueueTimeTo, _ := strconv.Atoi(maxQueueTimeParts[1])

			maxQueueTimeSeconds = generateRandomNumber(maxQueueTimeFrom, maxQueueTimeTo) * 60

			maxQueueCustomersParts := strings.Split(maxQueueCustomers, "-")
			maxQueueCustomersFrom, _ := strconv.Atoi(maxQueueCustomersParts[0])
			maxQueueCustomersTo, _ := strconv.Atoi(maxQueueCustomersParts[1])

			customers["customer"+strconv.Itoa(iCustomer)] = customer{
				customerId:          iCustomer,
				items:               len(products),
				checkoutId:          0,
				queueTimeSeconds:    0,
				maxQueueTimeSeconds: maxQueueTimeSeconds,
				maxQueueCustomers:   generateRandomNumber(maxQueueCustomersFrom, maxQueueCustomersTo),
				purchaseComplete:    false,
				leftQueue:           false,
				checkoutTime:        0,
				products:            products,
			}
		}

		stores["store"+strconv.Itoa(iStore)] = store{
			storeId:            iStore,
			checkouts:          checkouts,
			busyRanges:         busyRanges,
			weather:            weather,
			openingHours:       openingHours,
			totalCustomers:     generateRandomNumber(numberOfCustomersFrom, numberOfCustomersTo),
			processedCustomers: 0,
			hasFloorManager:    isFloorManager,
			customers:          customers,
		}
	}

	for kStore, eStore := range stores {
		fmt.Println(kStore)

		for kCheckout, eCheckout := range eStore.checkouts {
			fmt.Println(kCheckout)
			index := getQueueIndex(eStore, eCheckout)
			queues[index] = make(chan customer)
			go openCheckout(eStore, kCheckout, eCheckout)
		}

		customerSpawning(eStore)
	}

	//<-done
}
