package main

import (
	"fmt"
	"strconv"
)

func main() {
	var foo string
	for i := range 10 {
		foo += strconv.Itoa(i)
	}
	fmt.Println(foo)
}
