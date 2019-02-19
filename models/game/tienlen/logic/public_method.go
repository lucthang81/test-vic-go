package logic

func ContainCards(totalCards []string, checkCards []string) bool {
	return containCards(totalCards, checkCards)
}

func CloneSlice(slice []string) []string {
	return cloneSlice(slice)
}

func IsCard1BiggerThanCard2(logic TLLogic, card1 string, card2 string) bool {
	return isCard1BiggerThanCard2(logic, card1, card2)
}

func SortCards(logic TLLogic, cards []string) []string {
	return sortCards(logic, cards)
}
