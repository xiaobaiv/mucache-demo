package MovieReview

import (
	"demo/mucache"
	"demo/u"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// 定义存储评论的结构体
type ReviewStorage struct {
	sync.RWMutex
	Data     map[string]map[string]string // 外层map的key是movie_id，内层map的key是评论的key，值是评论内容
	CM       *mucache.CM                  // 缓存管理器的引用
	readsets mucache.ReadSets
}

// 新建一个ReviewStorage实例，同时指定缓存大小
func NewReviewStorage(cacheSize int) *ReviewStorage {
	cm := mucache.NewCM(cacheSize) // 在这里根据传入的缓存大小创建CM实例
	log.Println("卡住了吗")
	go cm.StartServer("8084") // 新的进程
	log.Println("没有")
	return &ReviewStorage{
		Data:     make(map[string]map[string]string),
		CM:       cm, // 初始化CM
		readsets: mucache.NewReadSets(),
	}
}

// 日志打印请求细节
func _logRequestDetails(functionName, movieID, key, value string, startTime time.Time, result string) {
	elapsedTime := time.Since(startTime)
	log.Printf("Function: %s, MovieID: %s, Key: %s, Value: %s, Result: %s, Time Taken: %v\n",
		functionName, movieID, key, value, result, elapsedTime)
}

// read处理函数
func (rs *ReviewStorage) HandleRead(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now() // 请求开始时间

	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	movieID := r.URL.Query().Get("movie_id")
	key := r.URL.Query().Get("key")

	// preReqStart & preWrite
	ca := u.CallArgs{
		Function: "HandleRead",
		Args:     []interface{}{movieID, key},
	}
	rs.CM.StartHandler(ca)
	rs.readsets.Add(ca, key)

	rs.RLock() // 读取锁
	defer rs.RUnlock()
	if data, ok := rs.Data[movieID]; ok {
		if value, ok := data[key]; ok {
			json.NewEncoder(w).Encode(map[string]string{"value": value})
			// 打印请求细节
			_logRequestDetails("handleRead", movieID, key, "", startTime, "Success")
			// preReturn
			readset, _ := rs.readsets.Pop(ca)
			var caller u.Service = "ReviewCM"
			var vs []u.Service
			rs.CM.EndHandler(ca, readset, caller, value, vs)
			return
		}
	}
	http.NotFound(w, r)
	// 打印请求细节
	_logRequestDetails("handleRead", movieID, key, "", startTime, "Not Found")

}

// write处理函数
func (rs *ReviewStorage) HandleWrite(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now() // 请求开始时间

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

	rs.Lock() // 写入锁
	defer rs.Unlock()
	if _, ok := rs.Data[input.MovieID]; !ok {
		rs.Data[input.MovieID] = make(map[string]string)
	}

	rs.Data[input.MovieID][input.Key] = input.Value
	// postWrite
	rs.CM.SendInvToCM("StorageCM", input.Key)

	w.WriteHeader(http.StatusCreated)
	// 打印请求细节
	_logRequestDetails("handleWrite", input.MovieID, input.Key, input.Value, startTime, "Created")
}
