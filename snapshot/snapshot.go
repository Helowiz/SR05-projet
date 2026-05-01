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
	Timestamp time.Time        `json:"timestamp"`
	Board     shape.WhiteBoard `json:"board"`
}

func Shot(board shape.WhiteBoard, id string) {
	display.Info("SNAP", "snapshot", board.String())

	i, err := strconv.Atoi(id)
	if err != nil {
		display.Error("SNAPSHOT", "shot", err.Error())
		return
	}

	snapshot := Snapshot{
		SiteID:    i,
		Timestamp: time.Now(),
		Board:     board,
	}

	saveLocalSnapshot(snapshot)
}

func saveLocalSnapshot(snapshot Snapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}
	display.Info("SNAPSHOT", "snapshot", string(data))

	filename := fmt.Sprintf("snapshot_site_%d_%s.json", snapshot.SiteID, snapshot.Timestamp.Format("20060102_150405"))
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		display.Error("SNAPSHOT", "saveLocalSnapshot", err.Error())
		return
	}
}
