-- Teams table
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    strength INT NOT NULL
);

-- Matches: weekly match results
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    week INT NOT NULL,
    home_team_id INT NOT NULL REFERENCES teams(id),
    away_team_id INT NOT NULL REFERENCES teams(id),
    home_goals INT NOT NULL,
    away_goals INT NOT NULL,
    UNIQUE(week, home_team_id, away_team_id)
);

-- Standings caching (optional)
CREATE TABLE standings (
    team_id INT PRIMARY KEY REFERENCES teams(id),
    played INT NOT NULL DEFAULT 0,
    wins INT NOT NULL DEFAULT 0,
    draws INT NOT NULL DEFAULT 0,
    losses INT NOT NULL DEFAULT 0,
    goals_for INT NOT NULL DEFAULT 0,
    goals_against INT NOT NULL DEFAULT 0,
    goal_diff INT NOT NULL DEFAULT 0,
    points INT NOT NULL DEFAULT 0
);