package socket

import (
	"PaoDeKuaiServer/game"
	"encoding/json"
	"math/rand"
	"time"
)

type RoomPlayItem struct {
	ws   *Client
	name string
}

type RoomServer struct {
	// 房间id
	roomId int

	// 扑克逻辑处理
	pokerLogic *game.PokerLogic

	// 玩家列表
	players []RoomPlayItem

	// 三位玩家手中的牌
	playerCards [3][17]int

	// 当前出牌玩家
	curPlayerIndex int

	// 当前最大牌情况，牌型，牌型中的头牌，具体牌
	curCard game.CardSet

	// 当前由于炸弹而翻倍的次数
	bombTimes int

	// 当前牌最大的位置
	nowBigger int

	// 当前有几个人过牌
	passNum int

	// 罚分情况
	punishScores map[int]int

	// 得分情况
	scores map[int]int
}

func NewRoomServer() *RoomServer {
	// TODO: 做一些实际初始化的事
	return &RoomServer{
		curPlayerIndex: 1,
		pokerLogic: &game.PokerLogic{},
		players: make([]RoomPlayItem, 3),
		curCard: game.CardSet{
			Type: game.INIT,
			Header: 0,
			Cards: nil,
		},
		punishScores: make(map[int]int),
		scores: make(map[int]int),
	}
}

func (roomServer *RoomServer) InitGame() {
	// 一开始没有罚分
	for i := 1; i < 4; i++ {
		roomServer.punishScores[i] = 0
	}

	// 发牌
	cards := roomServer.getNewCards51()

	// 发牌的同时, 记录首个出牌人的位置
	firstIndex := 0

	for i := 0; i < 17; i++ {
		roomServer.playerCards[0][i] = cards[i]
		roomServer.playerCards[1][i] = cards[i + 17]
		roomServer.playerCards[2][i] = cards[i + 34]

		if (firstIndex != 0) {
			continue
		}

		// 红桃3出头
		if (cards[i] == 1) {
			firstIndex = 1
		} else if (cards[i + 17] == 1) {
			firstIndex = 2
		} else if (cards[i + 34] == 1) {
			firstIndex = 3
		}
	}

	// 将牌消息发送给客户端
	for i := 0; i < 3; i++ {
		giveCardCommand := game.GiveCardCommand{}
		giveCardCommand.State = 0;
		giveCardCommand.RoomdId = roomServer.roomId
		giveCardCommand.Cards = roomServer.playerCards[i][0:17]

		msg := Message{}
		msg.Code = 0
		msg.Command = PLAY_GAME
		msg.Seq = 0
		msg.Content, _ = json.Marshal(giveCardCommand)

		tempMsg, _ := json.Marshal(msg)
		roomServer.players[i].ws.send  <- tempMsg
	}

	roomServer.curPlayerIndex = firstIndex
	roomServer.changeState(1)
}

// 状态变更  2是结算，1是游戏中
func (roomServer *RoomServer) changeState(state int) {
	stateChangeCommand := game.StateChangeCommand{
		State: state,
	}
	msg := Message{}
	switch (state) {
	case 1:
		stateChangeCommand.CurPlayerIndex = roomServer.curPlayerIndex
		stateChangeCommand.CurCard = roomServer.curCard
		msg.Command = PLAY_GAME
		break;
	case 2:
		stateChangeCommand.Scores = roomServer.scores
		msg.Command = PLAY_GAME
		break;
	default:
		return;
	}

	msg.Content, _ = json.Marshal(stateChangeCommand)
	roomServer.sendToRoomPlayers(msg)
}

// 处理出牌消息
func (roomServer *RoomServer) handlePlayCard(message *Message) {
	seq := message.Seq

	playCardCommand := game.PlayCardInCommand{}

	// Todo: 发送错误返回？
	err := json.Unmarshal(message.Content, &playCardCommand)
	if err != nil {
		return
	}

	index := playCardCommand.Index

	curCard := playCardCommand.CurCard;
	if (len(curCard.Cards) > 0) {
		// 判断是否符合出牌规则

		// 牌型检查
		cardType := roomServer.pokerLogic.CalcuPokerType(curCard.Cards)
		if cardType != curCard.Type { // 计算出的牌型与传过来的不匹配
			//log.Printf("计算出的类型为 %d,传过来的为 %d", cardType, curCard.Type)
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -2)
			return
		}

		// 头牌检查
		cardHeader := roomServer.pokerLogic.CalcuPokerHeader(curCard.Cards, curCard.Type)
		if cardHeader != curCard.Header { // 计算出的头牌与传过来的不匹配
			//log.Printf("计算出的头牌为 %d,传过来的为 %d", cardHeader, curCard.Header)
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -2)
			return
		}

		// 是否可出牌检查
		handCardNum := roomServer.getHandCardNum(index - 1)
		if ! roomServer.pokerLogic.CanOut(&curCard, &roomServer.curCard, handCardNum) {
			//log.Printf("当前牌型为 %d, 头牌为 %d", roomServer.curCard.Type, roomServer.curCard.Header)
			//log.Printf("新来的牌型为 %d, 头牌为 %d", curCard.Type, curCard.Header)
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -3)
			return
		}

		// 移除手中的牌
		if ! roomServer.removeCards(index - 1, curCard.Cards) {
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -1)
			return
		}
	} else {
		// TODO: 有大牌不出, 罚分, 骂得脑壳搭起
	}

	// 告知玩家出牌成功
	roomServer.sendAck(PLAYER_PLAYCARD, index, seq, 0)

	if curCard.Type != game.PASSED { // 如果不是过牌，处理新的最大牌
		roomServer.curCard = curCard
		roomServer.passNum = 0

		// 通知本轮出牌，以及下一个应出牌的玩家
		roomServer.addCurIndex()
		roomServer.sendNextCardOut()
	} else {
		roomServer.passNum++
		if (1 == roomServer.passNum) {
			roomServer.nowBigger = roomServer.curPlayerIndex - 1;
			if (roomServer.nowBigger == 0) {
				roomServer.nowBigger = 3
			}
			roomServer.addCurIndex()
			roomServer.sendPassMsg()
		} else { // 不是1就是2
			roomServer.passNum = 0
			roomServer.curPlayerIndex = roomServer.nowBigger
			roomServer.curCard = game.CardSet{
				Type: game.INIT,
				Header: 0,
				Cards: nil,
			}
			roomServer.nowBigger = 0
			roomServer.sendNextCardOut()
		}
	}
}

// 获取手牌数
func (roomServer *RoomServer) getHandCardNum(index int) int {
	cardGroup := &roomServer.playerCards[index]

	cardNum := 0

	for j := 0; j < 17; j++ {
		if (0 != cardGroup[j]) { // 清除对应的牌
			cardNum++
		}
	}

	return cardNum
}

func (roomServer *RoomServer) sendAck(command MessageCode, index int, seq int, code int) {
	ackMsg := Message{}
	ackMsg.Command = command
	ackMsg.Seq = seq
	ackMsg.Code = code
	roomServer.sendToOnePlayer(index, ackMsg)
	return
}

func (roomServer *RoomServer) sendPassMsg() {
	passCurCard := game.CardSet{
		Type: game.PASSED,
		Header: 0,
		Cards: nil,
	}

	cardOutCommand := game.CardOutCommand{
		State: 1,
		CurPlayerIndex: roomServer.curPlayerIndex,
		CurCard: passCurCard,
	}

	msg := Message{}
	msg.Command = PLAY_GAME
	msg.Content, _ = json.Marshal(cardOutCommand)
	roomServer.sendToRoomPlayers(msg)
}

func (roomServer *RoomServer) sendNextCardOut() {
	cardOutCommand := game.CardOutCommand{
		State: 1,
		CurPlayerIndex: roomServer.curPlayerIndex,
		CurCard: roomServer.curCard,
	}

	msg := Message{}
	msg.Command = PLAY_GAME
	msg.Content, _ = json.Marshal(cardOutCommand)
	roomServer.sendToRoomPlayers(msg)
}

// 移除牌
func (roomServer *RoomServer) removeCards(index int, cards []int) bool {
	cardGroup := &roomServer.playerCards[index]

	var haveHard bool
	for i, length := 0, len(cards); i < length; i++ {
		haveHard = false

		for j := 0; j < 17; j++ {
			if (cards[i] == cardGroup[j]) { // 清除对应的牌
				haveHard = true
				cardGroup[j] = 0
				break
			}
		}

		if ( ! haveHard) {
			return false
		}
	}

	// 检查是否出完牌了
	hasOut := true
	for j := 0; j < 17; j++ {
		if (0 != cardGroup[j]) {
			hasOut = false
			break
		}
	}

	if (hasOut) {
		roomServer.countScore(index + 1)
		roomServer.changeState(2)
		roomServer.exit()
	}

	return true
}

// 算分
func (roomServer *RoomServer) countScore(winIndex int) {
	// 先算正常剩余牌数, 每张牌1分, 报停则不算
	other1 := winIndex - 1
	if (other1 == 0) {
		other1 = 3
	}

	other2 := winIndex + 1
	if (other2 == 4) {
		other2 = 1
	}

	other1Score := 0
	other2Score := 0

	for j := 0; j < 17; j++ {
		if (0 != roomServer.playerCards[other1 - 1][j]) {
			other1Score++;
		}

		if (0 != roomServer.playerCards[other2 - 1][j]) {
			other2Score++;
		}
	}

	if (other1Score == 1) {
		other1Score = 0
	}

	if (other2Score == 1) {
		other2Score = 0
	}

	roomServer.scores[other1] = -other1Score
	roomServer.scores[other2] = -other2Score
	roomServer.scores[winIndex] = other1Score + other2Score

	// 加上罚分
	for i := 1; i < 4; i++ {
		roomServer.scores[i] -= roomServer.punishScores[i]
	}
}


// 下一个座位
func (roomServer *RoomServer) addCurIndex() {
	roomServer.curPlayerIndex++;
	if (roomServer.curPlayerIndex > 3) {
		roomServer.curPlayerIndex = 1; //每次到4就变回1
	}
}

// 给房间内单个用户发送消息
func (roomServer *RoomServer) sendToOnePlayer(index int, data Message) {
	jsonData, _ := json.Marshal(data)
	roomServer.players[index - 1].ws.send <- jsonData
}

// 给房间内所有用户发送消息
func (roomServer *RoomServer) sendToRoomPlayers(data Message) {
	jsonData, _ := json.Marshal(data)
	for i := 0; i < len(roomServer.players); i++ {
		roomServer.players[i].ws.send <- jsonData
	}
}

// 退出房间
func (roomServer *RoomServer) exit() {
	for i := 0; i < len(roomServer.players); i++ {
		roomServer.players[i].ws.roomId = 0
	}

	msg := Message{}
	msg.Command = ROOM_EXIT
	roomServer.sendToRoomPlayers(msg)
}

// 拿到一副新好的牌, 没有大小王, 并且移除一个2
func (roomServer *RoomServer) getNewCards51() []int {
	totalCardNum := 51
	cards := make([]int, totalCardNum)
	for i := 0; i < totalCardNum; i++ {
		cards[i] = i + 1
	}

	rand.Seed(time.Now().UnixNano())

	// 洗牌算法
	for i := totalCardNum - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		if i != j {
			temp := cards[i]
			cards[i] = cards[j]
			cards[j] = temp
		}
	}

	return cards
}
