package shape

import (
	"SR05_projet/display"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const SHAPE_KEY_VALUE_SEP = ":"
const SHAPE_ENTRY_SEP = ";"

type Shape struct {
	Cmd   string
	X     int
	Y     int
	Color string // format : #RRGGBB
	W     int    // only for rect
	H     int    // only for rect
	R     int    // only for circle
	Val   string // only for text
	Size  int    // only for text
}

// format ;:cmd:rect;:id:shape-0007;:x:166;:y:66;:w:150;:h:55;:color:#4488ff

// Fonction utilitaire permettant de convertir un objet Shape en une chaîne de
// caractères au format ";:clé:valeur;:clé:valeur;..."
func (s Shape) toString() string {
	str := SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "cmd" + SHAPE_KEY_VALUE_SEP + s.Cmd
	str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "x" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.X)
	str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "y" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.Y)
	str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "color" + SHAPE_KEY_VALUE_SEP + s.Color

	switch s.Cmd {
	case "rect":
		str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "w" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.W)
		str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "h" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.H)
	case "circle":
		str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "r" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.R)
	case "text":
		str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "val" + SHAPE_KEY_VALUE_SEP + s.Val
		str += SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP + "size" + SHAPE_KEY_VALUE_SEP + strconv.Itoa(s.Size)
	}
	return str
}

// Fonction utilitaire permettant de convertir une chaîne de caractères au format
// ";:clé:valeur;:clé:valeur;..." en un objet Shape
func ParseShape(s string) (Shape, error) {
	shape := Shape{}
	err := error(nil)
	entries := strings.Split(s, SHAPE_ENTRY_SEP)
	for _, entry := range entries {
		cleanEntry := strings.TrimPrefix(entry, SHAPE_KEY_VALUE_SEP)
		kv := strings.Split(cleanEntry, SHAPE_KEY_VALUE_SEP)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		switch key {
		case "cmd":
			shape.Cmd = value
		case "x":
			shape.X, err = strconv.Atoi(value)
		case "y":
			shape.Y, err = strconv.Atoi(value)
		case "color":
			shape.Color = value
		case "w":
			shape.W, err = strconv.Atoi(value)
		case "h":
			shape.H, err = strconv.Atoi(value)
		case "r":
			shape.R, err = strconv.Atoi(value)
		case "val":
			shape.Val = value
		case "size":
			shape.Size, err = strconv.Atoi(value)
		}
	}
	return shape, err
}

// Méthode permettant de mettre à jour un champ d'un objet Shape à partir d'une clé et d'une valeur
func (s *Shape) set(key string, value string) error {
	var err error
	switch key {
	case "cmd":
		s.Cmd = value
	case "x":
		s.X, err = strconv.Atoi(value)
	case "y":
		s.Y, err = strconv.Atoi(value)
	case "color":
		s.Color = value
	case "w":
		s.W, err = strconv.Atoi(value)
	case "h":
		s.H, err = strconv.Atoi(value)
	case "r":
		s.R, err = strconv.Atoi(value)
	case "val":
		s.Val = value
	case "size":
		s.Size, err = strconv.Atoi(value)
	}
	return err
}

func (s Shape) String() string {
	base := fmt.Sprintf("Type: %-6s | Pos: (%d,%d) | Couleur: %s", s.Cmd, s.X, s.Y, s.Color)

	// Détails spécifiques selon la commande
	switch s.Cmd {
	case "rect":
		return fmt.Sprintf("%s | Taille: %dx%d", base, s.W, s.H)
	case "circle":
		return fmt.Sprintf("%s | Rayon: %d", base, s.R)
	case "text":
		return fmt.Sprintf("%s | Texte: \"%s\" (Taille: %d)", base, s.Val, s.Size)
	default:
		return base // Pour les cas inconnus
	}
}

type WhiteBoard struct {
	Shapes map[string]Shape
}

// Fonction utilitaire permettant de créer un tableau blanc vide
func Empty_board() WhiteBoard {
	return WhiteBoard{Shapes: make(map[string]Shape)}
}

// Méthode permettant d'ajouter une forme au tableau blanc
func (wb *WhiteBoard) AddShape(id string, shape Shape) {
	if _, exists := wb.Shapes[id]; exists {
		display.Error("", "addShape", fmt.Sprintf("Shape with id %s already exists", id))
	} else {
		wb.Shapes[id] = shape
	}
	display.Info("", "addShape", wb.String())
}

// Méthode permettant de supprimer une forme du tableau blanc à partir de son ID
func (wb *WhiteBoard) RemoveShape(id string) {
	if _, exists := wb.Shapes[id]; exists {
		delete(wb.Shapes, id)
	} else {
		display.Error("", "removeShape", fmt.Sprintf("Shape with id %s does not exist", id))
	}
}

func (wb *WhiteBoard) Clear() {
	*wb = Empty_board()
}

// Méthode permettant de mettre à jour une forme du tableau blanc à partir de son ID et d'une nouvelle forme
func (wb *WhiteBoard) UpdateShape(id string, key string, value string) {
	if shape, exists := wb.Shapes[id]; exists {
		err := shape.set(key, value)
		if err != nil {
			display.Error("", "updateShape", fmt.Sprintf("Failed to update shape with id %s: %v", id, err))
		} else {
			wb.Shapes[id] = shape
		}
	} else {
		display.Error("", "updateShape", fmt.Sprintf("Shape with id %s does not exist", id))
	}
}

// to string d'un whiteboard
func (wb *WhiteBoard) String() string {
	if len(wb.Shapes) == 0 {
		return "vide"
	}

	keys := make([]string, 0, len(wb.Shapes))
	for k := range wb.Shapes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("WhiteBoard contient les formes suivantes :\n")

	for _, k := range keys {
		builder.WriteString(fmt.Sprintf("  - ID: %s | Forme: %v\n", k, wb.Shapes[k]))
	}

	return builder.String()
}

// parse les données à mettre à jours
func GetUpdateFields(msg string) map[string]string {
	updates := make(map[string]string)
	entries := strings.Split(msg, SHAPE_ENTRY_SEP)

	for _, entry := range entries {
		if entry == "" {
			continue
		}

		cleanEntry := strings.TrimPrefix(entry, SHAPE_KEY_VALUE_SEP)

		kv := strings.SplitN(cleanEntry, SHAPE_KEY_VALUE_SEP, 2)

		if len(kv) != 2 {
			continue
		}

		key, value := kv[0], kv[1]

		if key == "op" || key == "id" {
			continue
		}

		updates[key] = value
	}

	return updates
}
