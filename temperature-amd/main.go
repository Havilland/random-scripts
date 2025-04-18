package main

import (
	"fmt"
	"os/exec"
	"log"
	"strings"
	"github.com/antchfx/jsonquery"
	"errors"
	"time"
)

// RingBuffer represents a fixed-size circular buffer for float64 values
type RingBuffer struct {
	data        []float64
	size        int
	capacity    int
	head        int
	tail        int
	isFull      bool
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) (*RingBuffer, error) {
	if capacity <= 0 {
			return nil, errors.New("capacity must be positive")
	}
	
	return &RingBuffer{
			data:     make([]float64, capacity),
			size:     0,
			capacity: capacity,
			head:     0,
			tail:     0,
			isFull:   false,
	}, nil
}

// Add inserts a value into the buffer, potentially overwriting the oldest value
func (rb *RingBuffer) Add(value float64) {
	// If buffer is full, we'll overwrite the oldest item (at head)
	if rb.isFull {
			rb.head = (rb.head + 1) % rb.capacity
	} else {
			rb.size++
	}
	
	// Add the new value at the tail position
	rb.data[rb.tail] = value
	rb.tail = (rb.tail + 1) % rb.capacity
	
	// Check if buffer is now full
	rb.isFull = rb.size == rb.capacity
}

// GetAll returns all values currently in the buffer (oldest first)
func (rb *RingBuffer) GetAll() []float64 {
	if rb.size == 0 {
			return []float64{}
	}
	
	result := make([]float64, rb.size)
	for i := 0; i < rb.size; i++ {
			idx := (rb.head + i) % rb.capacity
			result[i] = rb.data[idx]
	}
	
	return result
}

// Average calculates the average of all values in the buffer
func (rb *RingBuffer) Average() (float64, error) {
	if rb.size == 0 {
			return 0, errors.New("buffer is empty")
	}
	
	sum := 0.0
	for i := 0; i < rb.size; i++ {
			idx := (rb.head + i) % rb.capacity
			sum += rb.data[idx]
	}
	
	return sum / float64(rb.size), nil
}

// Size returns the current number of elements in the buffer
func (rb *RingBuffer) Size() int {
	return rb.size
}

// Capacity returns the maximum capacity of the buffer
func (rb *RingBuffer) Capacity() int {
	return rb.capacity
}

func main() {
	// Initialize Buffers
	rbTctl_1, err := NewRingBuffer(60)
	if err != nil {
		log.Fatal(err)
	}
	rbTccd1_1, err := NewRingBuffer(60)
	if err != nil {
		log.Fatal(err)
	}
	rbTccd2_1, err := NewRingBuffer(60)
	if err != nil {
		log.Fatal(err)
	}
	rbTctl_5, err := NewRingBuffer(300)
	if err != nil {
		log.Fatal(err)
	}
	rbTccd1_5, err := NewRingBuffer(300)
	if err != nil {
		log.Fatal(err)
	}
	rbTccd2_5, err := NewRingBuffer(300)
	if err != nil {
		log.Fatal(err)
	}
	counter := 0

	for {
		cmd := exec.Command("sensors", "-j")

		stdin, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		
		data := string(stdin)

		doc, err := jsonquery.Parse(strings.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}

		// Get Temperatures for AMD CPU
		tctl := jsonquery.FindOne(doc, "k10temp-pci-00c3/Tctl/temp1_input").Value().(float64)
		tccd1 := jsonquery.FindOne(doc, "k10temp-pci-00c3/Tccd1/temp3_input").Value().(float64)
		tccd2 := jsonquery.FindOne(doc, "k10temp-pci-00c3/Tccd2/temp4_input").Value().(float64)
		
		rbTctl_1.Add(tctl)
		rbTccd1_1.Add(tccd1)
		rbTccd2_1.Add(tccd2)
		rbTctl_5.Add(tctl)
		rbTccd1_5.Add(tccd1)
		rbTccd2_5.Add(tccd2)

		// Calculate 1 min average
		tctlAvg_1, _ := rbTctl_1.Average()
		tccd1Avg_1, _ := rbTccd1_1.Average()
		tccd2Avg_1, _ := rbTccd2_1.Average()

		// Calculate 5 min average
		tctlAvg_5, _ := rbTctl_5.Average()
		tccd1Avg_5, _ := rbTccd1_5.Average()
		tccd2Avg_5, _ := rbTccd2_5.Average()

		fmt.Printf("*********************************\n")
		fmt.Printf("Time running: %#v\n", counter)
		fmt.Printf("Tctl avg: %.2f  %.2f\n", tctlAvg_1, tctlAvg_5)
		fmt.Printf("Tccd1 avg: %.2f  %.2f\n", tccd1Avg_1, tccd1Avg_5)
		fmt.Printf("Tccd2 avg: %.2f  %.2f\n", tccd2Avg_1, tccd2Avg_5)
		fmt.Printf("*********************************\n")
		time.Sleep(time.Second) 
		counter++
	}
}