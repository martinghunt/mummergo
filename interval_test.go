package mummergo

import "testing"

func TestIntervalUsesHalfOpenCoordinates(t *testing.T) {
	interval := Interval{Start: 2, End: 5}
	if interval.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", interval.Len())
	}
	for _, pos := range []int{2, 3, 4} {
		if !interval.Contains(pos) {
			t.Fatalf("Contains(%d) = false, want true", pos)
		}
	}
	for _, pos := range []int{1, 5} {
		if interval.Contains(pos) {
			t.Fatalf("Contains(%d) = true, want false", pos)
		}
	}
	if interval.Intersects(Interval{Start: 0, End: 2}) {
		t.Fatal("touching interval before should not intersect")
	}
	if interval.Intersects(Interval{Start: 5, End: 7}) {
		t.Fatal("touching interval after should not intersect")
	}
	if !interval.Intersects(Interval{Start: 4, End: 7}) {
		t.Fatal("overlapping interval should intersect")
	}
	if interval.Intersects(Interval{Start: 3, End: 3}) {
		t.Fatal("empty interval should not intersect")
	}
}

func TestNewIntervalFromOriented(t *testing.T) {
	tests := []struct {
		start int
		end   int
		want  Interval
	}{
		{start: 0, end: 100, want: Interval{Start: 0, End: 100}},
		{start: 99, end: -1, want: Interval{Start: 0, End: 100}},
		{start: 50, end: 7, want: Interval{Start: 8, End: 51}},
	}
	for _, test := range tests {
		if got := NewIntervalFromOriented(test.start, test.end); got != test.want {
			t.Fatalf("NewIntervalFromOriented(%d, %d) = %#v, want %#v", test.start, test.end, got, test.want)
		}
	}
}
