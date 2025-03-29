package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"os"
)

/*** define ***/
type editorConfig struct {
	oldTermios syscall.Termios
	ws         winsize
}

var E editorConfig

type winsize struct {
	Row uint16
	Col uint16
	X   uint16
	Y   uint16
}

// 0x1f = 00011111，即清除5、6位，变为控制字符
func CTRL_KEY(b byte) byte {
	return b & 0x1f
}

/*** terminal ***/

func die(s string) {
	var clearScreen = []byte("\x1b[2J")
	syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&clearScreen[0])), 4)
	var cursorLeftUp = []byte("\x1b[H")
	syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorLeftUp[0])), 3)

	fmt.Fprintf(os.Stderr, "%s\r\n", s)

	disableRawMode()
	os.Exit(1)
}

func enableRawMode() {
	// 获取当前终端属性
	var newTermios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&E.oldTermios)))
	if errno < 0 {
		die("tcgetattr error")
	}

	newTermios = E.oldTermios

	newTermios.Lflag &^= uint32(syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG)
	newTermios.Iflag &^= uint32(syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP)
	newTermios.Oflag &^= uint32(syscall.OPOST)
	newTermios.Cflag |= syscall.CS8
	newTermios.Cc[syscall.VMIN] = 0
	newTermios.Cc[syscall.VTIME] = 1

	// 写入新的终端属性
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newTermios)))
	if errno < 0 {
		die("tcsetattr error")
	}
}

func disableRawMode() {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&E.oldTermios)))
	if errno < 0 {
		die("tcsetattr error")
	}
}

func editorReadKey() byte {
	var buf = make([]byte, 1)
	for {
		r1, _, errno := syscall.Syscall(syscall.SYS_READ, os.Stdin.Fd(), uintptr(unsafe.Pointer(&buf[0])), 1)
		if r1 == 1 {
			break
		}
		if r1 < 0 && errno != syscall.EAGAIN {
			die("read error")
		}
	}
	return buf[0]
}

func getCursorPosition() int {
	var cursorPosition = []byte("\x1b[6n")
	_, _, errno := syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorPosition[0])), 4)
	if errno != 0 {
		return -1
	}
	fmt.Printf("\r\n")
	var buf = make([]byte, 32)
	var i int
	for {
		r1, _, _ := syscall.Syscall(syscall.SYS_READ, os.Stdin.Fd(), uintptr(unsafe.Pointer(&buf[i])), 1)
		if r1 != 1 {
			break
		}
		if buf[i] == 'R' {
			break
		}
		i++
	}
	if buf[0] != '\x1b' || buf[1] != '[' {
		return -1
	}
	n, err := fmt.Sscanf(string(buf[2:i]), "%d;%d", &E.ws.Row, &E.ws.Col)
	if n != 2 || err != nil {
		return -1
	}
	return 0
}

func getWindowSize() int {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&E.ws)))
	if errno != 0 || E.ws.Row == 0 || E.ws.Col == 0 {
		var cursorRightDown = []byte("\x1b[999C\x1b[999B")
		_, _, errno = syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorRightDown[0])), 12)
		if errno != 0 {
			return -1
		}
		return getCursorPosition()
	}
	return 0
}

/*** output ***/
func editorDrawRows() {
	var tlide = []byte("~")
	var newline = []byte("\r\n")
	for i := 0; i < int(E.ws.Row); i++ {
		syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&tlide[0])), 1)
		if i < int(E.ws.Row)-1 {
			syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&newline[0])), 2)
		}
	}
}

func editorRefreshScreen() {
	var clearScreen = []byte("\x1b[2J")
	// 向标准输出写入4个字节实现清屏。第一个字节为\x1b，表示ESC，第二个字节为[，第三个字节为2，第四个字节为j
	syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&clearScreen[0])), 4)
	var cursorLeftUp = []byte("\x1b[H")
	syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorLeftUp[0])), 3)

	editorDrawRows()
	syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorLeftUp[0])), 3)
}

/*** input ***/
func editorProcessKeyPress() {
	var c = editorReadKey()
	switch c {
	case CTRL_KEY('q'):
		var clearScreen = []byte("\x1b[2J")
		syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&clearScreen[0])), 4)
		var cursorLeftUp = []byte("\x1b[H")
		syscall.Syscall(syscall.SYS_WRITE, os.Stdout.Fd(), uintptr(unsafe.Pointer(&cursorLeftUp[0])), 3)

		disableRawMode()
		// os.Exit(0)之后main函数的defer不执行，所以在这里显式调用
		os.Exit(0)
		break
	}
}

/*** init ***/
func initEditor() {
	if getWindowSize() == -1 {
		die("getWindowSize error")
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	initEditor()

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
