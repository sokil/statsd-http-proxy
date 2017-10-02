package statsd

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

const metricTypeCount = "c"
const metricTypeGauge = "g"
const metricTypeTiming = "t"
const metricTypeSet = "s"

// The Client type
type Client struct {
	host          string
	port          int
	conn          net.Conn          // UDP connection to StatsD server
	rand          *rand.Rand        // rand generator to skip messages by sample rate
	keyBuffer     map[string]string // array of messages to send
	keyBufferLock sync.RWMutex      // mutex to lock buffer of keys
	autoflush     bool              // send metrics on every call
}

// NewClient creates new StatsD client
func NewClient(host string, port int) *Client {
	client := Client{
		host:      host,
		port:      port,
		rand:      rand.New(rand.NewSource(time.Now().Unix())),
		keyBuffer: make(map[string]string),
		autoflush: false,
	}
	return &client
}

// Open UDP connection to statsd server
func (client *Client) Open() {
	connectionString := fmt.Sprintf("%s:%d", client.host, client.port)
	conn, err := net.Dial("udp", connectionString)
	if err != nil {
		log.Println(err)
	}
	client.conn = conn
}

// Close UDP connection to statsd server
func (client *Client) Close() {
	client.conn.Close()
}

// SetAutoflush enables/disables buffered mode
// In buffered mode requires manual call of Flush()
// In autoflush mode message sends to server on every call
func (client *Client) SetAutoflush(autoflush bool) {
	client.autoflush = autoflush
}

// Timing track in milliseconds with sampling
func (client *Client) Timing(key string, time int64, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", time, metricTypeTiming)
	if sampleRate < 1 {
		if client.isSendAcceptedBySampleRate(sampleRate) {
			metricValue = fmt.Sprintf("%s|@%f", metricValue, sampleRate)
		} else {
			return
		}
	}

	client.addToBuffer(key, metricValue)
}

// Count tack
func (client *Client) Count(key string, delta int, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", delta, metricTypeCount)
	if sampleRate < 1 {
		if client.isSendAcceptedBySampleRate(sampleRate) {
			metricValue = fmt.Sprintf("%s|@%f", metricValue, sampleRate)
		} else {
			return
		}
	}

	client.addToBuffer(key, metricValue)
}

// Gauge track
func (client *Client) Gauge(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metricTypeGauge)
	client.addToBuffer(key, metricValue)
}

// Set tracking
func (client *Client) Set(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metricTypeSet)
	client.addToBuffer(key, metricValue)
}

// add to buffer and flush if auto flush enabled
func (client *Client) addToBuffer(key string, metricValue string) {
	// flush
	if client.autoflush {
		// send metric now
		go client.send(fmt.Sprintf("%s:%s", key, metricValue))
	} else {
		// add metric to buffer for next manual flush
		client.keyBufferLock.Lock()
		client.keyBuffer[key] = metricValue
		client.keyBufferLock.Unlock()

	}
}

// Check if acceptable by sample rate
func (client *Client) isSendAcceptedBySampleRate(sampleRate float32) bool {
	if sampleRate >= 1 {
		return true
	}
	randomNumber := client.rand.Float32()
	return randomNumber <= sampleRate
}

// Flush buffer to statsd daemon by UDP
func (client *Client) Flush() {
	// check if buffer has metrics
	if len(client.keyBuffer) == 0 {
		return
	}

	// prepare metric packet
	var metricPacketArray []string

	// lock
	client.keyBufferLock.Lock()

	// build packet
	for key, metricValue := range client.keyBuffer {
		metricPacketArray = append(metricPacketArray, fmt.Sprintf("%s:%s", key, metricValue))
	}
	metricPacket := strings.Join(metricPacketArray, "\n")

	// clear key buffer
	client.keyBuffer = make(map[string]string)

	// lock
	client.keyBufferLock.Unlock()

	// send packet
	go client.send(metricPacket)
}

// Send StatsD packet
func (client *Client) send(metricPacket string) {
	// send metric packet
	_, err := fmt.Fprintf(client.conn, metricPacket)
	if err != nil {
		log.Println(err)
	}
}
