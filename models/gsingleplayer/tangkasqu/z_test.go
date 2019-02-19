package tangkasqu

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/vic/vic_go/datacenter"
	z "github.com/vic/vic_go/models/cardgame"
	"github.com/vic/vic_go/record"
	"github.com/vic/vic_go/zconfig"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	_ = strings.ToLower
	_ = time.Second
}

func Test02(t *testing.T) {
	deck := NewDeck(true)
	//	fmt.Println(deck)
	if len(deck) != 54 {
		t.Error()
	}
}

func Test03(t *testing.T) {
	i := 0
	mapCount := map[string]int{}
	out := float64(0)
	in := float64(0)
	for i < 1000000 {
		fcs := Deal7Cards()
		if !(fcs[0].Rank == "Z" || fcs[1].Rank == "Z" || fcs[2].Rank == "Z" || fcs[3].Rank == "Z" || fcs[4].Rank == "Z") {
			// continue
		}
		//
		//		rank := CalcRankPoker5Card(fcs)
		//		var rankName string
		//		for name, v := range z.PokerType {
		//			if v == rank[0] {
		//				rankName = name
		//			}
		//		}
		//		rankName = rankName[11:len(rankName)]
		//		fmt.Println(fcs, rankName, rank)

		//
		//		rank := CalcRankTangkasqu5Card(fcs)
		//		var rankName string
		//		for name, v := range MapTypeToInt {
		//			if v == rank[0] {
		//				rankName = name
		//			}
		//		}
		//		rankName = rankName[2:len(rankName)]
		//		if rank[0] > 0 {
		//			fmt.Println(z.SortedByRank(fcs), rankName, rank)
		//		}

		//
		nBets := rand.Int63n(4)
		rank, _, pay, _ := CalcEnding(fcs, "5", 1, nBets)
		out += pay
		in += float64(nBets)
		mapCount[rank] += 1
		if rank != T_NOTHING || true {
			// fmt.Println(z.SortedByRank(fcs), rank)
		}
		//
		i += 1
	}
	//
	fmt.Println("out", out, "in", in)
	fmt.Println("payrate", out/in, "mapCount", mapCount)
}

func Test04(t *testing.T) {
	var h0, h1, h2, h3, h4, h5, h6 z.Card
	var h []z.Card
	var s string
	h0 = z.FNewCardFS("As")
	h1 = z.Card{Rank: "Z"}
	h2 = z.FNewCardFS("Kh")
	h3 = z.FNewCardFS("Th")
	h4 = z.FNewCardFS("9h")
	h5 = z.FNewCardFS("Jh")
	h6 = z.FNewCardFS("Qh")
	h = []z.Card{h0, h1, h2, h3, h4, h5, h6}
	s, _, _ = CalcRankTangkasqu7Card(h)
	if s != T_ROYAL_FLUSH {
		t.Error()
	}
	//
	h0 = z.FNewCardFS("9s")
	h1 = z.Card{Rank: "Z", Suit: "r"}
	h2 = z.FNewCardFS("9h")
	h3 = z.FNewCardFS("Th")
	h4 = z.FNewCardFS("9d")
	h5 = z.FNewCardFS("Jh")
	h6 = z.Card{Rank: "Z", Suit: "b"}
	h = []z.Card{h0, h1, h2, h3, h4, h5, h6}
	s, _, _ = CalcRankTangkasqu7Card(h)
	if s != T_5_OF_A_KIND {
		t.Error()
	}
	h0 = z.Card{Rank: "Z"}
	h1 = z.Card{Rank: "Z"}
	h2 = z.FNewCardFS("Kh")
	h3 = z.FNewCardFS("Ts")
	h4 = z.FNewCardFS("9s")
	h5 = z.FNewCardFS("Jh")
	h6 = z.FNewCardFS("Qh")
	h = []z.Card{h0, h1, h2, h3, h4, h5, h6}
	s, _, _ = CalcRankTangkasqu7Card(h)
	if s != T_ROYAL_FLUSH {
		t.Error()
	}
}

func Test05(t *testing.T) {
}

func Test06(t *testing.T) {

}
