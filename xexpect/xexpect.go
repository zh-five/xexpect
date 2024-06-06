package xexpect

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/iyzyi/aiopty/pty"
	"github.com/iyzyi/aiopty/term"
)

type SendInfo struct {
	Keyword string // 等待出现的字符串
	Send    string // 发送的字符串
}

func RunCmd(args []string, timeout int, sends []*SendInfo) {
	// open a pty with options
	opt := &pty.Options{
		Path: args[0],
		Args: args,
		Dir:  "",
		Env:  nil,
		Size: &pty.WinSize{
			Cols: 120,
			Rows: 30,
		},
		Type: "",
	}
	p, err := pty.OpenWithOptions(opt)
	if err != nil {
		log.Panic("Failed to create pty:", err)
		return
	}
	defer p.Close()

	// enable terminal
	t, err := term.Open(os.Stdin, os.Stdout, onSizeChange(p))
	if err != nil {
		doSends(timeout, p, sends)
		go func() {
			io.Copy(p, os.Stdin)
		}()
		io.Copy(os.Stdout, p)
		return
	}
	defer t.Close()

	doSends(timeout, p, sends)
	// start data exchange between terminal and pty
	go func() {
		io.Copy(p, t)
	}()
	io.Copy(t, p)
}

// When the terminal window size changes, synchronize the size of the pty
func onSizeChange(p *pty.Pty) func(uint16, uint16) {
	return func(cols, rows uint16) {
		size := &pty.WinSize{
			Cols: cols,
			Rows: rows,
		}
		p.SetSize(size)
	}
}

func doSends(timeout int, rw io.ReadWriter, sends []*SendInfo) {
	tch := time.After(time.Second * time.Duration(timeout))
	buf := make([]byte, 1024)
	start := 0

	for _, v := range sends {
		start = sendOne(rw, tch, v.Keyword, v.Send, buf, start)
	}
}

func sendOne(rw io.ReadWriter, tch <-chan time.Time, keyword, send string, buf []byte, start int) int {
	// 检查并尝试输入密码
	kLen := len(keyword)
	for {
		n, err := rw.Read(buf[start:])
		if err != nil {
			break
		}
		fmt.Print(string(buf[start : start+n]))
		if start+n > kLen {
			if strings.Contains(string(buf[:n+start]), keyword) {
				_, err := rw.Write([]byte(send + "\n"))
				if err != nil {
					fmt.Println("send error:", err)
				}
				break
			}
			copy(buf, buf[n+start-kLen:])
			start = kLen
		} else {
			start += n
		}

		select {
		case <-tch:
			fmt.Fprintln(os.Stderr, "timeout")
			os.Exit(1)
		default:
		}
	}

	return start
}
