package f1

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const ergastAPI = "https://api.jolpi.ca/ergast/f1/current"

// Event represents a single F1 event (Grand Prix) from our schedule.
type Event struct {
	Name     string    `json:"name"`
	Location string    `json:"location"`
	Circuit  string    `json:"circuit"`
	Sessions []Session `json:"sessions"`
}

// Session represents a single session within an F1 event.
type Session struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Date string `json:"date"`
}

// TeamColors maps constructor names to their Discord color codes
var TeamColors = map[string]int{
	"Red Bull":       0x1E5BC6,
	"Ferrari":        0xDC0000,
	"Mercedes":       0x00D2BE,
	"McLaren":        0xFF8700,
	"Aston Martin":   0x006F62,
	"Alpine":         0x0090FF,
	"Williams":       0x005AFF,
	"AlphaTauri":     0x2B4562,
	"Haas F1 Team":   0xFFFFFF,
	"Alfa Romeo":     0x960000,
	"RB F1 Team":     0x6600FF,
	"Sauber":         0x006400,
}

// DriverInfo maps driver names/codes to constructor names for color coding
var DriverInfo = map[string]string{
	"VER": "Red Bull", "PER": "Red Bull",
	"LEC": "Ferrari", "SAI": "Ferrari",
	"HAM": "Mercedes", "RUS": "Mercedes",
	"NOR": "McLaren", "PIA": "McLaren",
	"ALO": "Aston Martin", "STR": "Aston Martin",
	"OCO": "Alpine", "GAS": "Alpine",
	"ALB": "Williams", "SAR": "Williams",
	"TSU": "RB F1 Team", "LAW": "RB F1 Team",
	"MAG": "Haas F1 Team", "HUL": "Haas F1 Team",
	"BOT": "Alfa Romeo", "ZHO": "Alfa Romeo",
	"RIC": "RB F1 Team", "COL": "RB F1 Team",
}

// DriverStandingsResponse represents the structure of the Ergast API response for driver standings
type DriverStandingsResponse struct {
	MRData struct {
		StandingsTable struct {
			Season        string `json:"season"`
			StandingsLists []struct {
				Season         string `json:"season"`
				Round          string `json:"round"`
				DriverStandings []struct {
					Position  string `json:"position"`
					Points    string `json:"points"`
					Wins      string `json:"wins"`
					Driver    struct {
						GivenName  string `json:"givenName"`
						FamilyName string `json:"familyName"`
						Code       string `json:"code"`
					} `json:"Driver"`
				} `json:"DriverStandings"`
			} `json:"StandingsLists"`
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

// ConstructorStandingsResponse represents the structure of the Ergast API response for constructor standings
type ConstructorStandingsResponse struct {
	MRData struct {
		StandingsTable struct {
			Season        string `json:"season"`
			StandingsLists []struct {
				Season         string `json:"season"`
				Round          string `json:"round"`
				ConstructorStandings []struct {
					Position    string `json:"position"`
					Points      string `json:"points"`
					Wins        string `json:"wins"`
					Constructor struct {
						Name string `json:"name"`
					} `json:"Constructor"`
				} `json:"ConstructorStandings"`
			} `json:"StandingsLists"`
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

// QualifyingResponse represents the structure of the Ergast API response for qualifying results
type QualifyingResponse struct {
	MRData struct {
		RaceTable struct {
			Races []struct {
				RaceName string `json:"raceName"`
				Circuit  struct {
					CircuitName string `json:"circuitName"`
					Location    struct {
						Locality string `json:"locality"`
						Country  string `json:"country"`
					} `json:"Location"`
				} `json:"Circuit"`
				QualifyingResults []struct {
					Position   string `json:"position"`
					Driver     struct {
						GivenName  string `json:"givenName"`
						FamilyName string `json:"familyName"`
						Code       string `json:"code"`
					} `json:"Driver"`
					Constructor struct {
						Name string `json:"name"`
					} `json:"Constructor"`
					Q1 string `json:"Q1"`
					Q2 string `json:"Q2"`
					Q3 string `json:"Q3"`
				} `json:"QualifyingResults"`
			} `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

// RaceResultsResponse represents the structure of the Ergast API response for race results
type RaceResultsResponse struct {
	MRData struct {
		RaceTable struct {
			Races []struct {
				RaceName string `json:"raceName"`
				Circuit  struct {
					CircuitName string `json:"circuitName"`
					Location    struct {
						Locality string `json:"locality"`
						Country  string `json:"country"`
					} `json:"Location"`
				} `json:"Circuit"`
				Results []struct {
					Position     string `json:"position"`
					Points       string `json:"points"`
					Driver       struct {
						GivenName  string `json:"givenName"`
						FamilyName string `json:"familyName"`
						Code       string `json:"code"`
					} `json:"Driver"`
					Constructor struct {
						Name string `json:"name"`
					} `json:"Constructor"`
					Laps       string `json:"laps"`
					Status     string `json:"status"`
					Time       struct {
						Time   string `json:"time"`
						Millis string `json:"millis"`
					} `json:"Time"`
					FastestLap struct {
						Time struct {
							Time string `json:"time"`
						} `json:"Time"`
					} `json:"FastestLap"`
				} `json:"Results"`
			} `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

func fetchAndDecode(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	return nil
}

func FetchDriverStandings() (*DriverStandingsResponse, error) {
	var data DriverStandingsResponse
	url := fmt.Sprintf("%s/driverStandings.json", ergastAPI)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding driver standings: %v", err)
		return nil, err
	}
	return &data, nil
}

func FetchConstructorStandings() (*ConstructorStandingsResponse, error) {
	var data ConstructorStandingsResponse
	url := fmt.Sprintf("%s/constructorStandings.json", ergastAPI)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding constructor standings: %v", err)
		return nil, err
	}
	return &data, nil
}

func FetchLatestRaceResults() (*RaceResultsResponse, error) {
	var data RaceResultsResponse
	url := fmt.Sprintf("%s/last/results.json", ergastAPI)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding latest race results: %v", err)
		return nil, err
	}
	return &data, nil
}

func FetchQualifyingResults() (*QualifyingResponse, error) {
	var data QualifyingResponse
	url := fmt.Sprintf("%s/last/qualifying.json", ergastAPI)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding qualifying results: %v", err)
		return nil, err
	}
	return &data, nil
}
