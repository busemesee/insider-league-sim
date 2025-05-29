package predict

import (
    "database/sql"
)

// Prediction holds probability data for each team
type Prediction struct {
    TeamID      int     `json:"team_id"`
    Probability float64 `json:"probability"`
}

// CalculateSimplePredictions returns percentage-based predictions
// based solely on current points in the standings.
func CalculateSimplePredictions(db *sql.DB) ([]Prediction, error) {
    rows, err := db.Query(`SELECT team_id, points FROM standings`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    data := make(map[int]int)
    total := 0
    for rows.Next() {
        var id, pts int
        rows.Scan(&id, &pts)
        data[id] = pts
        total += pts
    }

    var preds []Prediction
    for id, pts := range data {
        prob := (float64(pts) / float64(total)) * 100
        preds = append(preds, Prediction{TeamID: id, Probability: prob})
    }
    return preds, nil
}
