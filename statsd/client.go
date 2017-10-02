package statsd

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
	"strings"
)

const metricTypeCount = 'c';
const metricTypeGauge = 'g';
const metricTypeTiming = 't';
const metricTypeSet = 's';

// The Client type
type Client struct {
	host      string
	port      int
	conn      net.Conn // UDP connection to StatsD server
	rand      *rand.Rand // rand generator to skip messages by sample rate
	keyBuffer map[string]string // array of messages to send
	autoflush bool // send metrics on every call
}

// New StatsD client
func NewClient(host string, port int) *Client {
	client := Client{
		host: host,
		port: port,
		rand: rand.New(rand.NewSource(time.Now().Unix())),
		keyBuffer: make(map[string]string),
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
	client.autoflush = false
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
func (client *Client) Count(key string, delta int, sampleRate float32) {
	metricValue := fmt.Sprintf("%d|%s", delta, metricTypeCount)
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
func (client *Client) Gauge(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metricTypeGauge)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
	}
}

// Set tracking
func (client *Client) Set(key string, value int) {
	metricValue := fmt.Sprintf("%d|%s", value, metricTypeSet)
	client.keyBuffer[key] = metricValue
	if client.autoflush {
		client.Flush()
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

// flush data to statsd daemon by UDP
func (client *Client) Flush() {
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
