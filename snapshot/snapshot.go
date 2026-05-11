package snapshot

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"SR05_projet/shape"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Snapshot struct {
	SiteID      int              `json:"site_id"`
	HorlogeVect map[int]int      `json:"horloge_vect"`
	Board       shape.WhiteBoard `json:"board"`
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
		SiteID: i,
		Board:  board,
	}
	return snapshot
}

func SaveSnapshot(global *GlobalSnapshot) {
	data, err := json.MarshalIndent(global, "", "  ")
	if err != nil {
		display.Error("SNAPSHOT", "SaveGlobalSnapshot", err.Error())
		return
	}

	filename := fmt.Sprintf("global_snapshot_%s.json", time.Now().Format("20060102_150405")) //TODO mettre horloge
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		display.Error("SNAPSHOT", "SaveGlobalSnapshot", err.Error())
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

func GetHVMap(global GlobalSnapshot) map[int]map[int]int {
	hv_map := make(map[int]map[int]int)
	for _, snap := range global.States {
		hv_map[snap.SiteID] = snap.HorlogeVect
	}
	return hv_map
}

// HVSnap c'est {A: {HVA}, B:{HVB}}
func CheckCoherenceSnap(HVSnap map[int]map[int]int) bool {
	for site_a, HVa := range HVSnap {
		for site_b, HVb := range HVSnap {
			if site_a == site_b { // c'est le même site
				continue
			}
			h_b_dans_a := protocol.GetVal(HVa, site_b)
			h_b_dans_b := protocol.GetVal(HVb, site_b)

			if h_b_dans_a > h_b_dans_b { // le max des coupures doit correspondres au V_i[i]
				display.Info("", "CheckCoherenceSnap", "Snapshot pas coherente")
				return false
			}
		}
	}
	return true
}
