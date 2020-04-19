package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

const key string = "api_key=" + config.RiotKey

var numberOfPulls int

//		*********	 SUMMONER SEARCH STRUCTS	*********		//
type summoner struct {
	ProfileIconID int
	Name          string
	Puuid         string
	SummonerLevel int
	AccountID     string
	ID            string
	RevisionDate  int
}

//		*********	  MATCH HISTORY STRUCTS		*********		//
type matchHistory struct {
	Matches    []match
	EndIndex   int
	StartIndex int
	TotalGames int
}

type match struct {
	Lane       string
	GameID     int64
	Champion   int
	PlatformID string
	Timestamp  int
	Queue      int
	Role       string
	Season     int
}

//		*********	   MATCH STATS STRUCTS		*********		//
type matchStats struct {
	SeasonID              int
	QueueID               int
	GameID                int64
	ParticipantIDentities []participantIDentities
	GameVersion           string
	PlatformID            string
	GameMode              string
	MapID                 int
	GameType              string
}

type participantIDentities struct {
	Player        player
	ParticipantID int
}

type player struct {
	CurrentPlatformID string
	SummonerName      string
	MatchHistoryURI   string
	PlatformID        string
	CurrentAccountID  string
	ProfileIcon       int
	SummonerID        string
	AccountID         string
}

//		*********	  	TIMELINE STRUCTS		*********		//
type timeline struct {
	Frames        []frame
	FrameInterval int
}

type frame struct {
	Timestamp         int
	ParticipantFrames participantFrames
	Events            []timelineEvents
}

type participantFrames struct {
	ParticipantFrames1  singularParticipant
	ParticipantFrames2  singularParticipant
	ParticipantFrames3  singularParticipant
	ParticipantFrames4  singularParticipant
	ParticipantFrames5  singularParticipant
	ParticipantFrames6  singularParticipant
	ParticipantFrames7  singularParticipant
	ParticipantFrames8  singularParticipant
	ParticipantFrames9  singularParticipant
	ParticipantFrames10 singularParticipant
}

type singularParticipant struct {
	TotalGold           int
	TeamScore           int
	ParticipantID       int
	Level               int
	CurrentGold         int
	MinionsKilled       int
	DominionScore       int
	Position            position
	Xp                  int
	JungleMinionsKilled int
}

type position struct {
	Y int
	X int
}

type timelineEvents struct {
	EventType               string   `json:"eventType"`
	TowerType               string   `json:"towerType"`
	TeamID                  int      `json:"teamID"`
	AscendedType            string   `json:"ascendedType"`
	KillerID                int      `json:"killerID"`
	LevelUpType             string   `json:"levelUpType"`
	PointCaptured           string   `json:"pointCaptured"`
	AssistingParticipantIDs []int    `json:"assistingParticipantIDs"`
	WardType                string   `json:"wardType"`
	MonsterType             string   `json:"monsterType"`
	Type                    string   `json:"type"`
	SkillSlot               int      `json:"skillSlot"`
	VictimID                int      `json:"victimID"`
	Timestamp               int64    `json:"timestamp"`
	AfterID                 int      `json:"afterID"`
	MonsterSubType          string   `json:"monsterSubType"`
	LaneType                string   `json:"laneType"`
	ItemID                  int      `json:"itemID"`
	ParticipantID           int      `json:"participantID"`
	BuildingType            string   `json:"buildingType"`
	CreatorID               int      `json:"creatorID"`
	Position                position `json:"position"`
	BeforeID                int      `json:"beforeID"`
}

func urlToStructSummoner(url string) (summoner, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	//defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
		}

		var data summoner
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println(err.Error())
		}

		return data, nil

	}
	var wrongSCode summoner
	if resp.StatusCode == 404 {

		return wrongSCode, errors.New("could not find EUW summoner name")
	}
	return wrongSCode, errors.New("status code -> " + strconv.Itoa(resp.StatusCode))
}

func urlToStructMatchHistory(url string) matchHistory {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	//defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var dataStruct matchHistory
	err = json.Unmarshal(body, &dataStruct)
	if err != nil {
		fmt.Println(err.Error())
	}

	return dataStruct
}

func urlToStructMatchStats(url string) matchStats {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	//defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var dataStruct matchStats
	err = json.Unmarshal(body, &dataStruct)
	if err != nil {
		fmt.Println(err.Error())
	}

	return dataStruct
}

func urlToTimeline(url string) timeline {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	//defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var dataStruct timeline
	err = json.Unmarshal(body, &dataStruct)
	if err != nil {
		fmt.Println(err.Error())
	}

	return dataStruct
}

func summonerSearch(name string) (summoner, error) {
	base := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/"
	url := base + name + "?" + key
	numberOfPulls++
	sum, err := urlToStructSummoner(url)
	fmt.Println(url)
	return sum, err
}

func matchHistorySearch(summoner summoner, start int) matchHistory {
	base := "https://euw1.api.riotgames.com/lol/match/v4/matchlists/by-account/"
	filters := "queue=420&beginIndex=" + strconv.Itoa(start) + "&"
	accountID := string(summoner.AccountID)
	url := base + accountID + "?" + filters + key
	fmt.Println(url)
	numberOfPulls++
	return urlToStructMatchHistory(url)
}

func matchStatsSearch(matchID int64) matchStats {
	gameID := strconv.FormatInt(matchID, 10)
	base := "https://euw1.api.riotgames.com/lol/match/v4/matches/"
	url := base + gameID + "?" + key
	numberOfPulls++
	return urlToStructMatchStats(url)
}

func timelineSearch(matchID int64) timeline {
	base := "https://euw1.api.riotgames.com/lol/match/v4/timelines/by-match/"
	url := base + strconv.FormatInt(matchID, 10) + "?" + key
	numberOfPulls++
	return urlToTimeline(url)
}

func makeGraph(name string) error {
	summoner, err := summonerSearch(name)
	if err != nil {
		return err
	}
	sampleSize := 100
	totalDeathsGraph := [60]float64{}
	totalKillsGraph := [60]float64{}
	totalAssistsGraph := [60]float64{}
	mH0 := matchHistorySearch(summoner, 0)
	summonerID := summoner.ID
	for k := 0; k < sampleSize; k++ {
		playerIDMatch := false
		var playerID int
		matchStatistics := matchStatsSearch(mH0.Matches[k].GameID)
		if len(matchStatistics.ParticipantIDentities) < 10 {
			time.Sleep(time.Minute)
			matchStatistics = matchStatsSearch(mH0.Matches[k].GameID)
		}
		for l := 0; playerIDMatch == false; l++ {
			if matchStatistics.ParticipantIDentities[l].Player.SummonerID == summonerID {
				playerID = matchStatistics.ParticipantIDentities[l].ParticipantID
				playerIDMatch = true
			}
		}
		timeline := timelineSearch(mH0.Matches[k].GameID)
		if numberOfPulls > 95 {
			time.Sleep(2 * time.Minute)
			numberOfPulls = 0
		}
		for i := 0; i < len(timeline.Frames); i++ {
			for j := 0; j < len(timeline.Frames[i].Events); j++ {
				if timeline.Frames[i].Events[j].Type == "CHAMPION_KILL" && timeline.Frames[i].Events[j].VictimID == playerID {
					totalDeathsGraph[i]++
				}
				if timeline.Frames[i].Events[j].Type == "CHAMPION_KILL" && timeline.Frames[i].Events[j].KillerID == playerID {
					totalKillsGraph[i]++
				}
				for k := 0; k < len(timeline.Frames[i].Events[j].AssistingParticipantIDs); k++ {
					if timeline.Frames[i].Events[j].Type == "CHAMPION_KILL" && timeline.Frames[i].Events[j].AssistingParticipantIDs[k] == playerID {
						totalAssistsGraph[i]++
					}
				}

			}
		}
	}

	for i := 1; i < len(totalDeathsGraph); i++ {
		totalDeathsGraph[i] = totalDeathsGraph[i] / float64(sampleSize)
	}
	for i := 1; i < len(totalKillsGraph); i++ {
		totalKillsGraph[i] = totalKillsGraph[i] / float64(sampleSize)
	}
	for i := 1; i < len(totalAssistsGraph); i++ {
		totalAssistsGraph[i] = totalAssistsGraph[i] / float64(sampleSize)
	}

	data := make(plotter.Values, 60)
	for i := 0; i < 60; i++ {
		data[i] = totalDeathsGraph[i]
	}
	data2 := make(plotter.Values, 60)
	for i := 0; i < 60; i++ {
		data2[i] = totalKillsGraph[i]
	}
	data3 := make(plotter.Values, 60)
	for i := 0; i < 60; i++ {
		data3[i] = totalAssistsGraph[i]
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p2, err := plot.New()
	if err != nil {
		panic(err)
	}
	p3, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Y.Label.Text = "Average deaths per min over " + strconv.Itoa(sampleSize) + " ranked games"
	p2.Y.Label.Text = "Average kills per min over " + strconv.Itoa(sampleSize) + " ranked games"
	p3.Y.Label.Text = "Average assists per min over " + strconv.Itoa(sampleSize) + " ranked games"

	p.X.Label.Text = "Minute"
	p2.X.Label.Text = "Minute"
	p3.X.Label.Text = "Minute"

	w := vg.Points(10)

	barsA, err := plotter.NewBarChart(data, w)
	if err != nil {
		panic(err)
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = plotutil.Color(2)
	barsB, err := plotter.NewBarChart(data2, w)
	if err != nil {
		panic(err)
	}
	barsB.LineStyle.Width = vg.Length(0)
	barsB.Color = plotutil.Color(2)
	barsC, err := plotter.NewBarChart(data3, w)
	if err != nil {
		panic(err)
	}
	barsC.LineStyle.Width = vg.Length(0)
	barsC.Color = plotutil.Color(2)

	p.Add(barsA)
	p.NominalX("0", "", "2", "", "4", "", "6", "", "8", "", "10", "", "12", "", "14", "", "16", "", "18", "", "20", "", "22", "", "24", "", "26", "", "28", "", "30", "", "32", "", "34", "", "36", "", "38", "", "40", "", "42", "", "44", "", "46", "", "48", "", "50", "", "52", "", "54", "", "56", "", "58", "", "60")
	p2.Add(barsB)
	p2.NominalX("0", "", "2", "", "4", "", "6", "", "8", "", "10", "", "12", "", "14", "", "16", "", "18", "", "20", "", "22", "", "24", "", "26", "", "28", "", "30", "", "32", "", "34", "", "36", "", "38", "", "40", "", "42", "", "44", "", "46", "", "48", "", "50", "", "52", "", "54", "", "56", "", "58", "", "60")
	p3.Add(barsC)
	p3.NominalX("0", "", "2", "", "4", "", "6", "", "8", "", "10", "", "12", "", "14", "", "16", "", "18", "", "20", "", "22", "", "24", "", "26", "", "28", "", "30", "", "32", "", "34", "", "36", "", "38", "", "40", "", "42", "", "44", "", "46", "", "48", "", "50", "", "52", "", "54", "", "56", "", "58", "", "60")

	plots := make([][]*plot.Plot, 3)
	plots[0] = make([]*plot.Plot, 1)
	plots[1] = make([]*plot.Plot, 1)
	plots[2] = make([]*plot.Plot, 1)

	plots[0][0] = p
	plots[1][0] = p2
	plots[2][0] = p3

	img := vgimg.New(10*vg.Inch, 12*vg.Inch)
	dc := draw.New(img)

	t := draw.Tiles{
		Rows:      3,
		Cols:      1,
		PadX:      vg.Millimeter,
		PadY:      vg.Millimeter,
		PadTop:    vg.Points(2),
		PadBottom: vg.Points(2),
		PadLeft:   vg.Points(2),
		PadRight:  vg.Points(2),
	}

	canvases := plot.Align(plots, t, dc)
	for j := 0; j < 3; j++ {
		for i := 0; i < 1; i++ {
			if plots[j][i] != nil {
				plots[j][i].Draw(canvases[j][i])
			}
		}
	}

	write, err := os.Create("barchart.png")
	if err != nil {
		panic(err)
	}
	defer write.Close()
	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(write); err != nil {
		panic(err)
	}
	return nil
}
