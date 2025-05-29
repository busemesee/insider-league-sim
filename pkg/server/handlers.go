package server

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "fmt"
    "strings"

    "github.com/gorilla/mux"
    "github.com/yourusername/insider-league-simulation/pkg/models"
    "github.com/yourusername/insider-league-simulation/pkg/predict"
    "github.com/yourusername/insider-league-simulation/pkg/simulation"
)

// AppContext holds shared resources
type AppContext struct {
    DB        *sql.DB
    Simulator simulation.Simulator
}

func (app *AppContext) RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/teams", app.CreateTeam).Methods("POST")
    r.HandleFunc("/teams", app.GetStandings).Methods("GET")
    r.HandleFunc("/matches", app.GetMatchesByWeek).Methods("GET")
    r.HandleFunc("/playweek", app.PlayWeek).Methods("POST")
    r.HandleFunc("/playall", app.PlayAll).Methods("POST")
    r.HandleFunc("/edit-result", app.EditResult).Methods("PUT")
    r.HandleFunc("/predict", app.GetPredictions).Methods("GET")
}

// CreateTeam handles creating a new team
func (app *AppContext) CreateTeam(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name     string `json:"name"`
        Strength int    `json:"strength"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    var id int
    err := app.DB.QueryRow(
        `INSERT INTO teams(name,strength) VALUES($1,$2) RETURNING id`,
        input.Name, input.Strength,
    ).Scan(&id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    _, _ = app.DB.Exec(`INSERT INTO standings(team_id) VALUES($1)`, id)
    w.WriteHeader(http.StatusCreated)
}

// GetStandings returns the current league standings
func (app *AppContext) GetStandings(w http.ResponseWriter, r *http.Request) {
    rows, err := app.DB.Query(`SELECT s.team_id, t.name AS team_name,
        s.played, s.wins, s.draws, s.losses,
        s.goals_for, s.goals_against, s.goal_diff, s.points
        FROM standings s
        JOIN teams t ON t.id = s.team_id
        ORDER BY s.points DESC, s.goal_diff DESC`)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()

    var st []struct {
        TeamID       int    `json:"team_id"`
        TeamName     string `json:"team_name"`
        Played       int    `json:"played"`
        Wins         int    `json:"wins"`
        Draws        int    `json:"draws"`
        Losses       int    `json:"losses"`
        GoalsFor     int    `json:"goals_for"`
        GoalsAgainst int    `json:"goals_against"`
        GoalDiff     int    `json:"goal_diff"`
        Points       int    `json:"points"`
    }
    for rows.Next() {
        var s struct {
            TeamID       int    `json:"team_id"`
            TeamName     string `json:"team_name"`
            Played       int    `json:"played"`
            Wins         int    `json:"wins"`
            Draws        int    `json:"draws"`
            Losses       int    `json:"losses"`
            GoalsFor     int    `json:"goals_for"`
            GoalsAgainst int    `json:"goals_against"`
            GoalDiff     int    `json:"goal_diff"`
            Points       int    `json:"points"`
        }
        rows.Scan(&s.TeamID, &s.TeamName, &s.Played, &s.Wins, &s.Draws,
                  &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDiff, &s.Points)
        st = append(st, s)
    }
    json.NewEncoder(w).Encode(st)
}

// GetMatchesByWeek returns matches for a given week
func (app *AppContext) GetMatchesByWeek(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("week")
    week, _ := strconv.Atoi(q)
    rows, err := app.DB.Query(`SELECT id, week, home_team_id, away_team_id, home_goals, away_goals
        FROM matches WHERE week=$1`, week)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    defer rows.Close()
    var ms []models.Match
    for rows.Next() {
        var m models.Match
        rows.Scan(&m.ID, &m.Week, &m.HomeTeamID, &m.AwayTeamID, &m.HomeGoals, &m.AwayGoals)
        ms = append(ms, m)
    }
    json.NewEncoder(w).Encode(ms)
}

// PlayWeek simulates matches for a specific week
func (app *AppContext) PlayWeek(w http.ResponseWriter, r *http.Request) {
    var req struct{ Week int }
    json.NewDecoder(r.Body).Decode(&req)
    tx, _ := app.DB.Begin()
    defer tx.Rollback()

    teams := []models.Team{}
    rows, _ := tx.Query(`SELECT id, strength FROM teams`)
    for rows.Next() {
        var t models.Team
        rows.Scan(&t.ID, &t.Strength)
        teams = append(teams, t)
    }
    for i := 0; i < len(teams); i++ {
        for j := i + 1; j < len(teams); j++ {
            h, a := app.Simulator.PlayMatch(teams[i].Strength, teams[j].Strength)
            _, _ = tx.Exec(`INSERT INTO matches(week, home_team_id, away_team_id, home_goals, away_goals)
                VALUES($1,$2,$3,$4,$5)`, req.Week, teams[i].ID, teams[j].ID, h, a)
            updateStandings(tx, teams[i].ID, teams[j].ID, h, a)
        }
    }
    tx.Commit()
    w.WriteHeader(http.StatusNoContent)
}

// PlayAll simulates all weeks
func (app *AppContext) PlayAll(w http.ResponseWriter, r *http.Request) {
    for wk := 1; wk <= 3; wk++ {
        _, _ = http.Post("http://localhost:8080/playweek", "application/json",
            strings.NewReader(fmt.Sprintf(`{"week": %d}`, wk)))
    }
    w.WriteHeader(http.StatusNoContent)
}
// EditResult updates a match result
func (app *AppContext) EditResult(w http.ResponseWriter, r *http.Request) {
    var req struct {
        MatchID   int `json:"match_id"`
        HomeGoals int `json:"home_goals"`
        AwayGoals int `json:"away_goals"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    tx, _ := app.DB.Begin()
    defer tx.Rollback()

    var old struct {
        HomeID int
        AwayID int
        HG     int
        AG     int
    }
    tx.QueryRow(`SELECT home_team_id, away_team_id, home_goals, away_goals
        FROM matches WHERE id=$1`, req.MatchID).
        Scan(&old.HomeID, &old.AwayID, &old.HG, &old.AG)

    revertStandings(tx, old.HomeID, old.AwayID, old.HG, old.AG)
    _, _ = tx.Exec(`UPDATE matches SET home_goals=$1, away_goals=$2 WHERE id=$3`,
        req.HomeGoals, req.AwayGoals, req.MatchID)
    updateStandings(tx, old.HomeID, old.AwayID, req.HomeGoals, req.AwayGoals)
    tx.Commit()
    w.WriteHeader(http.StatusNoContent)
}

// GetPredictions returns team probabilities
func (app *AppContext) GetPredictions(w http.ResponseWriter, r *http.Request) {
    preds, err := predict.CalculateSimplePredictions(app.DB)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Enrich with team names
    var output []struct {
        TeamID      int     `json:"team_id"`
        TeamName    string  `json:"team_name"`
        Probability float64 `json:"probability"`
    }
    for _, pr := range preds {
        var name string
        app.DB.QueryRow(`SELECT name FROM teams WHERE id=$1`, pr.TeamID).Scan(&name)
        output = append(output, struct {
            TeamID      int     `json:"team_id"`
            TeamName    string  `json:"team_name"`
            Probability float64 `json:"probability"`
        }{pr.TeamID, name, pr.Probability})
    }
    json.NewEncoder(w).Encode(output)
}

func updateStandings(tx *sql.Tx, homeID, awayID, hg, ag int) {
    tx.Exec(`INSERT INTO standings(team_id) VALUES($1) ON CONFLICT DO NOTHING`, homeID)
    tx.Exec(`INSERT INTO standings(team_id) VALUES($1) ON CONFLICT DO NOTHING`, awayID)
    tx.Exec(`UPDATE standings SET played = played + 1 WHERE team_id = $1`, homeID)
    tx.Exec(`UPDATE standings SET played = played + 1 WHERE team_id = $1`, awayID)
    tx.Exec(`UPDATE standings SET goals_for = goals_for + $1, goals_against = goals_against + $2, goal_diff = goal_diff + ($1-$2) WHERE team_id = $3`, hg, ag, homeID)
    tx.Exec(`UPDATE standings SET goals_for = goals_for + $1, goals_against = goals_against + $2, goal_diff = goal_diff + ($1-$2) WHERE team_id = $3`, ag, hg, awayID)
    if hg > ag {
        tx.Exec(`UPDATE standings SET wins = wins + 1, points = points + 3 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET losses = losses + 1 WHERE team_id = $1`, awayID)
    } else if hg == ag {
        tx.Exec(`UPDATE standings SET draws = draws + 1, points = points + 1 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET draws = draws + 1, points = points + 1 WHERE team_id = $1`, awayID)
    } else {
        tx.Exec(`UPDATE standings SET losses = losses + 1 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET wins = wins + 1, points = points + 3 WHERE team_id = $1`, awayID)
    }
}

func revertStandings(tx *sql.Tx, homeID, awayID, hg, ag int) {
    tx.Exec(`UPDATE standings SET played = played - 1 WHERE team_id = $1`, homeID)
    tx.Exec(`UPDATE standings SET played = played - 1 WHERE team_id = $1`, awayID)
    tx.Exec(`UPDATE standings SET goals_for = goals_for - $1, goals_against = goals_against - $2, goal_diff = goal_diff - ($1-$2) WHERE team_id = $3`, hg, ag, homeID)
    tx.Exec(`UPDATE standings SET goals_for = goals_for - $1, goals_against = goals_against - $2, goal_diff = goal_diff - ($1-$2) WHERE team_id = $3`, ag, hg, awayID)
    if hg > ag {
        tx.Exec(`UPDATE standings SET wins = wins - 1, points = points - 3 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET losses = losses - 1 WHERE team_id = $1`, awayID)
    } else if hg == ag {
        tx.Exec(`UPDATE standings SET draws = draws - 1, points = points - 1 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET draws = draws - 1, points = points - 1 WHERE team_id = $1`, awayID)
    } else {
        tx.Exec(`UPDATE standings SET losses = losses - 1 WHERE team_id = $1`, homeID)
        tx.Exec(`UPDATE standings SET wins = wins - 1, points = points - 3 WHERE team_id = $1`, awayID)
    }
}
