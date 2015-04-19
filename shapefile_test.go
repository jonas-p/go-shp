package shp

import "testing"

func TestBoxExtend(t *testing.T) {
	a := Box{-124.763068, 45.543541, -116.915989, 49.002494}
	b := Box{-92.888114, 42.49192, -86.805415, 47.080621}
	a.Extend(b)
	c := Box{-124.763068, 42.49192, -86.805415, 49.002494}
	if a.MinX != c.MinX {
		t.Errorf("a.MinX = %v, want %v", a.MinX, c.MinX)
	}
	if a.MinY != c.MinY {
		t.Errorf("a.MinY = %v, want %v", a.MinY, c.MinY)
	}
	if a.MaxX != c.MaxX {
		t.Errorf("a.MaxX = %v, want %v", a.MaxX, c.MaxX)
	}
	if a.MaxY != c.MaxY {
		t.Errorf("a.MaxY = %v, want %v", a.MaxY, c.MaxY)
	}
}
