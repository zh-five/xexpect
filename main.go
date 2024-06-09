package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zh-five/xtool/crypto/xaes"
)

const (
	EXT_js   = ".js"
	EXT_xexp = ".xexpect"
)

var KEY = []byte{0x84, 0xe7, 0xe0, 0x2e, 0xbc, 0x6b, 0xc1, 0xca, 0x64, 0x16, 0x55, 0xf7, 0x91, 0x1b, 0x23, 0xf3, 0x2f, 0xf8, 0x45, 0x9b, 0x2, 0xd2, 0xed, 0xb3, 0x8f, 0xc3, 0x36, 0x37, 0x83, 0xe8, 0x88, 0xfe}

func main() {
	jsPath := flag.String("f", "", "file, 运行xexpect js文件")
	jsCode := flag.String("c", "", "code, 运行xexpect js代码")
	enPath := flag.String("e", "", "encrypt, 加密xexpect js文件")
	flag.Parse()

	code := []byte{}
	if *jsPath != "" {
		code = readFileCode(*jsPath)
	} else if *jsCode != "" {
		code = []byte(*jsCode)
	} else if *enPath != "" {
		encryptFile(*enPath)
		return
	} else {
		code = readFromStdin()
	}

	js := NewJS()
	js.Run(string(code))
}

func readFileCode(jsPath string) (b []byte) {
	switch getExt(jsPath) {
	case EXT_js:
		b = readFromFile(jsPath)
	case EXT_xexp:
		b = readFromFile(jsPath)
		tmp, err := xaes.NewAES().Decrypt(KEY, b)
		if err != nil {
			errorf("Decrypt error: %v", err)
		}
		b = tmp
	default:
		errorf("xexpect js file ext error: want %s or %s", EXT_js, EXT_xexp)
	}
	return b
}

func readFromFile(jsPath string) []byte {
	b, err := os.ReadFile(jsPath)
	if err != nil {
		errorf("read js file error: %v", err)
	}

	return b
}

func encryptFile(jsPath string) {
	b, err := xaes.NewAES().Encrypt(KEY, readFromFile(jsPath))
	if err != nil {
		errorf("encrypt error: %v", err)
	}

	toFile := jsPath + EXT_xexp
	err = os.WriteFile(toFile, b, 0o644)
	if err != nil {
		errorf("encrypt WriteFile error: %v", err)
	}
	fmt.Println("encrypt to file:", toFile)
}

func getExt(jsPath string) string {
	return filepath.Ext(jsPath)
}

func readFromStdin() []byte {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		errorf("read stdin error: %v", err)
	}

	return b
}

func errorf(format string, vals ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", vals...)
	os.Exit(1)
}
