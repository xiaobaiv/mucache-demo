package mucache

import (
	"bytes"
	"demo/address"
	"demo/u"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
)

// Operation 是一个可以是 Call 也可以是 Inv 的接口
type Operation interface {
	isCall() bool
}

// Call 表示一个服务的调用
type Call struct {
	u.CallArgs
}

// Call 实现了 Operation 接口
func (c Call) isCall() bool {
	return true
}

// Inv 表示一个失效的操作
type Inv struct {
	Key u.Key
}

// Inv 实现了 Operation 接口
func (i Inv) isCall() bool {
	return false
}

// CM 代表缓存管理器
type CM struct {
	Cache   *LRUCache
	saved   Saved
	history []Operation // history 现在可以存储 Call 也可以存储 Inv
	mu      sync.Mutex
}

func NewCM(capacity int) *CM {
	return &CM{
		Cache:   NewLRUCache(capacity),
		saved:   *NewSaved(),
		history: make([]Operation, 0),
	}
}
func GetServiceAddress(service u.Service) (string, error) {

	address, ok := address.ServiceAddresses[service]
	if !ok {
		return "", fmt.Errorf("service not found")
	}

	return address, nil
}

func (cm *CM) SendToCM(caller u.Service, ca u.CallArgs, ret string, vs []u.Service) {
	// 构建 Save 消息的内容。
	saveMessage := struct {
		CallArgs u.CallArgs  `json:"call_args"`
		Result   string      `json:"result"`
		Vs       []u.Service `json:"services"`
	}{
		CallArgs: ca,
		Result:   ret,
		Vs:       vs,
	}

	// 序列化 saveMessage 为 JSON，以便发送
	jsonData, err := json.Marshal(saveMessage)
	if err != nil {
		log.Printf("Error marshalling save message: %v\n", err)
		return
	}

	// 函数 GetServiceAddress，它可以根据服务名返回服务的 IP 和端口
	address, err := GetServiceAddress(caller)
	if err != nil {
		log.Printf("Error getting address for service %s: %v\n", caller, err)
		return
	}

	// 发送消息到服务的地址。
	resp, err := http.Post("http://"+address+"/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending save message to %s: %v\n", address, err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send save message to %s, status code: %d\n", address, resp.StatusCode)
	}
}
func (cm *CM) SendInvToCM(caller u.Service, k string) {
	// 构建 Inv 消息的内容
	invMessage := struct {
		Key u.Key `json:"key"`
	}{
		Key: u.Key(k),
	}

	// 序列化 invMessage 为 JSON，以便发送
	jsonData, err := json.Marshal(invMessage)
	if err != nil {
		log.Printf("Error marshalling inv message: %v\n", err)
		return
	}

	// 获取服务的 IP 和端口
	address, err := GetServiceAddress(caller)
	if err != nil {
		log.Printf("Error getting address for service %s: %v\n", caller, err)
		return
	}

	// 发送消息到服务的地址
	resp, err := http.Post("http://"+address+"/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending inv message to %s: %v\n", address, err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send inv message to %s, status code: %d\n", address, resp.StatusCode)
	}
}
func (cm *CM) StartHandler(ca u.CallArgs) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	// 将 Call 对象加入到 history 中
	cm.history = append(cm.history, Call{CallArgs: ca})
}
func (cm *CM) EndHandler(ca u.CallArgs, rs Set, caller u.Service, result string, vs []u.Service) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 找到最后一个 Call(ca) 的索引
	var lastIndex int = -1
	for i := len(cm.history) - 1; i >= 0; i-- {
		if call, ok := cm.history[i].(Call); ok && reflect.DeepEqual(call.CallArgs, ca) {
			lastIndex = i
			break
		}
	}

	// 从最后一个 Call(ca) 的位置开始，检查是否有 Inv(k) 插入到 history 中，且 k 在 rs 中
	for i := lastIndex + 1; i < len(cm.history); i++ {
		if inv, ok := cm.history[i].(Inv); ok {
			if _, existsInRs := rs[string(inv.Key)]; existsInRs {
				break
			}
		}
	}
	// 没有则保存
	cm.SendToCM(caller, ca, result, vs)
	//cm.saved
	for element := range rs {
		cm.saved.Add(element, caller, ca)
	}

}
func (cm *CM) InvHandler(invKey u.Key) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	// 将 Inv 对象加入到 history 中
	cm.history = append(cm.history, Inv{Key: invKey})
	cm.Cache.Delete(string(invKey))
	result, ok := cm.saved.PopByKey(string(invKey))
	if ok {
		for caller, ca := range result {
			cm.SendInvToCM(caller, ca)
		}
	}
}
func (cm *CM) SaveHandler(ca u.CallArgs, ret string, vs []string) {
	cm.Cache.Set(ca.String(), ret, vs, ca)
}

func (cm *CM) Handler(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming JSON payload
	var msg map[string]json.RawMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 检查消息内容来区分消息类型
	if key, ok := msg["key"]; ok {
		// 处理 Inv 消息
		var k string
		err := json.Unmarshal(key, &k)
		if err != nil {
			http.Error(w, "Invalid Inv key", http.StatusBadRequest)
			return
		}
		cm.InvHandler(u.Key(k))
	} else if callArgs, ok := msg["call_args"]; ok {
		// 处理 Save 消息
		var ca u.CallArgs
		err := json.Unmarshal(callArgs, &ca)
		if err != nil {
			http.Error(w, "Invalid CallArgs", http.StatusBadRequest)
			return
		}
		var ret string
		var vs []u.Service
		err = json.Unmarshal(msg["result"], &ret)
		if err != nil {
			http.Error(w, "Invalid result", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(msg["services"], &vs)
		if err != nil {
			http.Error(w, "Invalid services", http.StatusBadRequest)
			return
		}
		// 转换 services 到 []string 以便传递给 SaveHandler
		var serviceStrings []string
		for _, service := range vs {
			serviceStrings = append(serviceStrings, string(service))
		}

		cm.SaveHandler(ca, ret, serviceStrings)
	} else {
		http.Error(w, "Unknown message type", http.StatusBadRequest)
	}
}

// StartServer 开始 HTTP 服务器
func (cm *CM) StartServer(port string) {
	http.HandleFunc("/", cm.Handler)
	// 服务器现在正在监听 Save 和 Inv 消息
	log.Println("CM started on port " + port + ".")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
