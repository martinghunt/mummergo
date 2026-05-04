package mummergo

// Interval represents a 0-based half-open coordinate range [Start, End).
// Start is included, End is excluded.
type Interval struct {
	Start int
	End   int
}

func NewInterval(start, end int) Interval {
	if start > end {
		start, end = end, start
	}
	return Interval{Start: start, End: end}
}

func NewIntervalFromOriented(start, end int) Interval {
	if start <= end {
		return Interval{Start: start, End: end}
	}
	return Interval{Start: end + 1, End: start + 1}
}

func (i Interval) Len() int {
	if i.End <= i.Start {
		return 0
	}
	return i.End - i.Start
}

func (i Interval) Empty() bool {
	return i.Len() == 0
}

func (i Interval) Contains(point int) bool {
	return i.Start <= point && point < i.End
}

func (i Interval) Intersects(other Interval) bool {
	if i.Empty() || other.Empty() {
		return false
	}
	return i.Start < other.End && other.Start < i.End
}

func (i Interval) DistanceToPoint(point int) int {
	if i.Contains(point) {
		return 0
	}
	if point < i.Start {
		return i.Start - point
	}
	return point - i.End + 1
}
