package wm

const (
	spreadChipsPerSymbol = 1
	spreadTargetScale    = 0.70
)

func shuffledSlotsAndChips(key string, totalSlots, neededSlots int) (slots []int, chips []int8) {
	if totalSlots <= 0 || neededSlots <= 0 {
		return nil, nil
	}
	if neededSlots > totalSlots {
		neededSlots = totalSlots
	}

	rng := NewPRNG(SeedFromKey("spread-v1:" + key))

	order := make([]int, totalSlots)
	for i := 0; i < totalSlots; i++ {
		order[i] = i
	}

	for i := totalSlots - 1; i > 0; i-- {
		j := int(rng.NextU64() % uint64(i+1))
		order[i], order[j] = order[j], order[i]
	}

	chips = make([]int8, neededSlots)
	for i := range chips {
		if rng.NextPM1() < 0 {
			chips[i] = -1
		} else {
			chips[i] = 1
		}
	}

	return order[:neededSlots], chips
}
