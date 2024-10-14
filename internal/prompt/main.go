package prompt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
)

type Prompt struct {
	Input   string        `json:"url"`
	Quality string        `json:"quality"`
	Start   time.Duration `json:"start"`
	End     time.Duration `json:"end"`
	Output  string        `json:"output"`
}

func (p *Prompt) UnmarshalJSON(b []byte) error {
	type Alias Prompt
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	var err error
	if aux.Start != "" {
		p.Start, err = time.ParseDuration(aux.Start)
		if err != nil {
			return err
		}
	}

	if aux.End != "" {
		p.End, err = time.ParseDuration(aux.End)
		if err != nil {
			return err
		}
	}
	return nil
}

func processFileInput(tw *twitch.Client, input string) ([]twitch.MediaUnit, chan types.ProgresbarChanData) {
	content, err := os.ReadFile(input)
	if err != nil {
		panic(err)
	}

	var body []Prompt
	if err := json.Unmarshal(content, &body); err != nil {
		panic(err)
	}

	var units []twitch.MediaUnit
	progressCh := make(chan types.ProgresbarChanData, len(body))

	for _, b := range body {
		unit, err := tw.NewMediaUnit(b.Input, b.Quality, b.Output, b.Start, b.End, progressCh)
		// TODO: do not skip
		if err != nil {
			fmt.Println(err)
			continue
		}
		units = append(units, unit)
	}

	return units, progressCh
}

func processFlagInput(tw *twitch.Client, input string) ([]twitch.MediaUnit, chan types.ProgresbarChanData) {
	content, err := os.ReadFile(input)
	if err != nil {
		panic(err)
	}

	var body []Prompt
	if err := json.Unmarshal(content, &body); err != nil {
		panic(err)
	}

	var units []twitch.MediaUnit
	progressCh := make(chan types.ProgresbarChanData, len(body))

	for _, b := range body {
		unit, err := tw.NewMediaUnit(b.Input, b.Quality, b.Output, b.Start, b.End, progressCh)
		// TODO: do not skip
		if err != nil {
			fmt.Println(err)
			continue
		}
		units = append(units, unit)
	}

	return units, progressCh
}

func (prompt *Prompt) ProcessInput(tw *twitch.Client) ([]twitch.MediaUnit, chan types.ProgresbarChanData) {
	if prompt.Input == "" {
		log.Fatalf("Input was not provided.")
	}

	var units []twitch.MediaUnit
	var progressCh chan types.ProgresbarChanData

	_, err := os.Stat(prompt.Input)
	if os.IsNotExist(err) {
		units, progressCh = processFlagInput(tw, prompt.Input)
	} else {
		units, progressCh = processFileInput(tw, prompt.Input)
	}
	return units, progressCh
}
