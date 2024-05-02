package main

// Formulas used are from https://datatracker.ietf.org/doc/html/rfc6298

import "math"

const (
	ALPHA float64 = 0.125
	BETA  float64 = 0.25
)

type MovingAverage interface {
	Add(float64)
	Value() float64
	Reset()
}

func NewMovingAverage() MovingAverage {
	return new(SimpleEWMA)
}

type SimpleEWMA struct {
	estimated float64
	deviation float64
}

func (e *SimpleEWMA) Add(value float64) {
	if e.estimated == 0 {
		e.estimated = value
		e.deviation = value / 2
	} else {
		e.estimated = (value * ALPHA) + (e.estimated * (1 - ALPHA))
		e.deviation = (math.Abs(value-e.estimated) * BETA) + (e.deviation * (1 - BETA))
	}
}

func (e *SimpleEWMA) Value() float64 {
	return 4*e.deviation + e.estimated
}

func (e *SimpleEWMA) Reset() {
	e.estimated = 0
	e.deviation = 0
}
