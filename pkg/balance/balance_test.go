package balance

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	var insts []*Instance
	for i := 0; i < 10; i++ {
		host := fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255))
		port, _ := strconv.Atoi(fmt.Sprintf("880%d", i))
		one := NewInstance(host, port)
		insts = append(insts, one)
	}
	var name = "round"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	for {
		inst, err := DoBalance(name, insts)
		if err != nil {
			fmt.Println("do balance err")
			time.Sleep(time.Second)
			continue
		}
		fmt.Println(inst)
		time.Sleep(time.Second)
	}
}
