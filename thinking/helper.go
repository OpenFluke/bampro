package main

import (
	"io"
	"math"
	"os"
	"sort"
)

func Mean(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	return sum / float64(len(nums))
}

func Min(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	min := nums[0]
	for _, v := range nums {
		if v < min {
			min = v
		}
	}
	return min
}

func Max(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, v := range nums {
		if v > max {
			max = v
		}
	}
	return max
}

func Median(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sorted := append([]float64{}, nums...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func distance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}

func findPlanetCenter(pos []float64, rawPlanets []string) []float64 {
	const spacing = 800.0
	minDist := math.MaxFloat64
	var closest []float64

	for _, p := range rawPlanets {
		vec, err := parseVec3(p)
		if err != nil {
			continue
		}
		center := []float64{vec.X * spacing, vec.Y * spacing, vec.Z * spacing}
		d := distance(pos, center)
		if d < minDist {
			minDist = d
			closest = center
		}
	}
	return closest
}

func copyFile(src, dst string) error {
	tmp := dst + ".tmp"
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	out, err := os.Create(tmp)
	if err != nil {
		in.Close()
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		in.Close()
		return err
	}
	out.Close()
	in.Close()
	return os.Rename(tmp, dst) // atomic on POSIX
}
