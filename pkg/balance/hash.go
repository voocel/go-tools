package balance

import (
	"fmt"
	"hash/crc32"
	"math/rand"
)

func init() {
	RegisterBalance("hash", &HashBalance{})
}

type HashBalance struct {
	key string
}

func (p *HashBalance) DoBalance(insts []*Instance, key ...string) (inst *Instance, err error) {
	defKey := fmt.Sprintf("%d", rand.Int())
	if len(key) > 0 {
		defKey = key[0]
	}
	lens := len(insts)
	if lens == 0 {
		err = fmt.Errorf("no balance")
		return
	}
	hashVal := crc32.Checksum([]byte(defKey), crc32.MakeTable(crc32.IEEE))
	index := int(hashVal) % lens
	inst = insts[index]
	return
}
