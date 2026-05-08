package snapshot

import (
	"SR05_projet/display"
	"SR05_projet/shape"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Snapshot struct {
	SiteID    int              `json:"site_id"`
	Timestamp time.Time        `json:"timestamp"` //TODO horloge vectorielle
	Board     shape.WhiteBoard `json:"board"`
}

type GlobalSnapshot struct {
	States     []Snapshot `json:"states"`
	MsgChannel []string   `json:"channel"` // Ce qu'il y a dans les canaux
}

func Shot(board shape.WhiteBoard, id string) *Snapshot {
	i, err := strconv.Atoi(id)
	if err != nil {
		display.Error("SNAPSHOT", "shot", err.Error())
		return nil
	}

	snapshot := &Snapshot{
		SiteID:    i,
		Timestamp: time.Now(),
		Board:     board,
	}
	return snapshot
}

func SaveSnapshot(snapshot *Snapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}

	filename := fmt.Sprintf("snapshot_site_%d_%s.json", snapshot.SiteID, snapshot.Timestamp.Format("20060102_150405"))
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}
}

func ToString(snap *Snapshot) (string, error) {
	data, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ToSnapshot(data string) (*Snapshot, error) {
	var snap Snapshot
	err := json.Unmarshal([]byte(data), &snap)
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func Merge(globalSnapshot *GlobalSnapshot, receiveSnapshot *Snapshot) *GlobalSnapshot {
	if globalSnapshot == nil {
		globalSnapshot = &GlobalSnapshot{
			States:     make([]Snapshot, 0),
			MsgChannel: make([]string, 0),
		}
	}
	globalSnapshot.States = append(globalSnapshot.States, *receiveSnapshot)
	return globalSnapshot
}

func MergeMsg(global *GlobalSnapshot, msg string) *GlobalSnapshot {
	global.MsgChannel = append(global.MsgChannel, msg)
	return global
}
