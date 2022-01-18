package mk48

import "encoding/json"

type CreateSession struct {
	GameID string `json:"game_id"`
}

func (CreateSession) Name() string { return "CreateSession" }

type SessionCreated struct {
	ArenaID   int         `json:"arena_id"`
	ServerID  int         `json:"server_id"`
	SessionID json.Number `json:"session_id"`
	PlayerID  json.Number `json:"player_id"`
}

type LeaderboardEntry struct {
	Player string `json:"alias"`
	Score  int    `json:"score"`
}

type LeaderboardPeriod string

const (
	AllTimeLeaderboard    LeaderboardPeriod = "AllTime"
	WeeklyTimeLeaderboard LeaderboardPeriod = "Weekly"
	DailyTimeLeaderboard  LeaderboardPeriod = "Daily"
)

type LeaderboardUpdate struct {
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	Period      LeaderboardPeriod  `json:"period"`
}

func (LeaderboardUpdate) Name() string { return "LeaderboardUpdate" }

type Message interface {
	Name() string
}

type Event struct {
	SessionCreated    *SessionCreated
	LeaderboardUpdate *LeaderboardUpdate `json:"LeaderboardUpdated"`
}
