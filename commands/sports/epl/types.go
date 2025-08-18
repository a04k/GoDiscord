package epl

// FPL Data Structures
type FPLFixture struct {
	ID          int    `json:"id"`
	TeamH       int    `json:"team_h"`
	TeamA       int    `json:"team_a"`
	TeamHScore  *int   `json:"team_h_score"`
	TeamAScore  *int   `json:"team_a_score"`
	KickoffTime string `json:"kickoff_time"`
	Event       int    `json:"event"`
	Finished    bool   `json:"finished"`
}

type FPLEvent struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Finished     bool   `json:"finished"`
	DeadlineTime string `json:"deadline_time"`
}

type FPLTeam struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type FPLBootstrapStatic struct {
	Teams  []FPLTeam  `json:"teams"`
	Events []FPLEvent `json:"events"`
}

type FPLBootstrap struct {
	Teams []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		ShortName       string `json:"short_name"`
		Strength        int    `json:"strength"`
		Position        int    `json:"-"`
		Played          int    `json:"played"`
		Win             int    `json:"win"`
		Draw            int    `json:"draw"`
		Loss            int    `json:"loss"`
		Points          int    `json:"points"`
		GoalsFor        int    `json:"goals_for"`
		GoalsAgainst    int    `json:"goals_against"`
		GoalDifference  int    `json:"goal_difference"`
	} `json:"teams"`
}