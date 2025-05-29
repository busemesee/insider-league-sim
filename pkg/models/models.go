package models

type Team struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Strength int    `json:"strength"`
}

type Match struct {
    ID         int `json:"id"`
    Week       int `json:"week"`
    HomeTeamID int `json:"home_team_id"`
    AwayTeamID int `json:"away_team_id"`
    HomeGoals  int `json:"home_goals"`
    AwayGoals  int `json:"away_goals"`
}

type Standing struct {
    TeamID       int `json:"team_id"`
    Played       int `json:"played"`
    Wins         int `json:"wins"`
    Draws        int `json:"draws"`
    Losses       int `json:"losses"`
    GoalsFor     int `json:"goals_for"`
    GoalsAgainst int `json:"goals_against"`
    GoalDiff     int `json:"goal_diff"`
    Points       int `json:"points"`
}
