package game

import "sort"

// 其实现在是空的，但是万一有用呢？所以就没有写成静态方法了
type PokerLogic struct {

}

// 计算牌型
func (p *PokerLogic) CalcuPokerType(cards []int) CardType {
	// 转换为点数
	points := p.CardsToPoints(cards)

	length := len(points)

	if length == 1 { // 一张牌，当然是单牌
		return SINGLE
	} else if length == 2 { // 两张牌，王炸或一对
		if points[0] == points[1] { // 对子
			return DOUBLE
		}
	} else if length == 4 { // 姊妹对， 3带1， 炸弹
		maxSameNum := p.CalcuMaxSameNum(points)
		diffNum := p.CalcuDiffPoint(points)
		if maxSameNum == 4 { // 四张相同，炸弹
			return BOMB
		} else if maxSameNum == 3 { // 三张点数相同的，3带1
			return THREE_ONE
		} else if maxSameNum == 2 && diffNum == 2 && points[0] + 1 == points[3] && points[3] < 15 { // 两种点数，最多有两张一样的，且点数相连，姊妹对
			return CONNECT_DOUBLE;
		}
	} else if length >= 5 && p.IsStraight(points) && points[length - 1] < 15 { // 大于等于5张，是点数连续，且最大点数不超过2， 则是顺子
		return STRAIGHT
	} else if length == 5 { // 5张，只需检查3带2
		// 最多有3张相等的
		if p.CalcuMaxSameNum(points) == 3 {
			return THREE_TWO
		}
	} else { // 大于6的情况，姊妹对或连续三带二
		maxSameNum := p.CalcuMaxSameNum(points)
		diffPointNum := p.CalcuDiffPoint(points)

		if length % 2 == 0 && maxSameNum == 2 && diffPointNum == length / 2 && (points[length - 1] - points[0] == length / 2 - 1) && points[length - 1] < 15 { //姊妹对
			return CONNECT_DOUBLE;
		}

		if length % 5 == 0 { // 连续三带二
			// 找出点数出现大于等于3次的最长递增点数列表，如果此列表长度大于等于 len / 5且最靠前的一段不会到2，则可以
			threeCards := p.GetSameNumMaxStraightPoints(points, 3);
			if (len(threeCards) >= length / 5 && threeCards[length / 5 - 1] < 15) {
				return CONNECT_THREE_TWO;
			}
		}
	}

	// 没有这个牌型的
	return ERROR
}

// 取出所有点数数量等于num的点数
// 例如，现在牌中有3个3，3个4，2个5，1个6， 取出数量等于3的点数，则返回[3, 4]，取出数量等于2的点数，则返回[5]，取出数量等于1的点数，则返回[6]，其他都返回空数组
func (p *PokerLogic) GetSameNumPoints(points []int, num int) []int {
	length := len(points)
	newPoints := make([]int, length)
	pointIndex := 0

	nowNum := 1

	for i := 1; i < length; i++ {
		if points[i] == points[i-1] { // 与前一张相同
			nowNum++
		} else { // 与前一张不同，若前一张出现num次，加入数组
			if nowNum == num {
				newPoints[pointIndex] = points[i-1]
				pointIndex++
			}
			nowNum = 1
		}
	}

	if nowNum == num {
		newPoints[pointIndex] = points[length-1]
		pointIndex++
	}

	return newPoints[0:pointIndex]
}

// 取出所有点数数量大于等于num的点数
// 例如，现在牌中有3个3，3个4，2个5，1个6， 取出数量大于等于3的点数，则返回[3, 4]，取出数量大于等于2的点数，则返回[3, 4, 5]，取出数量大于等于1的点数，则返回[3, 4, 5, 6]
func (p *PokerLogic) GetGeNumPoints(points []int, num int) []int {
	length := len(points)
	newPoints := make([]int, length)
	pointIndex := 0

	nowNum := 1

	for i := 1; i < length; i++ {
		if points[i] == points[i-1] { // 与前一张相同
			nowNum++
		} else { // 与前一张不同，若前一张出现大于等于num次，加入数组
			if nowNum >= num {
				newPoints[pointIndex] = points[i-1]
				pointIndex++
			}
			nowNum = 1
		}
	}

	if nowNum >= num {
		newPoints[pointIndex] = points[length-1]
		pointIndex++
	}

	return newPoints[0:pointIndex]
}

// 从所有点数数量大于等于num的列表中，取出最长连续递增子列表
func (p *PokerLogic) GetSameNumMaxStraightPoints(points []int, num int) []int {
	geNumPoints := p.GetGeNumPoints(points, num)

	// 没有，直接返回
	length := len(geNumPoints)
	if length == 0 {
		return geNumPoints
	}

	maxStartPoint := geNumPoints[0]
	maxNum := 1

	nowStartPoint := geNumPoints[0]
	nowNum := 1

	for i := 1; i < length; i++ {
		if geNumPoints[i] == geNumPoints[i-1] + 1 {// 比上一张多1
			nowNum++;
		} else { // 重新开始计算
			if (nowNum > maxNum) {
				maxNum = nowNum;
				maxStartPoint = nowStartPoint;
			}
			nowNum = 1;
			nowStartPoint = geNumPoints[i];
		}
	}

	if (nowNum > maxNum) {
		maxNum = nowNum;
		maxStartPoint = nowStartPoint;
	}

	newPoints := make([]int, maxNum)
	for i := 0; i < maxNum; i++ {
		newPoints[i] = maxStartPoint + i
	}

	return newPoints
}

// 是否是顺子
func (p *PokerLogic) IsStraight(points []int) bool {
	length := len(points)
	for i := 1; i < length; i++ {
		if points[i] != points[i-1] + 1 { // 与前一张相同
			return false
		}
	}

	return true
}

// 有多少种不同的点数
func (p *PokerLogic) CalcuDiffPoint(points []int) int {
	diffNum := 1

	length := len(points)
	for i := 1; i < length; i++ {
		if points[i] != points[i-1] { // 与前一张不同，则出现了新的点数
			diffNum++
		}
	}

	return diffNum
}

// 最多有几张点数相等的牌
func (p *PokerLogic) CalcuMaxSameNum(points []int) int {
	length := len(points)
	nowNum := 1
	maxNum := 1

	for i := 1; i < length; i++ {
		if points[i] == points[i-1] { // 与前一张相同
			nowNum++
		} else { // 与前一张不同，重新开始计数
			if nowNum > maxNum {
				maxNum = nowNum
			}
			nowNum = 1
		}
	}

	if nowNum > maxNum {
		maxNum = nowNum
	}

	return maxNum
}

// 牌id转点数
func (p *PokerLogic) CardsToPoints(cards []int) []int {
	length := len(cards)
	points := make([]int, length)

	var point int
	for i := 0; i < length; i++ {
		value := cards[i]
		if value < 53 { // id 1-4 对应的是3， 5-8对应4， 依此类推 45 - 48对应14(A)， 49-52对应15(2)
			if value % 4 == 0 {
				point = value / 4 + 2
			} else {
				point = value / 4 + 3
			}
		} else { // 小王和大王
			point = value /4 + 2 + value % 4
		}

		points[i] = point
	}

	// 按点数升序排序
	sort.Ints(points)

	return points
}

// 计算头牌
func (p *PokerLogic) CalcuPokerHeader(cards []int, cardType CardType) int {
	points := p.CardsToPoints(cards)

	switch cardType {
	case SINGLE, DOUBLE, CONNECT_DOUBLE, STRAIGHT, BOMB:
		return points[0]
	case THREE_ONE,THREE_TWO:
		return points[2]
	case CONNECT_THREE_TWO:
		length := len(points) / 5
		threeCards := p.GetSameNumMaxStraightPoints(points, 3)
		endIndex := len(threeCards) - 1
		for ; endIndex >= length - 1; endIndex-- {
			if threeCards[endIndex] < 15 {
				break
			}
		}
		return threeCards[endIndex - length + 1]
	}

	return 0
}

// 获得首个出现num次的点数
func (p *PokerLogic) FirstPoint(points []int, num int) int {
	nowNum := 1;
	length := len(points)

	for i := 1; i < length; i++ {
		if (points[i] == points[i-1]) { //与上一张相同，数量加1
			nowNum++;
		} else { //重新开始计算
			if (nowNum == num) {
				return points[i-1];
			}
			nowNum = 1;
		}
	}

	if (nowNum == num) {
		return points[length - 1];
	}

	return 0
}

// 是否可以出牌
func (p *PokerLogic) CanOut(newCardSet *CardSet, nowCardSet *CardSet, handCardNum int) bool {
	// 当前是第一次出牌，牌型正确即可
	if nowCardSet.Type == INIT && newCardSet.Type != ERROR {
		// 三带一要特殊处理,必须是最后才能出
		if newCardSet.Type == THREE_ONE {
			return handCardNum == 4
		}
		return true
	}

	// 炸弹，检查前面是不是也是炸弹
	if newCardSet.Type == BOMB {
		if nowCardSet.Type == BOMB {
			return newCardSet.Header > nowCardSet.Header
		} else { // 炸得喵呜喵呜
			return true
		}
	} else {
		// 同类型，张数相同，头牌更大
		if newCardSet.Type == nowCardSet.Type && len(newCardSet.Cards) == len(nowCardSet.Cards) && newCardSet.Header > nowCardSet.Header {
			return true
		}
	}

	return false
}
