package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
)

var (
	delayThreshold     = flag.Int64("deadline", 2000, "Number of milliseconds before a request is considered delayed.")
	maxActiveRequests  = flag.Int("active", 10, "Maximum number of requests active at one time.")
	maxDelayedRequests = flag.Int("delayed", 20, "Quit after this many Delayed requests.")
	maxTotalRequests   = flag.Int("total", 1000, "Quit after this many Total requets.")
	stripeKey          = flag.String("key", "", "MANDATORY ARGUMENT: Stripe API key to use while testing")
	debug              = flag.Bool("D", false, "Enable debug (primarily for development)")
	activeRequests     int
	delayedRequests    int
	totalRequests      int
	hostname           string
	err                error
	wg                 sync.WaitGroup
)

func pdebug(frmt string, vars ...interface{}) {
	if !*debug {
		return
	}
	fmt.Printf("DEBUG: %s\n", fmt.Sprintf(frmt, vars...))
}

func main() {
	doInit()
	var loopCount uint32
	for {
		if delayedRequests > *maxDelayedRequests || totalRequests > *maxTotalRequests {
			fmt.Printf("Something is outside tolerance: delayed = %d [%d]\ttotal = %d [%d]\n", delayedRequests, *maxDelayedRequests, totalRequests, *maxTotalRequests)
			goto AFTERFORLOOP
		}
		// Put in a circuit breaker here to keep us from having too many requests in flight at once
		// accounting for the edge case where active requests + delayed requests > max delayed requests so
		// we don't shoot past the max delayed request threshold.
		loopCount = 0
		for activeRequests >= *maxActiveRequests || activeRequests+delayedRequests >= *maxDelayedRequests {
			// Make sure we don't get caught in an infinite loop here if we got too many delayed requests
			if delayedRequests >= *maxDelayedRequests {
				fmt.Printf("delayed requests [%d] >= max dealyed requests [%d] so exiting\n", delayedRequests, *maxDelayedRequests)
				goto AFTERFORLOOP
			}
			// And in case both of those fail, this will hopefully keep us from getting stuck in the mud
			loopCount++
			if loopCount > 4294967000 {
				fmt.Printf("Bouncing out of loop with loopCount of %d\n", loopCount)
				goto AFTERFORLOOP
			}
		}
		// This is only used for reporting, so we're fine if there's a little racey-ness
		activeRequests++
		wg.Add(1)
		go createStripeTestCustomer(totalRequests)
		totalRequests++
	}
AFTERFORLOOP:
	wg.Wait()
	fmt.Printf("Final results: delayed = %d [%d]\ttotal = %d [%d]\n", delayedRequests, *maxDelayedRequests, totalRequests, *maxTotalRequests)
	fmt.Printf("All requests completed\n")
}

func usage(msg ...string) {
	fmt.Printf("ERROR: %s\n", msg[0])
	fmt.Printf("Usage: %s -key <Stripe Api Key> [optional arguments]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func doInit() {
	flag.Parse()
	if *stripeKey == "" {
		usage("Did not provide a Stripe API Key.")
	}
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	stripe.EnableTelemetry = true
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{
		Level: stripe.LevelWarn,
	}

	if *debug {
		showRunParameters()
	} else {
		pdebug("I guess debug wasn't enabled [%t]\n", *debug)
	}
}

func showRunParameters() {
	pdebug("### Stripe Customer Creation Tester ###")
	pdebug("    Using the following configuration:")
	pdebug(" Debug: %t", *debug)
	pdebug(" Delay Threshold: %dms", *delayThreshold)
	pdebug(" Max Active Requests: %d", *maxActiveRequests)
	pdebug(" Max Delayed Requests: %d", *maxDelayedRequests)
	pdebug(" Max Total Requests: %d", *maxTotalRequests)
	pdebug(" Stripe API Key is %d chars long with an md5sum of %x", len(*stripeKey), md5.Sum([]byte(*stripeKey)))
}

func createStripeTestCustomer(reqNum int) {
	defer wg.Done()
	params := &stripe.CustomerParams{}
	sc := &client.API{}
	sc.Init(*stripeKey, nil)

	now := time.Now()
	desc := fmt.Sprintf("Stripe Connectivity Test %d from %s at %d ms since epoch", reqNum, hostname, now.UnixMilli())
	params.AddExtra("description", desc)
	start := time.Now()
	cust, err := sc.Customers.New(params)

	// Boilerplate error handling from Stripe Docs
	if err != nil {
		// Try to safely cast a generic error to a stripe.Error so that we can get at
		// some additional Stripe-specific information about what went wrong.
		if stripeErr, ok := err.(*stripe.Error); ok {
			// The Code field will contain a basic identifier for the failure.
			switch stripeErr.Code {
			case stripe.ErrorCodeCardDeclined:
			case stripe.ErrorCodeExpiredCard:
			case stripe.ErrorCodeIncorrectCVC:
			case stripe.ErrorCodeIncorrectZip:
				// etc.
			}

			// The Err field can be coerced to a more specific error type with a type
			// assertion. This technique can be used to get more specialized
			// information for certain errors.
			if cardErr, ok := stripeErr.Err.(*stripe.CardError); ok {
				fmt.Printf("Card was declined with code: %v\n", cardErr.DeclineCode)
			} else {
				fmt.Printf("Other Stripe error occurred: %v\n", stripeErr.Error())
			}
		} else {
			fmt.Printf("Other error occurred: %v\n", err.Error())
		}

		// log.Fatal(err)
	}
	done := time.Now()
	activeRequests--

	duration := done.UnixMilli() - start.UnixMilli()

	fmt.Printf("### Request %d - %dms\n", reqNum, duration)
	if duration > *delayThreshold {
		delayedRequests++
		log.Printf("Took %dms {> %dms} to create customer %s [%d]\n", duration, *delayThreshold, cust.ID, reqNum)
	}
}
