package stats

import (
	"fmt"
	"time"
)

// Stats
type Stats struct {
	numError      int
	numSuccessful int
	started       time.Time
}

func NewStats() Stats {
	return Stats{0, 0, time.Now()}
}

// Add success to stats
func (s *Stats) AddSuccess() { s.numSuccessful++ }

// Add  error to stats
func (s *Stats) AddError() { s.numError++ }

// Prints the current stats
func (s *Stats) Print() {
	total := s.numSuccessful + s.numError
	secsPassed := time.Now().Sub(s.started).Seconds()

	fmt.Printf("\n=======================================\n")
	fmt.Printf("Total: %v\n", total)
	fmt.Printf("Successful: %v\n", s.numSuccessful)
	fmt.Printf("Error: %v\n", s.numError)
	fmt.Printf("%% Successful: %v\n", float64(s.numSuccessful)/float64(total))
	fmt.Printf("TPS: %v\n", float64(s.numSuccessful)/secsPassed)
	fmt.Printf("=======================================\n\n")
}
