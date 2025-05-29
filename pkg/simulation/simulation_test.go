package simulation

import "testing"

// PlayMatch negatif gol üretmemeli diye basit birim testi
func TestPlayMatch_NonNegativeGoals(t *testing.T) {
  sim := NewSimpleSimulator()
  home, away := sim.PlayMatch(50, 50)  // Güçler eşit
  if home < 0 || away < 0 {
    t.Errorf("Gol sayısı negatif olamaz: aldık %d - %d", home, away)
  }
}
