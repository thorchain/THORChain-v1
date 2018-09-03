package stats

import (
	"fmt"
	"time"
)

// Stats
type Stats struct {
	numAccountNotFound      int
	numNoCoinsToSend        int
	numOtherError           int
	numResourceNotAvailable int
	numSuccessful           int
	started                 time.Time
}

func NewStats() Stats {
	return Stats{0, 0, 0, 0, 0, time.Now()}
}

// Add success to stats
func (s *Stats) AddSuccess() { s.numSuccessful++ }

// Add account not found error to stats
func (s *Stats) AddAccountNotFound() { s.numAccountNotFound++ }

// Add no coins to send error to stats
func (s *Stats) AddNoCoinsToSend() { s.numNoCoinsToSend++ }

// Add other error to stats
func (s *Stats) AddOtherError() { s.numOtherError++ }

// Add resource not available error to stats
func (s *Stats) AddResourceNotAvailable() { s.numResourceNotAvailable++ }

// Prints the current stats
func (s *Stats) Print() {
	numUnsuccessful := s.numAccountNotFound + s.numNoCoinsToSend + s.numOtherError + s.numResourceNotAvailable
	total := s.numSuccessful + numUnsuccessful
	secsPassed := time.Now().Sub(s.started).Seconds()

	fmt.Printf("\n=======================================\n")
	fmt.Printf("Total: %v\n", total)
	fmt.Printf("Successful: %v\n", s.numSuccessful)
	fmt.Printf("%% Successful: %v\n", float64(s.numSuccessful)/float64(total))
	fmt.Printf("TPS: %v\n", float64(s.numSuccessful)/secsPassed)
	fmt.Printf("Account not found: %v\n", s.numAccountNotFound)
	fmt.Printf("No coins to send: %v\n", s.numNoCoinsToSend)
	fmt.Printf("Resource not available: %v\n", s.numResourceNotAvailable)
	fmt.Printf("Other error: %v\n", s.numOtherError)
	fmt.Printf("=======================================\n\n")
}
