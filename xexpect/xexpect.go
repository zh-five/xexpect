package xexpect

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/iyzyi/aiopty/pty"
	"github.com/iyzyi/aiopty/term"
)

const (
	stepNew      = 0
	stepRun      = 1
	stepExpect   = 2
	stepInteract = 3
	stepExit     = 4

	matchLen = 32
)

type XExpect struct {
	ptmx    *pty.Pty
	term    *term.Term
	timeout int // 总超时时间，秒

	step     int
	buf      []byte
	start    int
	matchLen int
}

type action struct {
	expect     string
	send       string
	isContinue bool // 匹配成功后是否继续匹配
	isReg      bool // 是否为正则
}

func NewXExpect() *XExpect {
	return &XExpect{
		ptmx:     nil,
		term:     nil,
		timeout:  10,
		step:     stepNew,
		buf:      make([]byte, 1024),
		start:    0,
		matchLen: matchLen,
	}
}

func (sf *XExpect) SetTimeout(second int) {
	if sf.step != stepNew {
		sf.errorf("SetTimeout() must be called before RunCmd()")
	}

	if second < 1 {
		sf.errorf("timeout(second) must be greater than 0 ")
	}

	sf.timeout = second
}

func (sf *XExpect) RunCmd(args []string) {
	if sf.step != stepNew {
		sf.errorf("RunCmd() can be called only once")
	}
	sf.step = stepRun

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
		sf.errorf("Failed to create pty: %v", err)
	}
	//defer p.Close()
	sf.ptmx = p

	// terminal
	t, err := term.Open(os.Stdin, os.Stdout, sf.onSizeChange(p))
	if err == nil {
		sf.term = t
	}

	// 响应手动输入
	if sf.term == nil {
		go func() {
			io.Copy(sf.ptmx, os.Stdin)
		}()
	} else {
		go func() {
			io.Copy(sf.ptmx, sf.term)
		}()
	}

	// timeout
	time.AfterFunc(time.Second*time.Duration(sf.timeout), func() {
		if sf.step > stepExpect {
			return
		}
		sf.close()
		sf.errorf("timeout exit")
	})
}

func (sf *XExpect) Expect(rule [][]string) (int, string) {
	if sf.step < stepRun {
		sf.errorf("Expect() must be called after RunCmd()")
	}
	if sf.step > stepExpect {
		sf.errorf("Expect() must be called befor Exit() and Interact() ")
	}
	sf.step = stepExpect

	listAction := sf.parseRule(rule)
	var isMatch bool
	matchStr := ""

	for {
		n, err := sf.ptmx.Read(sf.buf[sf.start:])
		if err != nil {
			break
		}
		fmt.Print(string(sf.buf[sf.start : sf.start+n])) // 显示pty输出
		if sf.start+n > sf.matchLen {
			text := string(sf.buf[:sf.start+n])
			copy(sf.buf, sf.buf[sf.start+n-sf.matchLen:])
			sf.start = sf.matchLen

			for i, act := range listAction {
				// 匹配
				isMatch = false
				matchStr = ""
				if len(act.expect) == 0 { // 不匹配直接发送
					isMatch = true
					matchStr = ""
				} else if act.isReg {
					matchStr = sf.regMatch(text, act.expect)
					if matchStr != "" {
						isMatch = true
					}
				} else {
					isMatch = strings.Contains(text, act.expect)
				}

				// 匹配成功，
				if isMatch {
					if len(act.send) > 0 {
						//fmt.Println(act.send, "end")
						_, err := sf.ptmx.Write([]byte(act.send))
						if err != nil {
							sf.errorf("send error: %v", err)
						}
					}
					if act.isContinue {
						break
					} else {
						return i, matchStr
					}
				}
			}
		} else {
			sf.start += n
		}
	}

	return -1, ""
}

func (sf *XExpect) Interact() {
	if sf.step < stepRun {
		sf.errorf("Expect() must be called after RunCmd()")
	}
	if sf.step == stepInteract {
		sf.errorf("Interact() does not allow repeated calls")
	}
	if sf.step == stepExit {
		sf.errorf("Interact() can not called after Exit()")
	}
	sf.step = stepInteract
	defer sf.close()

	if sf.term == nil {
		io.Copy(os.Stdout, sf.ptmx)
	} else {
		io.Copy(sf.term, sf.ptmx)
	}
}

func (sf *XExpect) Exit() {
	if sf.step < stepRun {
		sf.errorf("Expect() must be called after RunCmd()")
	}
	if sf.step == stepExit {
		sf.errorf("Exit() does not allow repeated calls")
	}
	sf.step = stepExit

	sf.close()
}

func (sf *XExpect) close() {
	if sf.term != nil {
		sf.term.Close()
	}
	if sf.ptmx != nil {
		sf.ptmx.Close()
	}
}

func (sf *XExpect) regMatch(text string, reg string) string {
	regex, err := regexp.Compile(reg)
	if err != nil {
		sf.errorf("reg (%s) error: %v", reg, err)
	}

	return regex.FindString(text)
}

func (sf *XExpect) parseRule(rule [][]string) []*action {
	out := make([]*action, 0, len(rule))

	numC := 0
	for _, v := range rule {
		if len(v) < 2 {
			sf.errorf("rule error: The actions must contain at least two elements: expect and send")
		}
		act := &action{
			expect:     v[0],
			send:       v[1],
			isContinue: false,
			isReg:      false,
		}
		for i := range v[2:] {
			if v[i] == "C" {
				act.isContinue = true
				numC++
			} else if v[i] == "E" {
				act.isReg = true
			} else {
				sf.errorf("rule error: only 'C' and 'E' flags are allowed for action")
			}
		}
		out = append(out, act)
	}

	if numC == len(rule) {
		sf.errorf("rule error: It's not allowed for all actions to set the C flag")
	}

	return out
}

func (sf *XExpect) onSizeChange(p *pty.Pty) func(uint16, uint16) {
	return func(cols, rows uint16) {
		size := &pty.WinSize{
			Cols: cols,
			Rows: rows,
		}
		p.SetSize(size)
	}
}

func (sf *XExpect) errorf(format string, vals ...any) {
	fmt.Fprintf(os.Stderr, format, vals...)
	os.Exit(1)
}

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
