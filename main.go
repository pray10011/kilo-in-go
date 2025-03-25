package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"io"
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

	newTermios.Lflag &= ^(uint32(syscall.ECHO | syscall.ICANON))

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
	var c []byte = make([]byte, 1)
	for {
		n, err := os.Stdin.Read(c)
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF")
				break
			} else {
				fmt.Println("Error:", err)
				break
			}
		}
		if c[0] == 'q' {
			fmt.Println("quit")
			break
		}
		fmt.Printf("read %d byte: [%q]\n", n, c[0])
	}
	fmt.Println("done, byebye")
}
