package balance

import (
	"fmt"
)

type BalanceMgr struct {
	allBalance map[string]Balance
}

var mgr = BalanceMgr{
	allBalance: make(map[string]Balance),
}

func (p *BalanceMgr) registerBalance(name string, b Balance) {
	p.allBalance[name] = b
}

func RegisterBalance(name string, b Balance) {
	mgr.registerBalance(name, b)
}

func DoBalance(name string, insts []*Instance) (inst *Instance, err error) {
	balance, ok := mgr.allBalance[name]
	if !ok {
		err = fmt.Errorf("not fount %s", name)
		fmt.Println("not found ", name)
		return
	}
	inst, err = balance.DoBalance(insts)
	if err != nil {
		err = fmt.Errorf(" %s erros", name)
		return
	}
	return
}
