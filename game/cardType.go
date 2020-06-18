package game

type CardType int

const (
	PASSED            CardType = -2 //过
	INIT              CardType = -1 //前面还没有牌（首家）
	ERROR             CardType = 0  //错误牌型
	SINGLE            CardType = 1  //单牌
	DOUBLE            CardType = 2  //对子
	CONNECT_DOUBLE    CardType = 3  //姊妹对
	THREE_TWO         CardType = 4  //三带二
	STRAIGHT          CardType = 5  //顺子
	THREE_ONE         CardType = 6  //三带一，仅最后出牌可出
	CONNECT_THREE_TWO CardType = 7  //连续三带二
	BOMB              CardType = 8  // 炸弹
)
