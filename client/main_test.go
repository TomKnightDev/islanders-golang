package main

import (
	"testing"

	"golang.org/x/image/math/f64"
)

func TestGetDistance(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		a := f64.Vec2{3, 2}
		b := f64.Vec2{9, 7}

		got := GetDistance(a, b)
		want := 7.810249675906654

		if got != want {
			t.Errorf("got %v, wanted %v", got, want)
		}
	})
}
