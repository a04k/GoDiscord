package f1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const ergastAPI = "https://api.jolpi.ca/ergast/f1/current"

// Event represents a single F1 event from the schedule.
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
	"Red Bull":     0x1E5BC6,
	"Ferrari":      0xDC0000,
	"Mercedes":     0x00D2BE,
	"McLaren":      0xFF8700,
	"Aston Martin": 0x006F62,
	"Alpine":       0x0090FF,
	"Williams":     0x005AFF,
	"AlphaTauri":   0x2B4562,
	"Haas F1 Team": 0xFFFFFF,
	"Alfa Romeo":   0x960000,
	"RB F1 Team":   0x6600FF,
	"Sauber":       0x006400,
}

// Ergast API Response Structs

type DriverStandingsResponse struct {
	MRData struct {
		StandingsTable struct {
			Season        string `json:"season"`
			StandingsLists []struct {
				DriverStandings []struct {
					Position string `json:"position"`
					Points   string `json:"points"`
					Wins     string `json:"wins"`
					Driver   struct {
						GivenName  string `json:"givenName"`
						FamilyName string `json:"familyName"`
						Code       string `json:"code"`
					} `json:"Driver"`
				} `json:"DriverStandings"`
			} `json:"StandingsLists"`
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

type ConstructorStandingsResponse struct {
	MRData struct {
		StandingsTable struct {
			Season        string `json:"season"`
			StandingsLists []struct {
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

type QualifyingResponse struct {
	MRData struct {
		RaceTable struct {
			Races []QualifyingResponse_MRDatum_RaceTable_Race `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type QualifyingResponse_MRDatum_RaceTable_Race struct {
	RaceName          string `json:"raceName"`
	Circuit           CircuitInfo `json:"Circuit"`
	QualifyingResults []QualifyingResult `json:"QualifyingResults"`
}

type RaceResultsResponse struct {
	MRData struct {
		RaceTable struct {
			Races []RaceResultsResponse_MRDatum_RaceTable_Race `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type RaceResultsResponse_MRDatum_RaceTable_Race struct {
	RaceName string        `json:"raceName"`
	Circuit  CircuitInfo   `json:"Circuit"`
	Results  []RaceResult `json:"Results"`
}

type SprintResultsResponse struct {
	MRData struct {
		RaceTable struct {
			Races []SprintResultsResponse_MRDatum_RaceTable_Race `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type SprintResultsResponse_MRDatum_RaceTable_Race struct {
	RaceName      string         `json:"raceName"`
	Circuit       CircuitInfo    `json:"Circuit"`
	SprintResults []SprintResult `json:"SprintResults"`
}

// Common sub-structs

type CircuitInfo struct {
	CircuitName string `json:"circuitName"`
	Location    struct {
		Country string `json:"country"`
	} `json:"Location"`
}

type QualifyingResult struct {
	Position    string `json:"position"`
	Driver      DriverInfo `json:"Driver"`
	Constructor ConstructorInfo `json:"Constructor"`
	Q1          string `json:"Q1"`
	Q2          string `json:"Q2"`
	Q3          string `json:"Q3"`
}

type RaceResult struct {
	Position    string `json:"position"`
	Points      string `json:"points"`
	Driver      DriverInfo `json:"Driver"`
	Constructor ConstructorInfo `json:"Constructor"`
	Laps        string `json:"laps"`
	Status      string `json:"status"`
	Time        struct {
		Time string `json:"time"`
	} `json:"Time"`
}

type SprintResult struct {
	Position    string `json:"position"`
	Points      string `json:"points"`
	Driver      DriverInfo `json:"Driver"`
	Constructor ConstructorInfo `json:"Constructor"`
	Laps        string `json:"laps"`
	Status      string `json:"status"`
	Time        struct {
		Time string `json:"time"`
	} `json:"Time"`
}

type DriverInfo struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Code       string `json:"code"`
}

type ConstructorInfo struct {
	Name string `json:"name"`
}

// --- Fetch Functions ---

func fetchAndDecode(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	return nil
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

func FetchRaceResultsByRound(round int) (*RaceResultsResponse, error) {
	var data RaceResultsResponse
	url := fmt.Sprintf("%s/%d/results.json", ergastAPI, round)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding race results for round %d: %v", round, err)
		return nil, err
	}
	return &data, nil
}

func FetchLatestQualifyingResults() (*QualifyingResponse, error) {
	var data QualifyingResponse
	url := fmt.Sprintf("%s/last/qualifying.json", ergastAPI)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding qualifying results: %v", err)
		return nil, err
	}
	return &data, nil
}

func FetchQualifyingResultsByRound(round int) (*QualifyingResponse, error) {
	var data QualifyingResponse
	url := fmt.Sprintf("%s/%d/qualifying.json", ergastAPI, round)
	if err := fetchAndDecode(url, &data); err != nil {
		log.Printf("Error fetching or decoding qualifying results for round %d: %v", round, err)
		return nil, err
	}
	return &data, nil
}

func FetchSprintResultsByRound(round int) (*SprintResultsResponse, error) {
	var data SprintResultsResponse
	url := fmt.Sprintf("%s/%d/sprint.json", ergastAPI, round)
	if err := fetchAndDecode(url, &data); err != nil {
		// Ergast often returns a 404 for rounds without a sprint race, which is expected.
		// We log it for debugging but return nil, nil to indicate no data.
		log.Printf("Note: Could not fetch sprint results for round %d (may not exist): %v", round, err)
		return nil, nil
	}
	return &data, nil
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

func FetchF1Events() ([]Event, error) {
	jsonFile, err := os.Open("commands/sports/f1/f1_schedule_2025.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var events []Event
	json.Unmarshal(byteValue, &events)

	return events, nil
}
