package wm

import "testing"

func TestSeedFromKeyKnownVectors(t *testing.T) {
	tests := []struct {
		key  string
		want uint64
	}{
		{key: "", want: 14695981039346656037},
		{key: "a", want: 12638187200555641996},
		{key: "abc", want: 16654208175385433931},
		{key: "hello", want: 11831194018420276491},
	}

	for _, tt := range tests {
		got := SeedFromKey(tt.key)
		if got != tt.want {
			t.Fatalf("SeedFromKey(%q)=%d want=%d", tt.key, got, tt.want)
		}
	}
}

func TestPRNGDeterministicSequence(t *testing.T) {
	seed := SeedFromKey("abc")
	rng := NewPRNG(seed)

	want := []uint64{
		17965033732531672793,
		13978924921052248732,
		11023063779569587599,
		4736322939913124213,
		6047635169297162149,
	}

	for i, expected := range want {
		got := rng.NextU64()
		if got != expected {
			t.Fatalf("sequence mismatch at %d: got=%d want=%d", i, got, expected)
		}
	}
}

func TestPRNGSameSeedSameOutput(t *testing.T) {
	seed := SeedFromKey("same")
	a := NewPRNG(seed)
	b := NewPRNG(seed)

	for i := 0; i < 32; i++ {
		ga := a.NextU64()
		gb := b.NextU64()
		if ga != gb {
			t.Fatalf("sequences diverged at %d: a=%d b=%d", i, ga, gb)
		}
	}
}

func TestPRNGRanges(t *testing.T) {
	rng := NewPRNG(SeedFromKey("range"))

	for i := 0; i < 500; i++ {
		v := rng.NextF32()
		if v < 0 || v >= 1 {
			t.Fatalf("NextF32 out of range: %f", v)
		}

		p := rng.NextPM1()
		if p < -1 || p >= 1 {
			t.Fatalf("NextPM1 out of range: %f", p)
		}
	}
}
