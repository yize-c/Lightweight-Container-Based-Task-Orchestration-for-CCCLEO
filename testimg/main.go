package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("hello from test job, args:", os.Args[1:])
	time.Sleep(500 * time.Millisecond)
	fmt.Println("done")
}
