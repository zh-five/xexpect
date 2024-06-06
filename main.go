package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zh-five/xexpect/xexpect"
)

const (
	flagE   string = "-e"
	flagS   string = "-s"
	flagT   string = "-t"
	flagArg string = "args"
)

var allAllow = map[string]string{
	flagE:   "expect",
	flagS:   "send",
	flagT:   "timeout/s",
	flagArg: "",
}

func main() {
	args, timeout, sends := flagParse()

	xexpect.RunCmd(args, timeout, sends)

	// fmt.Println("args:", args)
	// fmt.Println("timeout:", timeout)
	// fmt.Printf("sends:%d\n", len(sends))
	// for i, v := range sends {
	// 	fmt.Printf("\t%d: %+v\n", i, v)
	// }
}

func flagParse() ([]string, int, []*xexpect.SendInfo) {
	args := []string{}
	timeout := -1
	sends := []*xexpect.SendInfo{}

	allow := mkAllow([]string{flagE, flagT, flagArg})
	osArgs := os.Args[1:]
	aLen := len(osArgs)
	for i := 0; i < aLen; i++ {
		v := osArgs[i]

		switch v {
		case flagE:
			if _, ok := allow[flagE]; !ok {
				showErrorf("Error: don't expect -e (expect), expect: %s\n", getAllowDesc(allow))
			}
			sends = append(sends, &xexpect.SendInfo{
				Keyword: osArgs[i+1],
				Send:    "",
			})
			allow = mkAllow([]string{flagS})
			i++
		case flagS:
			if _, ok := allow[flagS]; !ok {
				showErrorf("Error: don't expect -s (send), expect: %s\n", getAllowDesc(allow))
			}
			sends[len(sends)-1].Send = osArgs[i+1]
			i++
			allow = mkAllow([]string{flagE, flagT, flagArg})
		case flagT:
			if _, ok := allow[flagT]; !ok {
				showErrorf("Error: don't expect -t (timeout/s), expect: %s\n", getAllowDesc(allow))
			}
			tmp := osArgs[i+1]
			i++
			allow = mkAllow([]string{flagE, flagArg})
			t, err := strconv.Atoi(tmp)
			if err != nil || t < 1 {
				showErrorf("Error: -t (timeout/s) value must be a positive integer")
			}
			timeout = t
		default:
			if _, ok := allow[flagArg]; !ok {
				showErrorf("Error: don't expect args, expect %s\n", getAllowDesc(allow))
			}
			args = osArgs[i:]
			goto END
		}
	}
END:
	if timeout < 1 {
		timeout = 30
	}

	return args, timeout, sends
}

func showErrorf(format string, val ...any) {
	fmt.Fprintf(os.Stderr, format, val...)
	fmt.Println()
	usage()
	os.Exit(1)
}

func getAllowDesc(allow map[string]string) string {
	list := make([]string, 0, len(allow))
	for k, v := range allow {
		tmp := k
		if v != "" {
			tmp += " (" + v + ") "
		}
		list = append(list, tmp)
	}

	//fmt.Println("keys:", keys)

	return strings.Join(list, ",")
}

func mkAllow(flags []string) map[string]string {
	allow := map[string]string{}
	for _, v := range flags {
		allow[v] = allAllow[v]
	}

	return allow
}

func usage() {
	fmt.Println(`usage:
xexpect [-t timeout/s] [[-e expect -s send] -e expect -s send ...] args
xexpect -h 
    -t: timeout/s default 30s
    -e: expect string, can be an empty string
    -s: send string, must follow the -e
    args: command argument
example: 
xexpect ssh -p 22 root@host
xexpect -t 3 -e 'password: ' -s 123456 ssh -p 22 root@host
xexpect -e 'password: ' -s "123456" ssh -p 22 root@host
xexpect -e 'password: ' -s "123456" -e "#" -s "cd /data/git/" ssh -p 22 root@host
    `)
}
