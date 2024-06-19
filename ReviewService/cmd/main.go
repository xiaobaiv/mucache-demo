package main

import (
	MovieReview "demo/ReviewService"
	"log"
	"net/http"
)

func main() {

	// 创建MovieReview服务实例
	movieReviewService := MovieReview.NewMovieReview("http://localhost:8080", 2)

	// 设置路由处理函数
	http.HandleFunc("/read", movieReviewService.HandleRead)
	http.HandleFunc("/write", movieReviewService.HandleWrite)

	// 启动HTTP服务器
	log.Println("Starting Movie Review server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
