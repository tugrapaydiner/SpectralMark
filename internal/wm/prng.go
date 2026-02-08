package wm

const (
	xorshift64StarMul      = 2685821657736338717
	xorshift64DefaultState = 0x9e3779b97f4a7c15
)

type PRNG struct {
	state uint64
}

func NewPRNG(seed uint64) *PRNG {
	if seed == 0 {
		seed = xorshift64DefaultState
	}
	return &PRNG{state: seed}
}

func (p *PRNG) NextU64() uint64 {
	if p == nil {
		return 0
	}

	x := p.state
	if x == 0 {
		x = xorshift64DefaultState
	}

	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	p.state = x

	return x * xorshift64StarMul
}

func (p *PRNG) NextF32() float32 {
	const inv24 = float32(1.0 / (1 << 24))
	return float32(p.NextU64()>>40) * inv24
}

func (p *PRNG) NextPM1() float32 {
	return p.NextF32()*2 - 1
}
