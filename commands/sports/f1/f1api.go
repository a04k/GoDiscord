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
    "Oracle Red Bull Racing":       0x4781D7, 
    "Scuderia Ferrari HP":          0xED1131, 
    "Mercedes-AMG Petronas F1 Team":0x00D7B6, 
    "McLaren":                      0xF47600, 
    "Aston Martin":                 0x229971, 
    "Alpine":                       0x00A1E8, 
    "Atlassian Williams Racing":   0x1868DB, 
    "Visa CashApp Racing Bulls":   0x6C98FF, 
    "Moneygram Haas F1 Team":      0x9C9FA2, 
    "Kick Sauber":                  0x01C00E,
}


// DriverInfo maps driver names/codes to constructor names for color coding
var DriverInfo = map[string]string{
    "VER": "Red Bull Racing",        // Max Verstappen – Red Bull Racing :contentReference[oaicite:0]{index=0}
    "TSU": "Red Bull Racing",        // Yuki Tsunoda – promoted mid-season to Red Bull Racing :contentReference[oaicite:1]{index=1}

    "LEC": "Ferrari",                // Charles Leclerc – Ferrari :contentReference[oaicite:2]{index=2}
    "HAM": "Ferrari",                // Lewis Hamilton – moved from Mercedes to Ferrari for 2025 :contentReference[oaicite:3]{index=3}

    "NOR": "McLaren",                // Lando Norris – McLaren :contentReference[oaicite:4]{index=4}
    "PIA": "McLaren",                // Oscar Piastri – McLaren :contentReference[oaicite:5]{index=5}

    "RUS": "Mercedes",               // George Russell – Mercedes :contentReference[oaicite:6]{index=6}
    "ANT": "Mercedes",               // Andrea Kimi Antonelli – promoted to Mercedes :contentReference[oaicite:7]{index=7}

    "ALO": "Aston Martin",           // Fernando Alonso – Aston Martin :contentReference[oaicite:8]{index=8}
    "STR": "Aston Martin",           // Lance Stroll – Aston Martin :contentReference[oaicite:9]{index=9}

    "GAS": "Alpine",                 // Pierre Gasly – Alpine :contentReference[oaicite:10]{index=10}
    "DOO": "Alpine",                 // Jack Doohan – Alpine :contentReference[oaicite:11]{index=11}

    "HUL": "Sauber",                 // Nico Hülkenberg – Sauber :contentReference[oaicite:12]{index=12}
    "BOR": "Sauber",                 // Gabriel Bortoleto – Sauber :contentReference[oaicite:13]{index=13}

    "BEA": "Haas",                   // Oliver Bearman – Haas :contentReference[oaicite:14]{index=14}
    "OCO": "Haas",                   // Esteban Ocon – Haas :contentReference[oaicite:15]{index=15}

    "ALB": "Williams",               // Alexander Albon – Williams :contentReference[oaicite:16]{index=16}
    "SAI": "Williams",               // Carlos Sainz Jr. – Williams :contentReference[oaicite:17]{index=17}

    "LAW": "Racing Bulls",           // Liam Lawson – Racing Bulls (junior Red Bull team) :contentReference[oaicite:18]{index=18}
    "HAD": "Racing Bulls",           // Isack Hadjar – Racing Bulls :contentReference[oaicite:19]{index=19}
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
