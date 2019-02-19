package bacay2

import (
	"fmt"
	//	"time"
	"testing"
	//"github.com/vic/vic_go/models/components"
)

var _, _ = fmt.Println("")

func TestConvertValueToInt(t *testing.T) {
	var actual int
	var expected int

	actual = ConvertValueToInt("5")
	expected = 5
	if actual != expected {
		t.Error()
	}

	actual = ConvertValueToInt("a")
	expected = 1
	if actual != expected {
		t.Error()
	}

	actual = ConvertValueToInt("q")
	expected = 10
	if actual != expected {
		t.Error()
	}
}

func TestFind(t *testing.T) {
	var actual int
	var expected int

	actual = Find(rankOrder, "5")
	expected = 3
	if actual != expected {
		t.Error()
	}

	actual = Find(rankOrder, "a")
	expected = 12
	if actual != expected {
		t.Error()
	}

	actual = Find(rankOrder, "j")
	expected = 9
	if actual != expected {
		t.Error()
	}
}

func TestGetMaxCard(t *testing.T) {
	var actual string
	var expected string

	actual = GetMaxCard([]string{"h 9", "d a", "c k"})
	expected = "d a"
	if actual != expected {
		t.Error()
	}

	actual = GetMaxCard([]string{"h 9", "h 8", "c k"})
	expected = "h 9"
	if actual != expected {
		t.Error()
	}
}

func TestCalcScore(t *testing.T) {
	var actual int
	var expected int

	actual, _ = CalcScore([]string{"h 9", "h 8", "c k"})
	expected = 7
	if actual != expected {
		t.Error()
	}

	actual, _ = CalcScore([]string{"h 9", "h a", "c k"})
	expected = 10
	if actual != expected {
		t.Error()
	}
}

func TestCompareTwoBacayHand(t *testing.T) {
	var actual string
	var expected string

	actual, _ = CompareTwoBacayHand([]string{"h 9", "h 8", "c 3"}, []string{"h 7", "d a", "c 2"})
	expected = "less"
	if actual != expected {
		t.Error()
	}

	actual, _ = CompareTwoBacayHand([]string{"h 9", "c 3", "s 3"}, []string{"d 7", "h 2", "c 6"})
	expected = "less"
	if actual != expected {
		t.Error()
	}

	actual, _ = CompareTwoBacayHand([]string{"h 9", "h 8", "d a"}, []string{"h 4", "d a", "c 3"})
	expected = "equal"
	if actual != expected {
		t.Error()
	}

	actual, _ = CompareTwoBacayHand([]string{"d 7", "d 6", "h 9"}, []string{"h 4", "d a", "c 3"})
	expected = "less"
	if actual != expected {
		t.Error()
	}
}
