package main

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/zh-five/xexpect/xexpect"
)

type JS struct {
	vm *goja.Runtime
	xe *xexpect.XExpect
}

func NewJS() *JS {
	return &JS{
		vm: goja.New(),
		xe: xexpect.NewXExpect(),
	}
}

func (sf *JS) Run(jsCode string) {
	sf.vm.Set("xe_run", sf.xe_run)                // func(args []string)
	sf.vm.Set("xe_setTimieout", sf.xe_setTimeout) // func(second int)
	sf.vm.Set("xe_matchs", sf.xe_matchs)          // func (rule [][]string) map[string]any{"idx": idx, "str": str}
	sf.vm.Set("xe_term", sf.xe_term)              // func()
	sf.vm.Set("xe_exit", sf.xe_exit)              // func()
	sf.vm.Set("xe_println", sf.xe_println)        // func(msg string)

	_, err := sf.vm.RunString(jsCode)
	if err != nil {
		panic(err)
	}
}

// func(args []string)
func (sf *JS) xe_run(value goja.FunctionCall) goja.Value {
	args := sf.formatArgs(value.Argument(0))

	fmt.Println(args)

	sf.xe.Run(args)
	return sf.vm.ToValue(nil)
}

func (sf *JS) xe_setTimeout(value goja.FunctionCall) goja.Value {
	sec, ok := value.Argument(0).Export().(int)
	if !ok {
		sf.errorf("setTimeout args error: must number")
	}

	sf.xe.SetTimeout(sec)
	return sf.vm.ToValue(nil)
}

// func (rule [][]string) map[string]any{"idx": idx, "str": str}
func (sf *JS) xe_matchs(value goja.FunctionCall) goja.Value {
	rule := sf.formatRule(value.Argument(0))

	idx, str := sf.xe.Matchs(rule)

	return sf.vm.ToValue(map[string]any{"idx": idx, "str": str})
}

func (sf *JS) xe_term(_ goja.FunctionCall) goja.Value {
	sf.xe.Term()

	return sf.vm.ToValue(nil)
}

func (sf *JS) xe_exit(_ goja.FunctionCall) goja.Value {
	sf.xe.Exit()

	return sf.vm.ToValue(nil)
}

func (sf *JS) xe_println(call goja.FunctionCall) goja.Value {
	str := call.Argument(0)
	fmt.Println(str.String())
	return str
}

func (sf *JS) formatArgs(value goja.Value) []string {
	arrayInterface, ok := value.Export().([]interface{})
	if !ok {
		sf.errorf("run args error: must string array")
	}

	var args []string
	for _, item := range arrayInterface {
		str, ok := item.(string)
		if !ok {
			sf.errorf("run args error: must string array")
		}
		args = append(args, str)
	}

	return args
}

func (sf *JS) formatRule(value goja.Value) [][]string {
	var rule [][]string

	// 断言Value是一个数组
	arr1, ok := value.Export().([]interface{})
	if !ok {
		errorf("matchs args error: must two-dimensional string array")
	}

	// 遍历第一层数组
	for i := 0; i < len(arr1); i++ {
		arr2, ok := arr1[i].([]interface{})
		if !ok {
			errorf("matchs args error: must two-dimensional string array")
		}

		var innerArray []string

		// 遍历内层数组
		for _, elem := range arr2 {
			strVal, ok := elem.(string)
			if !ok {
				errorf("matchs args error: must two-dimensional string array")

			}
			innerArray = append(innerArray, strVal)
		}

		rule = append(rule, innerArray)
	}

	return rule
}

func (sf *JS) errorf(format string, vals ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", vals...)
	if sf.xe != nil {
		sf.xe.Exit()
	}
	os.Exit(1)
}
