package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/parvez0/redis/cache"
	"github.com/parvez0/redis/responses"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init()  {
	fmt.Println("initializing database server")
	if os.Getenv("SLAVE_URL") == "" {
		os.Setenv("SLAVE_URL", "http://localhost:8090")
	}
	if os.Getenv("NODE_ROLE") == "" {
		os.Setenv("NODE_ROLE", "master")
	}

	if os.Getenv("GO_PORT") == "" {
		port := "8080"
		if os.Getenv("NODE_ROLE") == "slave" {
			port = "8090"
		}
		os.Setenv("GO_PORT", port)
	}
}

func main()  {
	port := os.Getenv("GO_PORT")
	role := os.Getenv("NODE_ROLE")

	// creating a go routine for listening to the shared channel
	go cache.InsertData()

	if role != "slave" {
		go cache.SendDataToSlave()
	}

	// registering handlers with http
	http.HandleFunc("/health-check", func(writer http.ResponseWriter, request *http.Request) {
		resp := responses.GenericResponse{
			Success: true,
			Message: "Server is up and ready to accept the connections",
			Data:    nil,
		}
		buf, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("JSONMarshalFailedHealthCheck: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Server encountered a problem"))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(buf)
		return
	})

	http.HandleFunc("/replicate", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			body := &cache.Replicate{}
			buf, err := ioutil.ReadAll(request.Body)
			if err != nil {
				fmt.Printf("FailedToReadRequestBodySetMethod: %s\n", err.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Server encountered a problem"))
			}
			defer request.Body.Close()
			err = json.Unmarshal(buf, body)
			if err != nil {
				fmt.Printf("JsonUnmarshalFailedSetMethod: %s", err.Error())
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte("please provide a json key, value pair"))
				return
			}
			go cache.ReplicateData(body)
			writer.WriteHeader(http.StatusAccepted)
			writer.Write(buf)
		default:
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Resource not found"))
			return
		}
	})

	http.HandleFunc("/set", func(writer http.ResponseWriter, request *http.Request) {
		defer func() {

		}()
		switch request.Method {
		case http.MethodPost:
			body := make(map[string]interface{})
			buf, err := ioutil.ReadAll(request.Body)
			if err != nil {
				fmt.Printf("FailedToReadRequestBodySetMethod: %s", err.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Server encountered a problem"))
			}
			defer request.Body.Close()
			err = json.Unmarshal(buf, &body)
			if err != nil {
				fmt.Printf("JsonUnmarshalFailedSetMethod: %s", err.Error())
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte("please provide a json key, value pair"))
				return
			}
			resp := cache.Set(body)
			buf, _ = json.Marshal(resp)
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(buf)
		default:
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Resource not found"))
			return
		}
	})

	http.HandleFunc("/get", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			key := request.URL.Query().Get("key")
			if key == "" {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte("mandatory field key is not provided"))
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(cache.Get(key))
			if err != nil {
				fmt.Println("GetKeyBufferEncoderFailed: ", err)
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Server encountered an error"))
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(buf.Bytes())
		default:
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Resource not found"))
			return
		}
	})

	http.HandleFunc("/list", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			buf, _ := json.Marshal(cache.List())
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(buf)
		default:
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Resource not found"))
			return
		}
	})

	http.HandleFunc("/delete", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodDelete:
			key := request.URL.Query().Get("key")
			if key == "" {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte("mandatory field key is not provided"))
			}
			buf, _ := json.Marshal(cache.Delete(key))
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(buf)
		default:
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Resource not found"))
			return
		}
	})

	// ********** end of handlers
	fmt.Printf("starting the %s server on port : %s\n", role, port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		fmt.Errorf("FailedToLoadServer: %s", err.Error())
	}
	// creating the channel for holding the signal which will be received
	sig := make(chan os.Signal)
	// waiting for kill signal from command line to close the connection
	signal.Notify(sig, syscall.SIGKILL | syscall.SIGTERM)
	<-sig
	fmt.Printf("closing the connection")
}
