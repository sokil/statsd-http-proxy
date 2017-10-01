package main

import (
	"time"
	"net/http"
	"log"
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
	"github.com/gorilla/mux"
	"./statsd"
)

const DEFAULT_HTTP_HOST = "127.0.0.1"
const DEFAULT_HTTP_PORT = 80

const DEFAULT_STATSD_HOST = "127.0.0.1"
const DEFAULT_STATSD_PORT = 8125

const JWT_HEADER_NAME = "X-JWT-Token"

// declare command line options
var httpHost = flag.String("http-host", DEFAULT_HTTP_HOST, "HTTP Host")
var httpPort = flag.Int("http-port", DEFAULT_HTTP_PORT, "HTTP Port")
var statsdHost = flag.String("statsd-host", DEFAULT_STATSD_HOST, "StatsD Host")
var statsdPort = flag.Int("statsd-port", DEFAULT_STATSD_PORT, "StatsD Port")
var JWTSecret = flag.String("jwt-secret", "", "Secret to encrypt JWT")
var verbose = flag.Bool("verbose", false, "Verbose")

// statsd client
var statsdClient statsd.StatsdClient

func main() {
	// get flags
	flag.Parse()

	// configure verbosity of logging
	if *verbose == true {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// create HTTP router
	router := mux.NewRouter().StrictSlash(true)

	// get server address to bind
	httpAddress := fmt.Sprintf("%s:%d", *httpHost, *httpPort)
	log.Printf("Starting HTTP server %s", httpAddress)

	// create http server
	s := &http.Server {
		Addr: httpAddress,
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// register http request handlers
	router.Handle(
		"/heartbeat",
		http.HandlerFunc(handleHeartbeatRequest),
	).Methods("GET")

	router.Handle(
		"/count/{key}",
		validateJWT(http.HandlerFunc(handleCountRequest)),
	).Methods("POST")

	router.Handle(
		"/gauge/{key}",
		validateJWT(http.HandlerFunc(handleGaugeRequest)),
	).Methods("POST")

	router.Handle(
		"/timing/{key}",
		validateJWT(http.HandlerFunc(handleTimingRequest)),
	).Methods("POST")

	router.Handle(
		"/set/{key}",
		validateJWT(http.HandlerFunc(handleSetRequest)),
	).Methods("POST")

	// Create a new StatsD connection
	statsdClient = *statsd.New(*statsdHost, *statsdPort)
	statsdClient.SetAutoflush(true)
	statsdClient.Open()

	// start http server
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// validate JWT middleware
func validateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *JWTSecret == "" {
			next.ServeHTTP(w, r)
		} else {
			// get actual JWT
			JWT := r.Header.Get(JWT_HEADER_NAME)
			log.Printf("Validate token %s by sectet %s", JWT, *JWTSecret)

			// unpack JWT
			if (false) {
				http.Error(w, "Forbidden", 403)
			}

			// JWT expired
			if (false) {
				http.Error(w, "Token expired", 498)
			}
		}

	})
}

func handleHeartbeatRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func handleCountRequest(w http.ResponseWriter, r *http.Request) {
	// get key
	vars := mux.Vars(r)
	key := vars["key"]

	// get delta
	var delta int = 1;
	deltaPostFormValue := r.PostFormValue("delta")
	if deltaPostFormValue != "" {
		var err error
		delta, err = strconv.Atoi(deltaPostFormValue)
		if err != nil {
			http.Error(w, "Invalid delta specified", 400)
		}
	}

	// get sample rate
	var sampleRate float64 = 1
	sampleRatePostFormValue := r.PostFormValue("sampleRate")
	if sampleRatePostFormValue != "" {
		var err error
		sampleRate, err = strconv.ParseFloat(sampleRatePostFormValue, 32)
		if err != nil {
			http.Error(w, "Invalid sample rate specified", 400)
		}

	}

	// send request
	statsdClient.Count(key, delta, float32(sampleRate))
}

func handleGaugeRequest(w http.ResponseWriter, r *http.Request) {
	// get key
	vars := mux.Vars(r)
	key := vars["key"]

	// get delta
	var value int = 1;
	valuePostFormValue := r.PostFormValue("value")
	if valuePostFormValue != "" {
		var err error
		value, err = strconv.Atoi(valuePostFormValue)
		if err != nil {
			http.Error(w, "Invalid delta specified", 400)
		}
	}

	// send request
	statsdClient.Gauge(key, value)
}

func handleTimingRequest(w http.ResponseWriter, r *http.Request) {
	// get key
	vars := mux.Vars(r)
	key := vars["key"]

	// get timing
	time, err := strconv.ParseInt(r.PostFormValue("time"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid time specified", 400)
	}

	// get sample rate
	var sampleRate float64 = 1
	sampleRatePostFormValue := r.PostFormValue("sampleRate")
	if sampleRatePostFormValue != "" {
		var err error
		sampleRate, err = strconv.ParseFloat(sampleRatePostFormValue, 32)
		if err != nil {
			http.Error(w, "Invalid sample rate specified", 400)
		}
	}

	// send request
	statsdClient.Timing(key, time, float32(sampleRate))
}

func handleSetRequest(w http.ResponseWriter, r *http.Request) {
	// get key
	vars := mux.Vars(r)
	key := vars["key"]

	// get delta
	var value int = 1;
	valuePostFormValue := r.PostFormValue("value")
	if valuePostFormValue != "" {
		var err error
		value, err = strconv.Atoi(valuePostFormValue)
		if err != nil {
			http.Error(w, "Invalid delta specified", 400)
		}
	}

	// send request
	statsdClient.Set(key, value)
}
