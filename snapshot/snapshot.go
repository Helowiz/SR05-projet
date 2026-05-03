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

func Shot(board shape.WhiteBoard, id string) *Snapshot {
	display.Info("SNAP", "snapshot", board.String())

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

	//saveLocalSnapshot(snapshot)
	return snapshot

}

func saveSnapshot(snapshot *Snapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}
	//isplay.Info("SNAPSHOT", "snapshot", string(data))

	filename := fmt.Sprintf("snapshot_site_%d_%s.json", snapshot.SiteID, snapshot.Timestamp.Format("20060102_150405"))
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}
}

func SaveGlobalSnapshot(allLocalSnapshot []Snapshot) {
	globalBoard := shape.Empty_board()

	for _, snap := range allLocalSnapshot {
		for id, shape := range snap.Board.Shapes {
			globalBoard.AddShape(id, shape)
		}
	}

	globalSnapshot := Snapshot{
		SiteID:    allLocalSnapshot[0].SiteID,
		Timestamp: allLocalSnapshot[0].Timestamp, //TODO Horloge Vecotrielle
		Board:     globalBoard,
	}

	saveSnapshot(&globalSnapshot)
}

func SnapshotToString(snap *Snapshot) (string, error) {
	data, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func StringToSnapshot(data string) (*Snapshot, error) {
	var snap Snapshot
	err := json.Unmarshal([]byte(data), &snap)
	if err != nil {
		return nil, err
	}
	return &snap, nil
}
