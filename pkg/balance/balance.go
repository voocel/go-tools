package balance

type Balance interface {
	// DoBalance 负载均衡算法
	DoBalance([]*Instance, ...string) (*Instance, error)
}
