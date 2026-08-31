package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("test_embed.exe")
	if err != nil {
		fmt.Println("err read:", err)
		return
	}
	fmt.Println("file size:", len(data))
	if len(data) < 4 {
		fmt.Println("too small")
		return
	}
	last4 := data[len(data)-4:]
	fmt.Printf("last 4 bytes: %x\n", last4)
	scriptLen := int(binary.LittleEndian.Uint32(last4))
	fmt.Println("script len:", scriptLen)
	if scriptLen > 0 && scriptLen < len(data)-4 {
		scriptBytes := data[len(data)-4-scriptLen : len(data)-4]
		fmt.Printf("script: %q\n", string(scriptBytes))
	}
}
