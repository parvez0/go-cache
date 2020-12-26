package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	ReplicateActionInsert = "insert"
	ReplicateActionDelete = "delete"
)

type Replicate struct {
	Action string `json:"action"`
	Key string `json:"key"`
	Data map[string]interface{} `json:"data"`
}

var replicateBucket = make(chan Replicate, 10)

func SendDataToSlave()  {
	for  {
		select {
		case data, open := <-replicateBucket:
			client := http.Client{
				Timeout: time.Second * 5,
			}
			fmt.Println("replicateBucket is open : ", open, "  :  ",data)
			buf, _ := json.Marshal(data)
			io := bytes.NewReader(buf)
			req, err := http.NewRequest(http.MethodPost, os.Getenv("SLAVE_URL") + "/replicate", io)
			if err != nil {
				fmt.Printf("SlaveConnectionFailed: failed to replicate data to slave, %s\n", err.Error())
				continue
			}
			headers := http.Header{}
			headers.Set("Content-Type", "application/json")
			req.Header = headers
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("SlaveConnectionFailed: failed to replicate data to slave, %s\n", err.Error())
				continue
			}
			if res.StatusCode == http.StatusNotFound {
				fmt.Println("SlaveNotReachable: check if slave is running on " + os.Getenv("SLAVE_URL"))
				continue
			}
			if res.StatusCode == http.StatusAccepted {
				fmt.Println("SlaveReplicatino: data replicated")
				continue
			}
		}
	}
}

func ReplicateData(meta *Replicate)  {
	switch meta.Action {
	case ReplicateActionInsert:
		Set(meta.Data)
	case ReplicateActionDelete:
		Delete(meta.Key)
	}
}