package u

import "fmt"

type Key string
type Service string
type MessageType string
type CallArgs struct {
	Function string
	Args     []interface{}
}

func (ca CallArgs) String() string {
	argsStr := ""
	argsStr += ca.Function
	for _, arg := range ca.Args {
		argsStr += fmt.Sprintf("%v_", arg) // 使用下划线作为分隔符
	}
	return fmt.Sprintf("%s(%s)", ca.Function, argsStr)
}
