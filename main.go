package main

import (
	"fmt"
	"syscall"
	"unicode"
	"unsafe"

	"os"
)

/* define */
var oldTermios syscall.Termios
// 0x1f = 00011111，即清除5、6位，变为控制字符
func CTRL_KEY(b byte) byte {
	return b & 0x1f
}

func die(s string) {
	fmt.Fprintf(os.Stderr, "%s\r\n",s)
	os.Exit(1)
}

func enableRawMode() {
	// 获取当前终端属性
	fd := int(os.Stdin.Fd())
	var newTermios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldTermios)))
	if errno != 0 {
		die("tcgetattr error")
	}

	newTermios = oldTermios

	newTermios.Lflag &^= uint32(syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG)
	newTermios.Iflag &^= uint32(syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP)
	newTermios.Oflag &^= uint32(syscall.OPOST)
	newTermios.Cflag |= syscall.CS8
	newTermios.Cc[syscall.VMIN] = 0
	newTermios.Cc[syscall.VTIME] = 1

	// 写入新的终端属性
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newTermios)))
	if errno != 0 {
		die("tcsetattr error")
	}
}

func disableRawMode() {
	fd := int(os.Stdin.Fd())
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldTermios)))
	if errno != 0 {
		die("tcsetattr error")
	}
}

/* init */
func main() {
	enableRawMode()
	defer disableRawMode()
	
	for {
		// _, err := os.Stdin.Read(buf)
		// if err != nil {
		// 	die("read error")
		// }
		var buf = make([]byte, 1)
		_,_,errno:=syscall.Syscall(syscall.SYS_READ, uintptr(os.Stdin.Fd()), uintptr(unsafe.Pointer(&buf[0])), 1)
		if errno != 0 {
			die("read error")
		}
		if buf[0] == CTRL_KEY('q') {
			break
		}

		// 判断是否为控制字符
		if unicode.IsControl(rune(buf[0])) {
			fmt.Printf("%d\r\n", buf[0])
		} else {
			fmt.Printf("%d (%q)\r\n", buf[0], buf[0])
		}
	}
}
