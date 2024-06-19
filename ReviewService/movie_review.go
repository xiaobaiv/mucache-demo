package MovieReview

import (
	"bytes"
	"demo/mucache"
	"demo/u"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// MovieReview 定义了一个处理电影评论的服务
type MovieReview struct {
	storageServiceURL string
	CM                *mucache.CM
	readsets          mucache.ReadSets
}

func NewMovieReview(storageServiceURL string, cacheSize int) *MovieReview {
	cm := mucache.NewCM(cacheSize) // 假设NewCM是创建CM实例的函数，且cacheSize是CM的缓存大小参数
	go cm.StartServer("8083")
	return &MovieReview{
		storageServiceURL: storageServiceURL,
		CM:                cm,
		readsets:          mucache.NewReadSets(),
	}
}

// logRequestDetails 用于记录请求的处理细节
func (mr *MovieReview) logRequestDetails(functionName, movieID, key, value string, startTime time.Time, result string) {
	elapsedTime := time.Since(startTime)
	log.Printf("Function: %s, MovieID: %s, Key: %s, Value: %s, Result: %s, Time Taken: %v\n",
		functionName, movieID, key, value, result, elapsedTime)
}

// HandleRead 处理读取电影评论的请求
func (mr *MovieReview) HandleRead(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now() // 请求开始时间

	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 提取参数
	movieID := r.URL.Query().Get("movie_id")
	key := r.URL.Query().Get("key")

	// preReqStart & preWrite
	ca := u.CallArgs{
		Function: "HandleRead",
		Args:     []interface{}{movieID, key},
	}
	mr.CM.StartHandler(ca)
	mr.readsets.Add(ca, key)
	//preCall
	if value, _, _, found := mr.CM.Cache.Get(ca.String()); found {
		// 将字符串值转换为[]byte
		byteValue := []byte(value)

		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue) // 使用转换后的[]byte类型作为参数
		mr.logRequestDetails("HandleRead", movieID, key, "", startTime, "Success (Cached)")
		return
	}

	// 向review storage的服务发送请求
	resp, err := http.Get(mr.storageServiceURL + "/read?movie_id=" + movieID + "&key=" + key)
	if err != nil {
		http.Error(w, "Failed to request review storage service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from review storage service", http.StatusInternalServerError)
		return
	}

	time.Sleep(1000 * time.Millisecond) // 模拟1000ms的网络延迟
	// 记录请求处理细节
	mr.logRequestDetails("HandleRead", movieID, key, "", startTime, "Success")

	// 将review storage服务的响应返回给客户端
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (mr *MovieReview) HandleWrite(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()             // 请求开始时间
	time.Sleep(1000 * time.Millisecond) // 模拟1000ms的网络延迟
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		MovieID string `json:"movie_id"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 序列化input为JSON
	jsonData, err := json.Marshal(input)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// 使用序列化后的JSON创建一个HTTP请求
	req, err := http.NewRequest("POST", mr.storageServiceURL+"/write", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request to review storage service", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to request review storage service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 记录请求处理细节
	responseStatus := "Created"
	if resp.StatusCode != http.StatusCreated {
		responseStatus = "Failed"
	}
	mr.logRequestDetails("HandleWrite", input.MovieID, input.Key, input.Value, startTime, responseStatus)

	w.WriteHeader(resp.StatusCode)
}
