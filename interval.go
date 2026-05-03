package mummergo

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

func (i Interval) Intersects(other Interval) bool {
	return i.Start <= other.End && other.Start <= i.End
}

func (i Interval) DistanceToPoint(point int) int {
	if point < i.Start {
		return i.Start - point
	}
	if point > i.End {
		return point - i.End
	}
	return 0
}
