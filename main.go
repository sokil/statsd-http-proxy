package main

import (
	"./statsd"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const defaultHTTPHost = "127.0.0.1"
const defaultHTTPPort = 80

const defaultStatsDHost = "127.0.0.1"
const defaultStatsDPort = 8125

const jwtHeaderName = "X-JWT-Token"

// declare command line options
var httpHost = flag.String("http-host", defaultHTTPHost, "HTTP Host")
var httpPort = flag.Int("http-port", defaultHTTPPort, "HTTP Port")
var statsdHost = flag.String("statsd-host", defaultStatsDHost, "StatsD Host")
var statsdPort = flag.Int("statsd-port", defaultStatsDPort, "StatsD Port")
var tokenSecret = flag.String("jwt-secret", "", "Secret to encrypt JWT")
var verbose = flag.Bool("verbose", false, "Verbose")

// statsd client
var statsdClient *statsd.Client

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
	s := &http.Server{
		Addr:           httpAddress,
		Handler:        router,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// register http request handlers
	router.Handle(
		"/heartbeat",
		validateCORS(http.HandlerFunc(handleHeartbeatRequest)),
	).Methods("GET")

	router.Handle(
		"/count/{key}",
		validateCORS(validateJWT(http.HandlerFunc(handleCountRequest))),
	).Methods("POST")

	router.Handle(
		"/gauge/{key}",
		validateCORS(validateJWT(http.HandlerFunc(handleGaugeRequest))),
	).Methods("POST")

	router.Handle(
		"/timing/{key}",
		validateCORS(validateJWT(http.HandlerFunc(handleTimingRequest))),
	).Methods("POST")

	router.Handle(
		"/set/{key}",
		validateCORS(validateJWT(http.HandlerFunc(handleSetRequest))),
	).Methods("POST")

	// Create a new StatsD connection
	statsdClient = statsd.NewClient(*statsdHost, *statsdPort)
	statsdClient.SetAutoflush(true)
	statsdClient.Open()

	// start http server
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// validate CORS headers
func validateCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With, Origin, Accept, Content-Type, Authentication")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
			w.Header().Add("Access-Control-Allow-Origin", origin)
			w.Header().Add("Access-Control-Expose-Headers", "X-Sentry-Error, Retry-After")
		}
		next.ServeHTTP(w, r)
	})
}

// validate JWT middleware
func validateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *tokenSecret == "" {
			next.ServeHTTP(w, r)
		} else {
			// get JWT
			tokenString := r.Header.Get(jwtHeaderName)
			if tokenString == "" {
				http.Error(w, "Token not specified", 401)
				return
			}

			// parse JWT
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(*tokenSecret), nil
			})

			if err != nil {
				http.Error(w, "Error parsing token", 403)
				return
			}

			// accept request
			next.ServeHTTP(w, r)
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
	var delta = 1
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
	var value = 1
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
	var value = 1
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
