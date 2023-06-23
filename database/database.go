package database

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type KeyValuePair struct {
	Value      string
	Expiration time.Time
}

// struct to store key-value pairs
type Database struct {
	data   map[string]*KeyValuePair
	lock   sync.RWMutex
	ticker *time.Ticker
}

// create a new instance of Database
func NewDatabase() *Database {
	ds := &Database{
		data: make(map[string]*KeyValuePair),
	}
	ds.startExpiryCleanup()
	return ds
}

// set the value in the database for the given key
func (ds *Database) Set(key, value string, expiry time.Duration, condition string) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if condition == "NX" {
		if _, exists := ds.data[key]; exists {
			return fmt.Errorf("key already exists")
		}
	} else if condition == "XX" {
		if _, exists := ds.data[key]; !exists {
			return fmt.Errorf("key does not exist")
		}
	}
	// log.Println("Adding key:value -> " + key + " : " + value + " with expiry: " + expiry.String() + " condition: " + condition)
	expiration := time.Time{}
	if expiry > 0 {
		expiration = time.Now().Add(expiry)
	}

	ds.data[key] = &KeyValuePair{
		Value:      value,
		Expiration: expiration,
	}

	return nil
}

// retrieve the value from the database for the given key
func (ds *Database) Get(key string) (string, error) {
	ds.lock.RLock()
	defer ds.lock.RUnlock()

	if kv, exists := ds.data[key]; exists {
		return kv.Value, nil
	}

	return "", fmt.Errorf("key not found")
}

// append values to the queue in the database for the given key
func (ds *Database) QPush(key string, values []string) {
	for {
		if ds.lock.TryLock() {
			if kv, exists := ds.data[key]; exists {
				kv.Value += " " + concatValues(values)
			} else {
				ds.data[key] = &KeyValuePair{
					Value:      concatValues(values),
					Expiration: time.Time{},
				}
			}
			ds.lock.Unlock()
			break // Exit the loop if the lock was acquired successfully
		}

		// Add a short delay before retrying to acquire the lock
		time.Sleep(100 * time.Millisecond)
	}
}

// retrieve and removes the last inserted value from the queue in the database for the given key
func (ds *Database) QPop(key string) (string, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if kv, exists := ds.data[key]; exists {
		values := splitValues(kv.Value)
		if len(values) > 0 {
			lastIndex := len(values) - 1
			lastValue := values[lastIndex]
			values = values[:lastIndex]
			newValue := concatValues(values)
			kv.Value = newValue
			return lastValue, nil
		}
		return "", fmt.Errorf("queue is empty")
	}

	return "", fmt.Errorf("key not found")
}

// retrieve and removes the last inserted value from the queue in the database for the given key.
// If the queue is empty, it blocks the request until a value is available or the timeout is reached.
func (ds *Database) BQPop(key string, timeout time.Duration) (string, error) {
	ds.lock.RLock()
	queue, exists := ds.data[key]
	ds.lock.RUnlock()

	if exists {
		if ds.lock.TryLock() {
			values := splitValues(queue.Value)
			if len(values) > 0 {
				lastIndex := len(values) - 1
				lastValue := values[lastIndex]
				values = values[:lastIndex]
				newValue := concatValues(values)
				queue.Value = newValue
				ds.lock.Unlock()
				return lastValue, nil
			}
			ds.lock.Unlock()
		}

		timer := time.NewTimer(timeout)
		valueCh := make(chan string, 1)

		go func() {
			for {
				if ds.lock.TryLock() {
					queue, exists := ds.data[key]
					if exists {
						values := splitValues(queue.Value)
						if len(values) > 0 {
							lastIndex := len(values) - 1
							lastValue := values[lastIndex]
							values = values[:lastIndex]
							newValue := concatValues(values)
							queue.Value = newValue
							ds.lock.Unlock()
							valueCh <- lastValue
							return
						}
					}
					ds.lock.Unlock()
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()

		select {
		case value := <-valueCh:
			return value, nil
		case <-timer.C:
			return "", nil
		}
	}
	ds.lock.Unlock()
	return "", fmt.Errorf("key not found")
}

// start a goroutine that periodically checks and removes expired keys from the database
func (ds *Database) startExpiryCleanup() {
	ds.ticker = time.NewTicker(time.Second)
	go func() {
		for range ds.ticker.C {
			ds.lock.Lock()
			for key, kv := range ds.data {
				if kv.Expiration != (time.Time{}) && time.Now().After(kv.Expiration) {
					delete(ds.data, key)
				}
			}
			ds.lock.Unlock()
		}
	}()
}

func concatValues(values []string) string {
	return strings.Join(values, " ")
}

func splitValues(value string) []string {
	return strings.Fields(value)
}
