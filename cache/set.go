package cache

import (
	"fmt"
	"github.com/parvez0/go-cache/responses"
	"sync"
)

var mutex sync.Mutex

// global holder of all the data
var database = make(map[string]interface{})

var bucket = make(chan map[string]interface{}, 10) //channel for storing the values, received by another goroutine

func InsertData()  {
	for {
		select {
		case data, ok := <-bucket:
			fmt.Println("channel open - ", ok)
			mutex.Lock()
			for k, v := range data {
				database[k] = v
			}
			slaveData := Replicate{
				Action: ReplicateActionInsert,
				Key:    "",
				Data:   data,
			}
			replicateBucket <- slaveData
			mutex.Unlock()
		}
	}
}

func Set(data map[string]interface{}) *responses.CurdOp {
	ops := responses.CurdOp{}
	for k, _ := range data {
		if database[k] != "" {
			ops.Modified += 1
			continue
		}
		ops.Inserted += 1
	}
	bucket <- data
	return &ops
}
