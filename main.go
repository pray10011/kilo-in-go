package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
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
		fmt.Printf("read %d byte: [%q]\n", n, c[0])
	}
	fmt.Println("done, byebye")
}
