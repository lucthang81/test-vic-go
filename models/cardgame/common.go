package cardgame

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	IS_DEBUGING = false
)

func init() {
	fmt.Print("")
	_ = time.Now()
	_ = errors.New("")
}

func Print(a ...interface{}) {
	if IS_DEBUGING {
		fmt.Println(a...)
	} else {

	}
}

//
type SizedList struct {
	MaxLen   int
	Elements []string
}

func NewSizedList(MaxLen int) SizedList {
	return SizedList{
		MaxLen:   MaxLen,
		Elements: make([]string, 0),
	}
}

func (list *SizedList) Append(newE string) {
	list.Elements = append(list.Elements, newE)
	if len(list.Elements) > list.MaxLen {
		list.Elements = list.Elements[len(list.Elements)-list.MaxLen:]
	}
}

// using for get all combinations nCk from list
//    choicesLI: lower bound index for choices
//    choicesUI: upper bound index for choices
//    comb: save a combination
func backTrack(
	combChan chan []int, k int, list []int,
	comb []int, stepLeft int, choicesLI int, choicesUI int) {
	if stepLeft == 0 {
		clonedComb := make([]int, len(comb))
		copy(clonedComb, comb)
		Print("combChan <- clonedComb", clonedComb)
		combChan <- clonedComb
		if k == 0 {
			Print("close(combChan) 1")
			close(combChan)
		}
	} else {
		Print("stepLeft", stepLeft, "choices", list[choicesLI:choicesUI])
		for i := choicesLI; i < choicesUI; i++ {
			comb[k-stepLeft] = list[i]
			Print(
				"choice", list[i],
				"newCurrentResult", comb[0:k-stepLeft+1],
				"newStepLeft", stepLeft-1,
				"newChoices", list[i+1:choicesUI])
			backTrack(
				combChan, k, list,
				comb, stepLeft-1, i+1, choicesUI)
		}
		Print("stepLeft", stepLeft, "completed")
		if stepLeft == k {
			// end the recursion when finish 1st step loop
			Print("close(combChan) 2")
			close(combChan)
		}
	}
}

//
func backTrack2(
	combChan chan []int, k int, list []int,
	comb []int, stepLeft int, choicesLI int, choicesUI int,
	deadline time.Time) {
	if stepLeft == 0 {
		clonedComb := make([]int, len(comb))
		copy(clonedComb, comb)
		Print("combChan <- clonedComb", clonedComb)
		combChan <- clonedComb
		if k == 0 {
			Print("close(combChan) 1")
			close(combChan)
		}
	} else {
		Print("stepLeft", stepLeft, "choices", list[choicesLI:choicesUI])
		for i := choicesLI; i < choicesUI; i++ {
			if time.Now().After(deadline) {
				break
			}
			comb[k-stepLeft] = list[i]
			Print(
				"choice", list[i],
				"newCurrentResult", comb[0:k-stepLeft+1],
				"newStepLeft", stepLeft-1,
				"newChoices", list[i+1:choicesUI])
			backTrack2(
				combChan, k, list,
				comb, stepLeft-1, i+1, choicesUI, deadline)
		}
		Print("stepLeft", stepLeft, "completed")
		if stepLeft == k {
			// end the recursion when finish 1st step loop
			Print("close(combChan) 2")
			close(combChan)
		}
	}
}

//
func GetCombinationToChan(list []int, k int, combChan chan []int) {
	backTrack(
		combChan, k, list,
		make([]int, k), k, 0, len(list))
}

//
func GetCombinationToChan2(
	list []int, k int, combChan chan []int, deadline time.Time) {
	backTrack2(
		combChan, k, list,
		make([]int, k), k, 0, len(list), deadline)
}

//
func GetCombinations(list []int, k int) [][]int {
	result := make([][]int, 0, 0)
	if k > len(list) {
		return result
	}
	combChan := make(chan []int)
	go GetCombinationToChan(list, k, combChan)
	for {
		comb, IsOpening := <-combChan
		if IsOpening {
			result = append(result, comb)
		} else {
			break
		}
	}
	return result
}

//
func GetCombinations2(list []int, k int, deadline time.Time) [][]int {
	result := make([][]int, 0, 0)
	if k > len(list) {
		return result
	}
	combChan := make(chan []int)
	go GetCombinationToChan2(list, k, combChan, deadline)
	for {
		comb, IsOpening := <-combChan
		if IsOpening {
			result = append(result, comb)
		} else {
			break
		}
	}
	return result
}

func GetCombinationsForStrings(ss []string, k int) [][]string {
	result := make([][]string, 0)
	iCombs := GetCombinations(Range(len(ss)), k)
	for _, iComb := range iCombs {
		comb := make([]string, k)
		for i, ci := range iComb {
			comb[i] = ss[ci]
		}
		result = append(result, comb)
	}
	return result
}

// return [0, 1, 2, .., n-1]
func Range(n int) []int {
	result := make([]int, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindStringInSlice(sub string, list []string) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindInt64InSlice(sub int64, list []int64) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// Return the lowest index of arg1 in where arg0 is found,
// If not found return -1
func FindIntInSlice(sub int, list []int) int {
	for index, element := range list {
		if sub == element {
			return index
		}
	}
	return -1
}

// sub two slice in64
func SubtractedInt64s(fullSet []int64, subSet []int64) []int64 {
	result := make([]int64, 0, len(fullSet))
	for _, card := range fullSet {
		if FindInt64InSlice(card, subSet) == -1 {
			result = append(result, card)
		}
	}
	return result
}

// Return the number of k-combinations of n
func CalcNOCombs(in int, ik int) int {
	if !(in > 0) {
		return 0
	}
	if !((0 <= ik) && (ik <= in)) {
		return 0
	}
	n := int64(in)
	k := int64(ik)
	upProduct := int64(1)
	downProduct := int64(1)
	for i := int64(1); i <= k; i++ {
		upProduct *= n - k + i
		downProduct *= i
	}
	return int(upProduct / downProduct)
}

// compare two slice int;
// return true if arr1 >= arr2
func Compare2ListInt(arr1 []int, arr2 []int) bool {
	// min len(arr1), len(arr2)
	var n int
	if len(arr1) <= len(arr2) {
		n = len(arr1)
	} else {
		n = len(arr2)
	}
	for i := 0; i < n; i++ {
		if arr1[i] < arr2[i] {
			return false
		} else if arr1[i] > arr2[i] {
			return true
		} else { // arr1[i] == arr2[i]
			continue
		}
	}
	// arr1[i] == arr2[i] foreach i = 0..n-1
	if len(arr1) >= len(arr2) {
		return true
	} else {
		return false
	}
}

func HumanFormatNumber(n interface{}) string {
	var s string
	if nInt64, isOk := n.(int64); isOk {
		s = fmt.Sprintf("%v", nInt64)
	}
	if nStr, isOk := n.(string); isOk {
		s = nStr
	}
	if nFloat64, isOk := n.(float64); isOk {
		s = fmt.Sprintf("%.0f", nFloat64)
	}
	// list chars of result
	temp := []string{}
	seperator := ","
	for i := len(s) - 1; i >= 0; i-- {
		temp = append(temp, string(s[i]))
		if ((len(s)-i)%3 == 0) && i > 0 {
			temp = append(temp, seperator)
		}
	}
	reverse := []string{}
	for i := len(temp) - 1; i >= 0; i-- {
		reverse = append(reverse, temp[i])
	}
	return strings.Join(reverse, "")
}

// Return the index where to insert item x in list a, assuming a is sorted desc.
// The return value i is such that:
// all e in a[:i] have e >= x, and all e in a[i:] have e < x.
// Optional args lo (default 0) and hi (default len(a)) bound the
// slice of a to be searched.
func bisectRight(a []int64, x int64, lo int, hi int) (int, error) {
	if a == nil {
		return 0, errors.New("a is nil")
	}
	if lo < 0 {
		return 0, errors.New("lo must be non-negative")
	}
	for lo < hi {
		mid := (lo + hi) / 2
		if x > a[mid] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo, nil
}

func BisectRight(a []int64, x int64) (int, error) {
	return bisectRight(a, x, 0, len(a))
}

// shuffle the input array
func ShuffleInts(ints []int) {
	cards := ints
	n := len(cards)
	for i := n - 1; i >= 1; i-- {
		j := rand.Intn(i + 1) // 0 <= j <=i
		temp := cards[i]
		cards[i] = cards[j]
		cards[j] = temp
	}
}

// shuffle the input array
func ShuffleFloat64s(float64s []float64) {
	cards := float64s
	n := len(cards)
	for i := n - 1; i >= 1; i-- {
		j := rand.Intn(i + 1) // 0 <= j <=i
		temp := cards[i]
		cards[i] = cards[j]
		cards[j] = temp
	}
}
