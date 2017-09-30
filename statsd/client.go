package statsd

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
	"strings"
)

const METRIC_TYPE_COUNT = 'c';
const METRIC_TYPE_GAUGE = 'g';
const METRIC_TYPE_TIMING = 't';
const METRIC_TYPE_SET = 's';

// The StatsdClient type defines the relevant properties of a StatsD connection.
type StatsdClient struct {
	host      string
	port      int
	conn      net.Conn
	rand      *rand.Rand        // rand generator to skip messages by sample rate
	keyBuffer map[string]string // array of messages to send
	autoflush bool // send metrics on every call
}

// Factory method to initialize udp connection
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125, false)
func New(host string, port int, autoflush bool) *StatsdClient {
	client := StatsdClient {
		host: host,
		port: port,
		rand: rand.New(rand.NewSource(time.Now().Unix())),
		keyBuffer: make(map[string]string),
	}
	client.Open()
	return &client
}

// Method to open udp connection, called by default client factory
func (client *StatsdClient) Open() {
	connectionString := fmt.Sprintf("%s:%d", client.host, client.port)
	conn, err := net.Dial("udp", connectionString)
	if err != nil {
		log.Println(err)
	}
	client.conn = conn
}

// Method to close udp connection
func (client *StatsdClient) Close() {
	client.conn.Close()
}

// Method to close udp connection
func (client *StatsdClient) SetAutoflush(autoflush bool) {
	client.autoflush = autoflush
}

// Log timing information (in milliseconds) with sampling
func (client *StatsdClient) Timing(key string, time int64, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", time, METRIC_TYPE_TIMING)
	if sampleRate < 1 {
		if (client.isSendAcceptedBySampleRate(sampleRate)) {
			metricValue = fmt.Sprintf("%s|@%f", metricValue, sampleRate)
		} else {
			return
		}
	}
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

// Arbitrarily updates a list of stats by a delta
func (client *StatsdClient) Count(key string, delta int, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", delta, METRIC_TYPE_COUNT)
	if sampleRate < 1 {
		if (client.isSendAcceptedBySampleRate(sampleRate)) {
			metricValue = fmt.Sprintf("%s|@%f", metricValue, sampleRate)
		} else {
			return
		}
	}
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

func (client *StatsdClient) Gauge(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, METRIC_TYPE_GAUGE)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

func (client *StatsdClient) Set(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, METRIC_TYPE_SET)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

// check if acceptable by sample rate
func (client *StatsdClient) isSendAcceptedBySampleRate(sampleRate float32) bool {
	if sampleRate >= 1 {
		return true
	}
	randomNumber := client.rand.Float32()
	return randomNumber <= sampleRate
}

// Sends data to udp statsd daemon
func (client *StatsdClient) Flush() {
	// prepare metric packet
	metricPacketArray := make([]string, len(client.keyBuffer))
	for key, metricValue := range client.keyBuffer {
		metricPacketArray = append(metricPacketArray, fmt.Sprintf("%s:%s", key, metricValue))
	}
	metricPacket := strings.Join(metricPacketArray, "|")

	// send metric packet
	_, err := fmt.Fprintf(client.conn, metricPacket)
	if err != nil {
		log.Println(err)
	}
}
