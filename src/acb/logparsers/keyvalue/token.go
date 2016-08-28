package keyvalue

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT
	STRING

	// Misc characters
	EQUAL // =
)
