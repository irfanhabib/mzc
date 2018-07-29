package queue

import (
	"fmt"
)

// Queue interface
type Queue interface {
	Enqueue(string) error
	Dequeue() (string, error)
	IsEmpty() bool
}

// BasicQueue basic Implementation
type BasicQueue struct {
	queue []string
	index int
}

func New(size int) Queue {
	return &BasicQueue{
		queue: make([]string, size),
		index: -1,
	}
}

func (this *BasicQueue) IsEmpty() bool {
	return this.index == -1
}
func (this *BasicQueue) Enqueue(val string) error {
	if this.index < len(this.queue) {
		if this.index == -1 {
			this.index = 0
		} else {
			this.index++
		}
		this.queue[this.index] = val
	} else {
		return fmt.Errorf("Queue is full")
	}
	return nil

}

func (this *BasicQueue) Dequeue() (string, error) {
	if this.index >= 0 {
		val := this.queue[this.index]
		this.index--
		return val, nil
	}
	return "", fmt.Errorf("Queue is empty")

}
