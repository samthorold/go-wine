package domain

// Rating is a Tasting's 5-point measure of absolute enjoyment. It is the
// deterministic weight the Taste profile math depends on, so it is a value
// object that cannot exist outside 1..5.
type Rating int

// NewRating constructs a Rating, rejecting anything outside 1..5.
func NewRating(v int) (Rating, error) {
	if v < 1 || v > 5 {
		return 0, ErrInvalidRating
	}
	return Rating(v), nil
}

func (r Rating) Int() int { return int(r) }
