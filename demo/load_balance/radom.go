package load_balance

type LoadBalanceType int

const (
	LBRandom LoadBalanceType = iota
	LBRoundRobin
	LBWeightRoundRobin
	LBConsistentHash
)
