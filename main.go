package main

import (
	"fmt"
	"syscall"
	"unicode"
	"unsafe"

	"os"
)

var oldTermios syscall.Termios

func enableRawMode() {
	// 获取当前终端属性
	fd := int(os.Stdin.Fd())
	var newTermios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldTermios)))
	if errno != 0 {
		fmt.Println("get termios error:", errno)
		return
	}

	newTermios = oldTermios

	newTermios.Lflag &^= uint32(syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG)
	newTermios.Iflag &^= uint32(syscall.IXON | syscall.ICRNL)

	// 写入新的终端属性
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newTermios)))
	if errno != 0 {
		fmt.Println("set termios error:", errno)
		return
	}
}

func disableRawMode() {
	fd := int(os.Stdin.Fd())
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldTermios)))
	if errno != 0 {
		fmt.Println("set termios error:", errno)
		return
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	for {
		var buf [1]byte
		n, _, _ := syscall.Syscall(syscall.SYS_READ, uintptr(os.Stdin.Fd()), uintptr(unsafe.Pointer(&buf[0])), 1)
		if n != 1 || buf[0] == 'q' {
			break
		}

		// 判断是否为控制字符
		if unicode.IsControl(rune(buf[0])) {
			fmt.Printf("%d\n", buf[0])
		} else {
			fmt.Printf("%d (%q)\n", buf[0], buf[0])
		}
	}
}
