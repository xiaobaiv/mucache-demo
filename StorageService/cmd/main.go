package main

import (
	MovieReview "demo/StorageService"
	"log"
	"net/http"
)

func main() {
	reviewStorage := MovieReview.NewReviewStorage(2)
	// 为了示例，这里预先填充一些数据
	reviewStorage.Data["1"] = map[string]string{"key1": "It's a great movie."}
	reviewStorage.Data["2"] = map[string]string{"key2": "It's an amazing movie."}
	reviewStorage.Data["3"] = map[string]string{"key3": "The movie was excellent."}
	reviewStorage.Data["4"] = map[string]string{"key4": "I really enjoyed this film."}
	reviewStorage.Data["5"] = map[string]string{"key5": "This movie is fantastic!"}
	reviewStorage.Data["6"] = map[string]string{"key6": "The plot was captivating."}
	reviewStorage.Data["7"] = map[string]string{"key7": "It's a must-watch!"}
	reviewStorage.Data["8"] = map[string]string{"key8": "I was blown away by this movie."}
	reviewStorage.Data["9"] = map[string]string{"key9": "One of the best movies I've seen."}
	reviewStorage.Data["10"] = map[string]string{"key10": "Highly recommend this movie!"}

	http.HandleFunc("/read", reviewStorage.HandleRead)
	http.HandleFunc("/write", reviewStorage.HandleWrite)

	log.Println("Starting review storage service on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
