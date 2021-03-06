package game

// 发起匹配content
type MatchCommand struct {
	Name int `json:"name"`
}

// 匹配结果content
type MatchResultCommand struct {
	Players []string `json:"players"`
	RoomId int `json:"roomId"`
}

// 发牌content
type GiveCardCommand struct {
	State int `json:"state"`
	RoomdId int `json:"roomId"`
	Cards []int `json:"cards"`
}

type CardSet struct {
	Type CardType `json:"type"`
	Header int `json:"header"`
	Cards []int `json:"cards"`
}

// 状态变化command
type StateChangeCommand struct {
	State int `json:"state"`
	CurPlayerIndex int `json:"curPlayerIndex"`
	CurCard CardSet `json:"curCard"`
	Scores map[int]int `json:"scores"`
	NowScore int `json:"nowScore"`
}

// 房间内基本信息
type CommonRoomCommand struct {
	RoomId int `json:"roomId"`
	Index int `json:"index"`
}

// 玩家上传的出牌消息
type PlayCardInCommand struct {
	RoomId int `json:"roomId"`
	Index int `json:"index"`
	CurCard CardSet `json:"curCards"`
}

// 出牌消息
type CardOutCommand struct {
	State int `json:"state"`
	CurPlayerIndex int `json:"curPlayerIndex"`
	CurCard CardSet `json:"curCard"`
}

// 罚分消息
type PunishCommand struct {
	State int `json:"state"` // state = 4 表示罚分
	Seat int `json:"seat"` // 被罚的座位
	Score int `json:"score"` // 被罚的分数
	PunCard CardSet `json:"punCard"` // 被罚丢弃的牌
}