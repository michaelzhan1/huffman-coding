package compress

const (
	AlphabetSize = 1 << 8
	MaxCodeBits  = AlphabetSize - 1
	MaxCodeBytes = (MaxCodeBits + 7) / 8
)
