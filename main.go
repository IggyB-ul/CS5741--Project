package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//TODO:
// - Make output as much information as possible
// - Comment code more so that others can understand (only when the full code is finished)

type store struct {
	storeId                          int
	checkouts                        map[string]*checkout
	busyRanges                       map[string]busyRange
	weather                          optionFactor
	openingHours                     string
	totalCustomers                   int
	customers                        map[string]*customer
	processedCustomers               SafeCounter
	notProcessedCustomersQueuingTime SafeCounter
	notProcessedCustomersQueuingDeep SafeCounter
	hasFloorManager                  bool
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
	paymentTime          int
	checkoutDesirability int
	currentDeep          SafeCounter
	status               string
	totalCustomersServed SafeCounter
	totalItemsCheckedOut SafeCounter
}

func (c *checkout) scanProduct(customer *customer, product *product) {
	// Scan according to product scan time and cashier efficiency
	// Bag according to scan time scaled by an average customer bagging factor
	_, simWorldCurrentTimeString := dualClock.getSimWorldCurrentTime()
	fmt.Printf("%s:Checkout%2d: SCANNING -> Customer: %3d, Product: %4d | SimScanTime;%5.2f;\n",
		simWorldCurrentTimeString, c.checkoutId, customer.customerId, product.productId, product.processTimeSecond*c.cashierEfficiency)
	gClock.scaleSleepTimeForSimulation(product.processTimeSecond * c.cashierEfficiency)
	//fmt.Printf("Checkout%2d;BAGGING ;Prod.id;%3d;SimBagTime;%3.2f;\n",
	//	c.checkoutId,product.productId, product.processTimeSecond * 1.2)
	gClock.scaleSleepTimeForSimulation(product.processTimeSecond * 1.2) // DOR Careful now - fix this magic number later
	// Really we need to add more concurrency here for scanning/ bagging
	c.totalItemsCheckedOut.Inc()
}

type customer struct {
	customerId          int
	items               int
	checkoutId          int
	queueTimeStart      int64
	queueTimeEnd        int64
	queueTimeSeconds    int64
	maxQueueTimeSeconds int64
	maxQueueCustomers   int
	purchaseComplete    bool
	leftQueue           bool
	checkoutTime        int64
	checkoutTimeStart   int64
	checkoutTimeEnd     int64
	products            map[string]product
}

type clock struct {
	secondsAreOneHour int
}

func (c clock) convertFromSeconds(seconds int) float64 {
	return float64(seconds) / 60 / 60 * float64(c.secondsAreOneHour)
}

func (c clock) scaleSleepTimeForSimulation(seconds float64) {
	// Work in seconds usually for easier human understanding,
	// then call this function for any sleep times in the simulated world
	// secondsAreOneHour REAL WORLD seconds == 60 * 60 == 3600 Simulated seconds
	// so a sleep for 9 seconds in the simulation corresponds to
	// 9 * secondsAreOneHour/3600 in the Real world.
	//
	// As we need to allow for simulating hundredths of a second, we will also need to scale up
	// the value to avoid truncation when converting to int, and compensate for that by using
	// microseconds or nanoseconds
	// So.. 5.55 seconds in the simulation. secondsAreOneHour = 10
	// 5.55 seconds = 5550000 microseconds in simulation = 5550000/3600 microseconds in real world
	// 1541.666 recurring microseconds
	// I think we are safe to truncate that to 1541 microseconds
	timeToSleepScaledUpFloat := seconds * 1000000 * float64(c.secondsAreOneHour) / 3600
	timeToScanScaledUpInt := int(timeToSleepScaledUpFloat)
	timeToSleepInRealWorld := time.Duration(timeToScanScaledUpInt) * time.Microsecond
	time.Sleep(timeToSleepInRealWorld)
}

type dualTimeClock struct {
	secondsAreOneHour    int
	realWorldStartTime   int64
	realWorldCurrentTime int64
	simWorldDayNumber    int
	simWorldStartTime    int64
	simWorldCurrentTime  int64
}

func (dtc *dualTimeClock) initRealWorldStartTime() {
	dtc.realWorldStartTime = time.Now().UnixNano() // number of nanoseconds since January 1, 1970 UTC
}
func (dtc *dualTimeClock) initSimWorldDayClock(storeOpenTimeHoursInt int) {
	dtc.simWorldStartTime = 0 // it's groundhog day!
	openingTimeInSeconds := int64(60 * 60 * storeOpenTimeHoursInt)
	dtc.simWorldStartTime = openingTimeInSeconds // it's opening time!
}
func (dtc *dualTimeClock) getRealWorldCurrentTime() int64 {
	dtc.realWorldCurrentTime = time.Now().UnixNano()
	return dtc.realWorldCurrentTime
}
func (dtc *dualTimeClock) getSimWorldCurrentTime() (int64, string) {
	// So we know that secondsAreOneHour seconds in the real world == 3600 seconds in sim world
	// Let's assume the minimum value for secondsAreOneHour == 1
	// secondsAreOneHour milliseconds in the real world == 3.6 seconds in sim world
	// secondsAreOneHour microseconds in the real world == 0.0036 seconds in sim world
	// Surely ^that's^ adequate resolution?
	// secondsAreOneHour nanoseconds in the real world == 0.0000036 seconds in sim world
	// 1 microsecond in the real world = (0.0036/secondsAreOneHour) seconds in sim world
	var secondsSinceOpening float64
	var secondsSinceStoreEpoch int64
	//var humanReadableTimeString string
	realWorldMicroSecondsElapsed := (time.Now().UnixNano() - dtc.realWorldStartTime) / 1000
	scalingRealTimeToSimTime := 0.0036 / float64(dtc.secondsAreOneHour)
	secondsSinceOpening = float64(realWorldMicroSecondsElapsed) * scalingRealTimeToSimTime
	secondsSinceStoreEpoch = dtc.simWorldStartTime + int64(secondsSinceOpening)
	unixTime := time.Unix(secondsSinceStoreEpoch, 0)
	humanReadableTimeString := fmt.Sprintf("%02d:%02d:%02d", unixTime.Hour(), unixTime.Minute(), unixTime.Second())
	return secondsSinceStoreEpoch, humanReadableTimeString
}

func (dtc *dualTimeClock) diffInSeconds(start int64, end int64) int64 {
	realWorldMicroSecondsElapsed := end - start
	return realWorldMicroSecondsElapsed
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
	processTimeSecond float64
}

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	mu sync.Mutex
	v  int
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc() {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.v = c.v + 1
	c.mu.Unlock()
}

// Dec increments the counter for the given key.
func (c *SafeCounter) Dec() {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.v = c.v - 1
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value() int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.v
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
	// moved seed to a one time only position in main
	// Reseeding every time was 'resetting the clock' on the randomness
	//rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func getBusyFactor(store *store) float64 {
	_, currentTime := dualClock.getSimWorldCurrentTime()
	//Example: 08 becomes 8
	hour, _ := strconv.Atoi(currentTime[0:2])

	return float64(store.busyRanges["busyRange_"+strconv.Itoa(hour)].busyOptionFactor.factor)
}

func openCheckout(store *store, checkoutName string, checkout *checkout) {
	fmt.Println("Opening: " + checkoutName)

	for {
		//Time between one payment and next person
		gClock.scaleSleepTimeForSimulation(float64(30))
		queueIndex := getQueueIndex(store, checkout)
		customer := <-queues[queueIndex]
		customer.queueTimeEnd, _ = dualClock.getSimWorldCurrentTime()

		if customer.queueTimeStart != customer.queueTimeEnd {
			customer.queueTimeSeconds = dualClock.diffInSeconds(customer.queueTimeStart, customer.queueTimeEnd)
		}

		if customer.queueTimeSeconds > customer.maxQueueTimeSeconds {
			customer.leftQueue = true
			store.notProcessedCustomersQueuingTime.Inc()
			continue
		}

		if checkout.currentDeep.Value() > customer.maxQueueCustomers {
			customer.leftQueue = true
			store.notProcessedCustomersQueuingDeep.Inc()
			continue
		}

		customer.checkoutId = checkout.checkoutId
		customer.checkoutTimeStart, _ = dualClock.getSimWorldCurrentTime()
		checkout.status = "BUSY"
		_, simWorldCurrentTimeString := dualClock.getSimWorldCurrentTime()
		fmt.Printf("%s:Customer %4d arrived at Checkout %2d with %3d items\n",
			simWorldCurrentTimeString, customer.customerId, checkout.checkoutId, customer.items)
		fmt.Println("Customer " + strconv.Itoa(customer.customerId) + ", Arrived at checkout: " + strconv.Itoa(checkout.checkoutId))
		for _, eProduct := range customer.products {
			checkout.scanProduct(customer, &eProduct)
		}
		_, simWorldCurrentTimeString = dualClock.getSimWorldCurrentTime()
		fmt.Printf("%s:Customer %4d is paying at Checkout %2d...\n",
			simWorldCurrentTimeString, customer.customerId, checkout.checkoutId)
		gClock.scaleSleepTimeForSimulation(float64(checkout.paymentTime))
		_, simWorldCurrentTimeString = dualClock.getSimWorldCurrentTime()
		fmt.Printf("%s:Customer %4d is finished at Checkout %2d.\n",
			simWorldCurrentTimeString, customer.customerId, checkout.checkoutId)

		customer.purchaseComplete = true
		checkout.status = "IDLE"
		checkout.totalCustomersServed.Inc()
		checkout.currentDeep.Dec()
		store.processedCustomers.Inc()
		customer.checkoutTimeEnd, _ = dualClock.getSimWorldCurrentTime()
		customer.checkoutTime = dualClock.diffInSeconds(customer.checkoutTimeStart, customer.checkoutTimeEnd)
		c.Inc()
		ch <- 1
	}

}

var c = SafeCounter{v: 0}
var queues = make(map[string]chan *customer)
var gClock clock
var dualClock dualTimeClock

func getCheckoutWithShorterQueue(store *store, nextCustomerNumberOfProducts int) *checkout {

	lowestDeep := -1
	var selectedCheckout string

	for kCheckout := range store.checkouts {
		tmpCheckout := store.checkouts[kCheckout]

		if lowestDeep < 0 && (nextCustomerNumberOfProducts <= tmpCheckout.maxItems || tmpCheckout.maxItems == 0) {
			lowestDeep = tmpCheckout.currentDeep.Value()
			selectedCheckout = kCheckout
		}

		if tmpCheckout.currentDeep.Value() < lowestDeep && (nextCustomerNumberOfProducts <= tmpCheckout.maxItems || tmpCheckout.maxItems == 0) {
			lowestDeep = tmpCheckout.currentDeep.Value()
			selectedCheckout = kCheckout
		}

	}

	return store.checkouts[selectedCheckout]
}

func getCheckoutRandomly(store *store, nextCustomerNumberOfProducts int) *checkout {

	tmpCheckouts := make(map[int]string)

	i := 0

	for kCheckout := range store.checkouts {
		tmpCheckout := store.checkouts[kCheckout]

		if nextCustomerNumberOfProducts <= tmpCheckout.maxItems || tmpCheckout.maxItems == 0 {
			tmpCheckouts[i] = kCheckout
		}
	}

	rangeEnds := len(tmpCheckouts)

	return store.checkouts[tmpCheckouts[generateRandomNumber(1, rangeEnds)]]
}

func customerSpawning(eStore *store) {

	i := 0

	for kCustomer := range eStore.customers {
		//People will arrive every 5 minutes normally.
		invertedFactor := getBusyFactor(eStore) - 2
		if invertedFactor < 0 {
			invertedFactor = 1
		}
		gClock.scaleSleepTimeForSimulation(300 * invertedFactor)

		nextCustomerNumberOfProducts := len(eStore.customers[kCustomer].products)

		var checkout *checkout
		if eStore.hasFloorManager {
			checkout = getCheckoutWithShorterQueue(eStore, nextCustomerNumberOfProducts)
		} else {
			checkout = getCheckoutRandomly(eStore, nextCustomerNumberOfProducts)
		}

		checkout.currentDeep.Inc()
		queueIndex := getQueueIndex(eStore, checkout)

		fmt.Println("Queue: " + queueIndex + " has length: " + strconv.Itoa(checkout.currentDeep.Value()))
		eStore.customers[kCustomer].queueTimeStart, _ = dualClock.getSimWorldCurrentTime()
		queues[queueIndex] <- eStore.customers[kCustomer]
		i++
	}
}

func getQueueIndex(eStore *store, eCheckout *checkout) string {
	return "store_" + strconv.Itoa(eStore.storeId) + "_checkout_" + strconv.Itoa(eCheckout.checkoutId)
}

//var wg sync.WaitGroup
var ch chan int

func main() {

	var lastStringReader string
	var stores = map[string]*store{}
	rand.Seed(time.Now().UnixNano())
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
		"1",
		useDefaultSettings)
	oneHourIsInSeconds, _ := strconv.Atoi(lastStringReader)

	gClock = clock{secondsAreOneHour: oneHourIsInSeconds}
	dualClock = dualTimeClock{secondsAreOneHour: oneHourIsInSeconds}

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
			"[Store "+strconv.Itoa(iStore)+"] How many customers do you want to generate? Range response [300-500] "+
				"means from 300 to 500 customers a day.",
			true,
			"500-1000",
			useDefaultSettings)
		//// number of products
		numberOfProducts := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many products do you want to generate per customer? Range "+
				"response [1-150] means from 1 to 150 products per customer.",
			true,
			"1-150",
			useDefaultSettings)
		//// number of products
		productProcessTime := readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How much should it take a product to be scanned? Range response in "+
				"seconds [0.5-10] means from 0.5 second to 10 seconds per product.",
			true,
			"0.5-10",
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
			"100-150",
			useDefaultSettings)

		//// number of checkouts
		lastStringReader = readFromConsole(
			"[Store "+strconv.Itoa(iStore)+"] How many checkouts will this store have? [10] ",
			true,
			"10",
			useDefaultSettings)

		numberOfCheckouts, _ := strconv.Atoi(lastStringReader)

		var checkouts = map[string]*checkout{}

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

			checkouts["checkout"+strconv.Itoa(iCheckout)] = &checkout{
				checkoutId:           iCheckout,
				cashierEfficiency:    cashierEfficiency,
				paymentTime:          60,
				maxItems:             maxItems,
				checkoutDesirability: checkoutDesirability,
				currentDeep:          SafeCounter{v: 0},
				status:               "IDLE",
				totalItemsCheckedOut: SafeCounter{v: 0},
				totalCustomersServed: SafeCounter{v: 0},
			}
		}

		numberOfCustomersParts := strings.Split(numberOfCustomers, "-")
		numberOfCustomersFrom, _ := strconv.Atoi(numberOfCustomersParts[0])
		numberOfCustomersTo, _ := strconv.Atoi(numberOfCustomersParts[1])

		var customers = map[string]*customer{}

		//Apply weather conditions:
		numberOfCustomersTo = int(float32(numberOfCustomersTo) * weather.factor)

		for iCustomer := numberOfCustomersFrom; iCustomer < numberOfCustomersTo; iCustomer++ {

			numberOfProductsParts := strings.Split(numberOfProducts, "-")
			numberOfProductsFrom, _ := strconv.Atoi(numberOfProductsParts[0])
			numberOfProductsTo, _ := strconv.Atoi(numberOfProductsParts[1])

			var products = map[string]product{}
			numberOfProductsForCustomer := generateRandomNumber(numberOfProductsFrom, numberOfProductsTo)
			for iProduct := 1; iProduct <= numberOfProductsForCustomer; iProduct++ {

				productProcessTimeParts := strings.Split(productProcessTime, "-")
				productProcessTimeFrom, _ := strconv.ParseFloat(productProcessTimeParts[0], 64)
				productProcessTimeTo, _ := strconv.ParseFloat(productProcessTimeParts[1], 64)

				// we gave the user the example/default of 0.5 - 10s
				// for practicality, let's only deal with tenths of second for scanning times
				// rand only deals with ints so we need to multiply by 10, then convert to an int
				// then divide by 10 to get tenths of a second in a sensible range for
				// scanning groceries
				processTimeCalc := float64(generateRandomNumber(
					int(10*productProcessTimeFrom), int(10*productProcessTimeTo)))
				processTimeCalc = processTimeCalc / 10.0
				products["product"+strconv.Itoa(iProduct)] = product{
					productId:         iProduct,
					processTimeSecond: processTimeCalc,
				}
			}

			var maxQueueTimeSeconds int64

			maxQueueTimeParts := strings.Split(maxQueueTime, "-")
			maxQueueTimeFrom, _ := strconv.Atoi(maxQueueTimeParts[0])
			maxQueueTimeTo, _ := strconv.Atoi(maxQueueTimeParts[1])

			maxQueueTimeSeconds = int64(generateRandomNumber(maxQueueTimeFrom, maxQueueTimeTo) * 60)

			maxQueueCustomersParts := strings.Split(maxQueueCustomers, "-")
			maxQueueCustomersFrom, _ := strconv.Atoi(maxQueueCustomersParts[0])
			maxQueueCustomersTo, _ := strconv.Atoi(maxQueueCustomersParts[1])

			customers["customer"+strconv.Itoa(iCustomer)] = &customer{
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

		stores["store"+strconv.Itoa(iStore)] = &store{
			storeId:            iStore,
			checkouts:          checkouts,
			busyRanges:         busyRanges,
			weather:            weather,
			openingHours:       openingHours,
			totalCustomers:     generateRandomNumber(numberOfCustomersFrom, numberOfCustomersTo),
			hasFloorManager:    isFloorManager,
			customers:          customers,
			processedCustomers: SafeCounter{v: 0},
		}
	}
	dualClock.initRealWorldStartTime()
	fmt.Println("START Sanity Check of Dual Clock")
	dualClock.initSimWorldDayClock(8)
	realWorldStartTimeTempString := time.Now()
	realWorldStartTimeTempSeconds := time.Now().Unix()
	simWorldStartTimeTempInt, simWorldStartTimeTempString := dualClock.getSimWorldCurrentTime()
	fmt.Printf("Realworld start time is %s.\n", realWorldStartTimeTempString)
	fmt.Printf("Simworld start time is %s.\n", simWorldStartTimeTempString)
	fmt.Println("Sleep 10 seconds")
	time.Sleep(1 * time.Second)
	fmt.Printf("Realworld elapsed time is %d.\n", time.Now().Unix()-realWorldStartTimeTempSeconds)
	simWorldCurrentTimeInt, simWorldCurrentTimeString := dualClock.getSimWorldCurrentTime()
	fmt.Printf("Simworld elapsed time is %d.\n", simWorldCurrentTimeInt-simWorldStartTimeTempInt)
	fmt.Printf("Simworld current time is %s.\n", simWorldCurrentTimeString)
	fmt.Println("END Sanity Check of Dual Clock")

	totalCheckouts := 0
	allCustomerToBeProcessed := 0
	ch = make(chan int)

	for kStore, eStore := range stores {
		fmt.Println(kStore)

		allCustomerToBeProcessed = allCustomerToBeProcessed + len(eStore.customers)
		for kCheckout, eCheckout := range eStore.checkouts {
			fmt.Println(kCheckout)
			index := getQueueIndex(stores[kStore], eCheckout)
			queues[index] = make(chan *customer)
			//wg.Add(1)
			go openCheckout(eStore, kCheckout, eCheckout)
			totalCheckouts++
		}

	}

	for _, eStore := range stores {
		go customerSpawning(eStore)
	}

	totalProcessedCustomers := SafeCounter{v: 0}

	for {
		newCustomerProcessed := <-ch

		if newCustomerProcessed > 0 {
			totalProcessedCustomers.Inc()
		}

		if totalProcessedCustomers.Value() == allCustomerToBeProcessed {
			close(ch)
			break
		}
	}

	for kStore := range stores {
		eStore := stores[kStore]

		fmt.Println("---Store: " + kStore + ", Customer processed: " + strconv.Itoa(eStore.processedCustomers.Value()))

		for kCheckout := range eStore.checkouts {
			out := eStore.checkouts[kCheckout]

			fmt.Println("---Checkout: " + kCheckout + ", Customers processed: " + strconv.Itoa(
				out.totalCustomersServed.Value()) + ", Products processed: " + strconv.Itoa(
				out.totalItemsCheckedOut.Value()))
		}
	}
	//<-done
}
