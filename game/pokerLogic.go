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

	for i := 0; i < length; i++ {
		points[i] = p.ValueToPoints(cards[i])
	}

	// 按点数升序排序
	sort.Ints(points)

	return points
}

func (p *PokerLogic) ValueToPoints(value int) int {
	// id 1-4 对应的是3， 5-8对应4， 依此类推 45 - 48对应14(A)， 49-52对应15(2)
	if value % 4 == 0 {
		return value / 4 + 2
	} else {
		return value / 4 + 3
	}
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

// 从手牌中计算获取更大的牌型集
func (p *PokerLogic) GetLargerCardSet(cards []int, nowCardSet *CardSet) *CardSet {
	sort.Ints(cards)
	length := len(cards)
	points := p.CardsToPoints(cards)

	// 第一步, 检查炸弹, 获取最大的炸弹并比较
	maxBombPoint := p.MaxFirstPoint(points, 4)

	if maxBombPoint != 0 {
		if nowCardSet.Type != BOMB || maxBombPoint > nowCardSet.Header {
			tempCards := p.GetPointPlusCard(cards, maxBombPoint, 4, 0)

			largeCardSet := &CardSet{
				Type: BOMB,
				Header: maxBombPoint,
				Cards: tempCards,
			}

			return largeCardSet
		}
	}

	// 牌数不够了
	if length < len(nowCardSet.Cards) {
		return nil
	}

	switch nowCardSet.Type {
	case BOMB: // 碰上炸弹
		return nil
	case SINGLE:	// 单张, 取最大的一张
		maxPoint := p.MaxFirstPoint(points, 1)
		if maxPoint > nowCardSet.Header {
			tempCards := p.GetPointPlusCard(cards, maxPoint, 1, 0)

			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: maxPoint,
				Cards: tempCards,
			}

			return largeCardSet
		}
	case DOUBLE:	// 两张, 取最大的一对进行比较
		maxPoint := p.MaxFirstPoint(points, 2)
		if maxPoint > nowCardSet.Header {
			tempCards := p.GetPointPlusCard(cards, maxPoint, 2, 0)

			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: maxPoint,
				Cards: tempCards,
			}

			return largeCardSet
		}
	case THREE_TWO:	// 三带二,取最大的三张进行比较
		maxPoint := p.MaxFirstPoint(points, 3)
		if maxPoint > nowCardSet.Header {
			tempCards := p.GetPointPlusCard(cards, maxPoint, 3, 2)

			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: maxPoint,
				Cards: tempCards,
			}

			return largeCardSet
		}
	case CONNECT_DOUBLE: // 姊妹对
		hasResult, tempCards := p.GetMaxConnectCard(cards, len(nowCardSet.Cards)/2, 2, nowCardSet.Header + len(nowCardSet.Cards)/2 - 1)
		if hasResult {
			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: p.ValueToPoints(tempCards[0]),
				Cards: tempCards,
			}

			return largeCardSet
		}
	case STRAIGHT: // 顺子
		hasResult, tempCards := p.GetMaxConnectCard(cards, len(nowCardSet.Cards), 1, nowCardSet.Header + len(nowCardSet.Cards) - 1)
		if hasResult {
			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: p.ValueToPoints(tempCards[0]),
				Cards: tempCards,
			}

			return largeCardSet
		}
	case CONNECT_THREE_TWO: // 连续三带二,取最大的连续三张进行比较
		// 先取连续三张，取到了，则继续取剩下的牌
		hasResult, tempCards := p.GetMaxConnectCard(cards, len(nowCardSet.Cards) / 5, 3, nowCardSet.Header + len(nowCardSet.Cards) / 5 - 1)
		if hasResult {
			existMap := make(map[int]int)
			for _, value := range tempCards {
				existMap[value] = 1
			}

			newCards := make([]int, len(nowCardSet.Cards))

			plusNum := len(nowCardSet.Cards) / 5 * 2
			nowPlusNum := 0
			j := 0

			for i := length - 1; i >= 0; i-- {
				_, ok := existMap[cards[i]]
				if ok { // 在连续三张中，直接放入
					newCards[j] = cards[i]
					j++
				} else {
					if nowPlusNum < plusNum {
						newCards[j] = cards[i]
						j++
						nowPlusNum++
					}
				}
			}

			sort.Ints(newCards)

			largeCardSet := &CardSet{
				Type: nowCardSet.Type,
				Header: p.ValueToPoints(newCards[0]),
				Cards: newCards,
			}

			return largeCardSet
		}
		return nil
	}

	return nil
}

// 从大到小，获取连点数牌，点数数量为num, 每个点数出现pointNum次, 最大牌的点数必须大于nowMaxPoint
func (p *PokerLogic) GetMaxConnectCard(cards []int, num int, pointNum int, nowMaxPoint int)(bool, []int) {
	if nowMaxPoint >= 14 { // 不可能比A还大
		return false, make([]int, 1)
	}

	findResult := false
	tempCards := make([]int, num * pointNum)

	length := len(cards)

	// 当前已经拿到的点数数量
	nowNum := 0

	// 当前遍历点数已经拿到出现次数
	nowPointNum := 0

	// 下一个要放入的索引
	j := 0

	var tempPoint int
	for i := length - 1; i >= 0; i-- {
		// 跳过起始的2
		tempPoint = p.ValueToPoints(cards[i])
		if tempPoint == 15 {
			continue
		}

		// 首个，比较大小
		if nowNum == 0 {
			if tempPoint <= nowMaxPoint {
				break
			}

			if nowPointNum == 0 || tempPoint == p.ValueToPoints(cards[i+1]) { // 当前是第一个, 或与前一张点数相同， 则nowPointNum增加
				tempCards[j] = cards[i]
				j++
				nowPointNum++
			} else { // 与前一张点数不同， 重新开始
				nowNum = 0
				nowPointNum = 0
				j = 0
				i++
			}
		} else {
			if nowPointNum == 0 { // 需要比前面小1点的一张
				if tempPoint == p.ValueToPoints(cards[i+1]) - 1 { // 满足
					tempCards[j] = cards[i]
					j++
					nowPointNum++
				} else if tempPoint == p.ValueToPoints(cards[i+1]) { // 相同， 什么也不做

				} else { // 还原基本法
					nowNum = 0
					nowPointNum = 0
					j = 0
					i++
				}
			} else { // 需要与前面相同的一张
				if tempPoint == p.ValueToPoints(cards[i+1]) { // 满足
					tempCards[j] = cards[i]
					j++
					nowPointNum++
				} else { // 还原基本法
					nowNum = 0
					nowPointNum = 0
					j = 0
					i++
				}
			}
		}

		// 最后统一判断
		if nowPointNum >= pointNum {
			nowPointNum = 0
			nowNum++
			if nowNum >= num {
				break
			}
		}
	}

	if nowNum >= num {
		findResult = true
		sort.Ints(tempCards)
	}

	return findResult, tempCards
}

// 从大到小，获取点数等于point的num1张，其他的num2张
func (p *PokerLogic) GetPointPlusCard(cards []int, point int, num1 int, num2 int) []int {
	end := num1 + num2
	tempCards := make([]int, num1 + num2);

	nowNum1 := 0
	nowNum2 := 0

	var tempPoint int
	var j int = 0
	for i := len(cards) - 1; i >= 0; i-- {
		tempPoint = p.ValueToPoints(cards[i])
		if point == tempPoint && nowNum1 < num1 {
			tempCards[j] = cards[i]
			j++
			if j == end {
				break
			}
			nowNum1++
		} else if nowNum2 < num2 {
			tempCards[j] = cards[i]
			j++
			if j == end {
				break
			}
			nowNum2++
		}
	}

	sort.Ints(tempCards)
	return tempCards
}

// 获取最大的出现num次的点数
func (p *PokerLogic) MaxFirstPoint(points []int, num int) int {
	nowNum := 1;
	length := len(points)

	if num == 1 {
		return points[length - 1]
	}

	for i := length - 1; i > 0; i-- {
		if (points[i] == points[i-1]) { //与上一张相同，数量加1
			nowNum++;
			if nowNum >= num {
				return points[i]
			}
		} else { //重新开始计算
			nowNum = 1;
		}
	}

	return 0
}