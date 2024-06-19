package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const movieReviewServiceURL = "http://localhost:8081"

func logRequestDetails_(operation, movieID, key, value string, startTime time.Time, responseStatus string) {
	elapsedTime := time.Since(startTime)
	fmt.Printf("Operation: %s, MovieID: %s, Key: %s, Value: %s, Response Status: %s, Time Taken: %v\n",
		operation, movieID, key, value, responseStatus, elapsedTime)
}

func read(movieID, key string) {
	startTime := time.Now() // 请求开始时间
	response, err := http.Get(fmt.Sprintf("%s/read?movie_id=%s&key=%s", movieReviewServiceURL, movieID, key))
	responseStatus := "Unknown"
	if err != nil {
		fmt.Println("Error calling movie review service:", err)
	} else {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
		} else {
			fmt.Println("Response from movie review service:", string(body))
		}
		responseStatus = response.Status
	}

	logRequestDetails_("read", movieID, key, "", startTime, responseStatus)
}

func write(movieID, key, value string) {
	startTime := time.Now() // 请求开始时间
	data := map[string]string{
		"movie_id": movieID,
		"key":      key,
		"value":    value,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return
	}

	response, err := http.Post(movieReviewServiceURL+"/write", "application/json", bytes.NewBuffer(jsonData))
	responseStatus := "Unknown"
	if err != nil {
		fmt.Println("Error calling movie review service:", err)
	} else {
		defer response.Body.Close()
		if response.StatusCode == http.StatusCreated {
			fmt.Println("Write operation successful")
		} else {
			fmt.Println("Write operation failed with status code:", response.Status)
		}
		responseStatus = response.Status
	}

	logRequestDetails_("write", movieID, key, value, startTime, responseStatus)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("read operation: page -operation=read -movie_id=<movie_id> -key=<key>")
		fmt.Println("write operation: page -operation=write -movie_id=<movie_id> -key=<key> -value=<value>")
	}

	operation := flag.String("operation", "", "Operation to perform (read/write)")
	movieID := flag.String("movie_id", "", "ID of the movie")
	key := flag.String("key", "", "Key for the operation")
	value := flag.String("value", "", "Value for write operation")
	flag.Parse()

	if *operation != "read" && *operation != "write" {
		flag.Usage()
		os.Exit(1)
	}

	switch *operation {
	case "read":
		read(*movieID, *key)
	case "write":
		write(*movieID, *key, *value)
	default:
		flag.Usage()
		os.Exit(1)
	}
}
