package main

import (
	"SR05_projet/protocol"
	"fmt"
)

func main() {
	m := make(map[int][]protocol.Interval)
	fmt.Println(m)
	//protocol.UpdateInterval(m, 0, 1)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 2)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 3)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 4)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 8)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 10)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 9)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 10)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 30)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 0)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 6)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 5)
	fmt.Println(m)
	protocol.UpdateInterval(m, 0, 1)
	fmt.Println(m)
}
