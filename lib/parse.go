package lib

import (
	"encoding/json"
	"fmt"
	"sort"
)

// BufferToMessage takes in a message buffer and returns a message string.
func BufferToMessage(buf []byte) (string, error) {
	var msg string

	if data := jsonToSonarr(buf); data != nil {
		if data.Valid() {
			msg += data.Message()
		}
	}

	if data := jsonToLidarr(buf); data != nil {
		if data.Valid() {
			msg += data.Message()
		}
	}

	if data := jsonToGoogleCloud(buf); data != nil {
		if data.Valid() {
			msg += data.Message()
		}
	}

	if data := jsonToPlex(buf); data != nil {
		if data.Valid() {
			msg += data.Message()
		}
	}

	if msg == "" {
		var f map[string]string
		if err := json.Unmarshal(buf, &f); err != nil {
			return "", fmt.Errorf("decoding json to map: %w", err)
		}

		var keys []string
		for k := range f {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			msg += fmt.Sprintf("%s: %s\n", k, f[k])
		}
	}

	return msg, nil
}

// Sonarr is the structure of messages we get from Sonarr.
//
// Generated by https://mholt.github.io/json-to-go/
type Sonarr struct {
	Episodes []struct {
		ID             int    `json:"id"`
		EpisodeNumber  int    `json:"episodeNumber"`
		SeasonNumber   int    `json:"seasonNumber"`
		Title          string `json:"title"`
		QualityVersion int    `json:"qualityVersion"`
	} `json:"episodes"`
	EventType string `json:"eventType"`
	Series    struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Path   string `json:"path"`
		TvdbID int    `json:"tvdbId"`
	} `json:"series"`
}

func jsonToSonarr(buf []byte) *Sonarr {
	var data Sonarr
	if err := json.Unmarshal(buf, &data); err != nil {
		log.WithError(err).Error("decoding json to Sonarr")
		return nil
	}
	log.WithField("data", data).Debug("Sonarr data decoded")

	return &data
}

// Message returns a string representation of this object for human consumption.
func (j *Sonarr) Message() string {
	var msg string
	for _, ep := range j.Episodes {
		msg += fmt.Sprintf("Sonarr: %s %dx%02d - %q\n", j.Series.Title, ep.SeasonNumber, ep.EpisodeNumber, j.EventType)
	}

	return msg
}

// Valid checks that the data is good.
func (j *Sonarr) Valid() bool {
	return j.EventType != "" && len(j.Episodes) > 0
}

// GoogleCloud is the structure of messages we get from Google Cloud Platform Alerting.
//
// Generated by https://mholt.github.io/json-to-go/
type GoogleCloud struct {
	Incident struct {
		IncidentID   string `json:"incident_id"`
		ResourceID   string `json:"resource_id"`
		ResourceName string `json:"resource_name"`
		Resource     struct {
			Type   string `json:"type"`
			Labels struct {
				Host string `json:"host"`
			} `json:"labels"`
		} `json:"resource"`
		ResourceTypeDisplayName string `json:"resource_type_display_name"`
		Metric                  struct {
			Type        string `json:"type"`
			DisplayName string `json:"displayName"`
		} `json:"metric"`
		StartedAt     int    `json:"started_at"`
		PolicyName    string `json:"policy_name"`
		ConditionName string `json:"condition_name"`
		Condition     struct {
			Name               string `json:"name"`
			DisplayName        string `json:"displayName"`
			ConditionThreshold struct {
				Filter       string `json:"filter"`
				Aggregations []struct {
					AlignmentPeriod    string   `json:"alignmentPeriod"`
					PerSeriesAligner   string   `json:"perSeriesAligner"`
					CrossSeriesReducer string   `json:"crossSeriesReducer"`
					GroupByFields      []string `json:"groupByFields"`
				} `json:"aggregations"`
				Comparison     string  `json:"comparison"`
				ThresholdValue float64 `json:"thresholdValue"`
				Duration       string  `json:"duration"`
				Trigger        struct {
					Count int `json:"count"`
				} `json:"trigger"`
			} `json:"conditionThreshold"`
		} `json:"condition"`
		URL     string      `json:"url"`
		State   string      `json:"state"`
		EndedAt interface{} `json:"ended_at"`
		Summary string      `json:"summary"`
	} `json:"incident"`
	Version string `json:"version"`
}

func jsonToGoogleCloud(buf []byte) *GoogleCloud {
	var data GoogleCloud
	if err := json.Unmarshal(buf, &data); err != nil {
		log.WithError(err).Error("decoding json to GoogleCloud")
		return nil
	}
	log.WithField("data", data).Debug("GoogleCloud data decoded")

	return &data
}

// Message returns a string representation of this object for human consumption.
func (j *GoogleCloud) Message() string {
	return fmt.Sprintf("GCP Alert - %q\n", j.Incident.Summary)
}

// Valid checks that the data is good.
func (j *GoogleCloud) Valid() bool {
	return j.Incident.IncidentID != ""
}

// Lidarr provides a structure for Lidarr updates.
//
// Generated by https://mholt.github.io/json-to-go/
type Lidarr struct {
	Albums []struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		QualityVersion int    `json:"qualityVersion"`
	} `json:"albums"`
	EventType string `json:"eventType"`
	Artist    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
		MbID string `json:"mbId"`
	} `json:"artist"`
}

func jsonToLidarr(buf []byte) *Lidarr {
	var data Lidarr
	if err := json.Unmarshal(buf, &data); err != nil {
		log.WithError(err).Error("decoding json to Lidarr")
		return nil
	}
	log.WithField("data", data).Debug("Lidarr data decoded")

	return &data
}

// Message returns a string representation of this object for human consumption.
func (j *Lidarr) Message() string {
	var msg string
	for _, ep := range j.Albums {
		msg += fmt.Sprintf("Lidarr: %s - %q - %s\n", j.Artist.Name, ep.Title, j.EventType)
	}
	return msg
}

// Valid checks that the data is good.
func (j *Lidarr) Valid() bool {
	return j.EventType != "" && len(j.Albums) > 0
}

// Plex provides a structure for Plex updates.
//
// Generated by https://mholt.github.io/json-to-go/
type Plex struct {
	Event   string `json:"event"`
	User    bool   `json:"user"`
	Owner   bool   `json:"owner"`
	Account struct {
		ID    int    `json:"id"`
		Thumb string `json:"thumb"`
		Title string `json:"title"`
	} `json:"Account"`
	Server struct {
		Title string `json:"title"`
		UUID  string `json:"uuid"`
	} `json:"Server"`
	Player struct {
		Local         bool   `json:"local"`
		PublicAddress string `json:"publicAddress"`
		Title         string `json:"title"`
		UUID          string `json:"uuid"`
	} `json:"Player"`
	Metadata struct {
		LibrarySectionType    string  `json:"librarySectionType"`
		RatingKey             string  `json:"ratingKey"`
		Key                   string  `json:"key"`
		ParentRatingKey       string  `json:"parentRatingKey"`
		GrandparentRatingKey  string  `json:"grandparentRatingKey"`
		GUID                  string  `json:"guid"`
		ParentGUID            string  `json:"parentGuid"`
		GrandparentGUID       string  `json:"grandparentGuid"`
		Type                  string  `json:"type"`
		Title                 string  `json:"title"`
		GrandparentTitle      string  `json:"grandparentTitle"`
		ParentTitle           string  `json:"parentTitle"`
		ContentRating         string  `json:"contentRating"`
		Summary               string  `json:"summary"`
		Index                 int     `json:"index"`
		ParentIndex           int     `json:"parentIndex"`
		Rating                float64 `json:"rating"`
		ViewCount             int     `json:"viewCount"`
		LastViewedAt          int     `json:"lastViewedAt"`
		Year                  int     `json:"year"`
		Thumb                 string  `json:"thumb"`
		Art                   string  `json:"art"`
		ParentThumb           string  `json:"parentThumb"`
		GrandparentThumb      string  `json:"grandparentThumb"`
		GrandparentArt        string  `json:"grandparentArt"`
		GrandparentTheme      string  `json:"grandparentTheme"`
		OriginallyAvailableAt string  `json:"originallyAvailableAt"`
		AddedAt               int     `json:"addedAt"`
		UpdatedAt             int     `json:"updatedAt"`
	} `json:"Metadata"`
}

func jsonToPlex(buf []byte) *Plex {
	var data Plex
	if err := json.Unmarshal(buf, &data); err != nil {
		log.WithError(err).Error("decoding json to Plex")
		return nil
	}
	log.WithField("data", data).Info("Plex data decoded")

	return &data
}

// Message returns a string representation of this object for human consumption.
func (j *Plex) Message() string {
	return fmt.Sprintf("Plex - %q : %s %dx%d\n", j.Event, j.Metadata.GrandparentTitle, j.Metadata.ParentIndex, j.Metadata.Index)
}

// Valid checks that the data is good.
func (j *Plex) Valid() bool {
	return j.Event != ""
}
