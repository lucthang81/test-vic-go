package zhelp

// return the next element in the list, next(n-1) = 0
func GetNextPlayer(element int64, array []int64) int64 {
	if len(array) == 0 {
		panic("len(array) == 0")
	} else {
		i := int(-1)
		for k, _ := range array {
			if array[k] == element {
				i = k
				break
			}
		}
		if i == -1 {
			panic("i == -1")
		} else if i == len(array)-1 {
			return array[0]
		} else {
			return array[i+1]
		}
	}
}

// return the previous element in the list prev(0) = n-1
func GetPrevPlayer(element int64, array []int64) int64 {
	if len(array) == 0 {
		panic("len(array) == 0")
	} else {
		i := int(-1)
		for k, _ := range array {
			if array[k] == element {
				i = k
				break
			}
		}
		if i == -1 {
			panic("i == -1")
		} else if i == 0 {
			return array[len(array)-1]
		} else {
			return array[i-1]
		}
	}
}

// return the next element in the list, next(n-1) = 0
func GetNextSeat(element int, array []int) int {
	if len(array) == 0 {
		panic("len(array) == 0")
	} else {
		i := int(-1)
		for k, _ := range array {
			if array[k] == element {
				i = k
				break
			}
		}
		if i == -1 {
			panic("i == -1")
		} else if i == len(array)-1 {
			return array[0]
		} else {
			return array[i+1]
		}
	}
}

// return the previous element in the list prev(0) = n-1
func GetPrevSeat(element int, array []int) int {
	if len(array) == 0 {
		panic("len(array) == 0")
	} else {
		i := int(-1)
		for k, _ := range array {
			if array[k] == element {
				i = k
				break
			}
		}
		if i == -1 {
			panic("i == -1")
		} else if i == 0 {
			return array[len(array)-1]
		} else {
			return array[i-1]
		}
	}
}
