package statsd

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
	"strings"
)

const metric_type_count = 'c';
const metric_type_gauge = 'g';
const metric_type_timing = 't';
const metric_type_set = 's';

// The StatsdClient type
type StatsdClient struct {
	host      string
	port      int
	conn      net.Conn // UDP connection to StatsD server
	rand      *rand.Rand // rand generator to skip messages by sample rate
	keyBuffer map[string]string // array of messages to send
	autoflush bool // send metrics on every call
}

// Factory method to create new StatsD client
func New(host string, port int) *StatsdClient {
	client := StatsdClient {
		host: host,
		port: port,
		rand: rand.New(rand.NewSource(time.Now().Unix())),
		keyBuffer: make(map[string]string),
	}
	return &client
}

// Open UDP connection to statsd server
func (client *StatsdClient) Open() {
	connectionString := fmt.Sprintf("%s:%d", client.host, client.port)
	conn, err := net.Dial("udp", connectionString)
	if err != nil {
		log.Println(err)
	}
	client.conn = conn
	client.autoflush = false
}

// Close UDP connection to statsd server
func (client *StatsdClient) Close() {
	client.conn.Close()
}

// Enable/disable buffered mode
// In buffered mode requires manual call of Flush()
// In autoflush mode message sends to server on every call
func (client *StatsdClient) SetAutoflush(autoflush bool) {
	client.autoflush = autoflush
}

// Timing track in milliseconds with sampling
func (client *StatsdClient) Timing(key string, time int64, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", time, metric_type_timing)
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

// Count tack
func (client *StatsdClient) Count(key string, delta int, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", delta, metric_type_count)
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

// Gauge track
func (client *StatsdClient) Gauge(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metric_type_gauge)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

// Set tracking
func (client *StatsdClient) Set(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metric_type_set)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

// Check if acceptable by sample rate
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
