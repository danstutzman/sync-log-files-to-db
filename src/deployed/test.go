package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "hello world to stderr")
	fmt.Printf("HELLO FROM GOLANG WITH ARGS %v\n", os.Args)
}
