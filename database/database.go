package database

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// KeyValuePair struct
type KeyValuePair struct {
	Value      string
	Expiration time.Time
}

// Database struct to store key-value pairs
type Database struct {
	data   map[string]*KeyValuePair
	lock   sync.Mutex
	ticker *time.Ticker
}

// NewDatabase creates a new instance of Database
func NewDatabase() *Database {
	ds := &Database{
		data: make(map[string]*KeyValuePair),
	}
	ds.startExpiryCleanup()
	return ds
}

// Set sets the value in the database for the given key
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
	fmt.Println("Adding value: " + value + " with expiry: " + expiry.String() + " condition: " + condition)
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

// Get retrieves the value from the database for the given key
func (ds *Database) Get(key string) (string, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if kv, exists := ds.data[key]; exists {
		return kv.Value, nil
	}

	return "", fmt.Errorf("key not found")
}

// QPush appends values to the queue in the database for the given key
func (ds *Database) QPush(key string, values []string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if kv, exists := ds.data[key]; exists {
		kv.Value += concatValues(values)
	} else {
		ds.data[key] = &KeyValuePair{
			Value:      concatValues(values),
			Expiration: time.Time{},
		}
	}
}

// QPop retrieves and removes the last inserted value from the queue in the database for the given key
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

// BQPop retrieves and removes the last inserted value from the queue in the database for the given key.
// If the queue is empty, it blocks the request until a value is available or the timeout is reached.
func (ds *Database) BQPop(key string, timeout time.Duration) (string, error) {
	ds.lock.Lock()
	kv, exists := ds.data[key]
	ds.lock.Unlock()

	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if kv != nil {
		return ds.qPopWithTimeout(key, kv, timeout)
	}

	return ds.bqPopWithTimeout(key, timeout)
}

// qPopWithTimeout retrieves and removes the last inserted value from the queue with the given key and expiration time.
// If the queue is empty, it blocks the request until a value is available or the timeout is reached.
func (ds *Database) qPopWithTimeout(key string, kv *KeyValuePair, timeout time.Duration) (string, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	values := splitValues(kv.Value)
	if len(values) > 0 {
		lastIndex := len(values) - 1
		lastValue := values[lastIndex]
		values = values[:lastIndex]
		newValue := concatValues(values)
		kv.Value = newValue
		return lastValue, nil
	}

	if timeout == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		return "", fmt.Errorf("queue is empty")
	case <-ds.ticker.C:
		ds.lock.Lock()
		defer ds.lock.Unlock()
		if kv, exists := ds.data[key]; exists {
			return ds.qPopWithTimeout(key, kv, 0)
		}
		return "", fmt.Errorf("key not found")
	}
}

// bqPopWithTimeout retrieves and removes the last inserted value from the queue with the given key.
// If the queue is empty, it blocks the request until a value is available or the timeout is reached.
func (ds *Database) bqPopWithTimeout(key string, timeout time.Duration) (string, error) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		return "", fmt.Errorf("queue is empty")
	case <-ds.ticker.C:
		ds.lock.Lock()
		defer ds.lock.Unlock()
		kv, exists := ds.data[key]
		if exists && kv != nil {
			return ds.qPopWithTimeout(key, kv, 0)
		}
		return "", fmt.Errorf("key not found")
	}
}

// startExpiryCleanup starts a goroutine that periodically checks and removes expired keys from the database
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

// concatValues concatenates the values with space separator
func concatValues(values []string) string {
	return strings.Join(values, " ")
}

// splitValues splits the value string into a slice of values
func splitValues(value string) []string {
	return strings.Fields(value)
}
