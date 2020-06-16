package socket

import (
	"encoding/json"
)

type MessageCode int

const (
	MATCH_PLAYER     MessageCode = 1
	PLAY_GAME        MessageCode = 2
	PLAYER_PLAYCARD  MessageCode = 3
	ROOM_EXIT         MessageCode = 4
)

type Message struct {
	Seq     int             `json:"seq"`
	Code    int             `json:"code"`
	Command MessageCode     `json:"command"`
	Content json.RawMessage `json:"content"`
}
