package silver

// KidProfile represents basic kid profile information
type KidProfile struct {
	ProfileID    string // UUID
	FullName     string
	Nickname     string
	Age          int
	DateOfBirth  string
	TotalBalance float64 // Optional, used by transformer_v2
}
