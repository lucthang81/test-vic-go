package gamemm

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/vic/vic_go/models/currency"
	"github.com/vic/vic_go/models/player"
	//	"github.com/vic/vic_go/utils"
)

const (
	ACTION_FIND_MATCH         = "ACTION_FIND_MATCH"
	ACTION_STOP_FINDING_MATCH = "ACTION_STOP_FINDING_MATCH"

	NTIMES_WAIT_FOR_BETTER_LOBBY   = 30
	DURATION_WAIT_FOR_BETTER_LOBBY = 300 * time.Millisecond
)

func init() {
	fmt.Print("")
	_ = currency.Money
}

// players choose same rule will be matched
type Rule struct {
	RequirementMoney int64
	MoneyType        string
}

// phòng chờ,
// phòng chờ đủ người sẽ bắt đầu trận đấu,
type Lobby struct {
	// number of players need to start match
	nPlayers int
	Rule     Rule

	isPlaying       bool
	MapPidToPlayers map[int64]*player.Player
}

//
type Action struct {
	ActionName   string
	PlayerId     int64
	Data         map[string]interface{}
	ChanResponse chan *ActionResponse
}

//
type ActionResponse struct {
	Err  error
	Data map[string]interface{}
}

type MatchMaker struct {
	game GameInferface
	//
	nPlayersToStartPlaying int
	// true if player is queuing, in a waitingLobby or playing a match,
	// change to true in FindLobby,
	// change to false in StopFindingLobby or FinishMatch
	MapPidToIsQueuing map[int64]bool
	//
	mapPidToLobby map[int64]*Lobby
	//
	MapPidToMatch map[int64]MatchInterface
	// đang có lệnh dừng tìm trận,
	mapPidToIsStopQueuing map[int64]bool
	//
	mapRuleToLobbies map[Rule][]*Lobby
	// for evade rematch same players
	mapPidToLastMatchPids map[int64]map[int64]*player.Player

	ChanAction chan *Action

	Mutex sync.Mutex
}

func NewMatchMaker(gamemm GameInferface, nPlayersToStartPlaying int) *MatchMaker {
	matchMaker := &MatchMaker{
		game: gamemm,
		nPlayersToStartPlaying: nPlayersToStartPlaying,

		MapPidToIsQueuing:     make(map[int64]bool),
		mapPidToLobby:         make(map[int64]*Lobby),
		MapPidToMatch:         make(map[int64]MatchInterface),
		mapPidToIsStopQueuing: make(map[int64]bool),
		mapRuleToLobbies:      make(map[Rule][]*Lobby),                  // init more in FindLobby
		mapPidToLastMatchPids: make(map[int64]map[int64]*player.Player), // init more in FindLobby
		ChanAction:            make(chan *Action),
	}

	go LoopReceiveActions(matchMaker)

	return matchMaker
}

// decrease money when starting finding match
func (mm *MatchMaker) FindMatch(player *player.Player, rule Rule) error {
	mm.Mutex.Lock()
	if mm.MapPidToIsQueuing[player.Id()] == true {
		mm.Mutex.Unlock()
		return errors.New("Bạn đang tìm trận đấu rồi.")
	} else {
		if player.GetAvailableMoney(rule.MoneyType) < rule.RequirementMoney {
			mm.Mutex.Unlock()
			return errors.New("Bạn không đủ tiền.")
		} else {
			// init vars
			mm.MapPidToIsQueuing[player.Id()] = true
			if _, isIn := mm.mapRuleToLobbies[rule]; isIn == false {
				mm.mapRuleToLobbies[rule] = make([]*Lobby, 0)
			}
			mm.Mutex.Unlock()
			//
			player.ChangeMoneyAndLog(
				-rule.RequirementMoney, rule.MoneyType, false, "",
				ACTION_FIND_MATCH, mm.game.GetGameCode(), "")
			//
			go func() {
				for i := 0; i < NTIMES_WAIT_FOR_BETTER_LOBBY; i++ {
					mm.Mutex.Lock()
					if mm.mapPidToIsStopQueuing[player.Id()] == true {
						mm.MapPidToIsQueuing[player.Id()] = false
						mm.mapPidToIsStopQueuing[player.Id()] = false
						mm.Mutex.Unlock()
						player.ChangeMoneyAndLog(
							rule.RequirementMoney, rule.MoneyType, false, "",
							ACTION_STOP_FINDING_MATCH, mm.game.GetGameCode(), "")
					} else {
						mm.Mutex.Unlock()
					}
					//
					mm.Mutex.Lock()
					readyLobby := []*Lobby{}
					// fmt.Println("hihi", rule, mm.mapRuleToLobbies[rule])
					for _, lobby := range mm.mapRuleToLobbies[rule] {
						if lobby.isPlaying == false {
							readyLobby = append(readyLobby, lobby)
						}
					}
					if len(readyLobby) == 0 {
						mm.Mutex.Unlock()
						mm.createLobby(player, rule)
						return
					} else {
						var choosenLobby *Lobby
						for _, lobby := range readyLobby {
							if mm.checkIsRematch(player, lobby) == false {
								choosenLobby = lobby
							}
						}
						if i == NTIMES_WAIT_FOR_BETTER_LOBBY-1 && choosenLobby == nil {
							choosenLobby = readyLobby[0]
						}
						mm.Mutex.Unlock()

						if choosenLobby != nil {
							mm.joinLobby(player, choosenLobby)
							return
						}
					}
					time.Sleep(DURATION_WAIT_FOR_BETTER_LOBBY)
				}
			}()
			return nil
		}
	}
}

func (mm *MatchMaker) StopFindingMatch(player *player.Player) error {
	var err error
	err = nil
	isNeedRefund := false
	var rule Rule
	mm.Mutex.Lock()
	if mm.MapPidToIsQueuing[player.Id()] == false {
		err = errors.New("Không thể dừng tìm trận, bạn chưa tìm trận.")
	} else {
		currentLobby := mm.mapPidToLobby[player.Id()]
		if currentLobby == nil {
			mm.mapPidToIsStopQueuing[player.Id()] = true
		} else {
			if currentLobby.isPlaying == true {
				err = errors.New("Không thể dừng tìm trận, bạn đang chơi rồi.")
			} else {
				delete(currentLobby.MapPidToPlayers, player.Id())
				delete(mm.mapPidToLobby, player.Id())
				mm.MapPidToIsQueuing[player.Id()] = false
				isNeedRefund = true
				rule = currentLobby.Rule
			}
		}
	}
	mm.Mutex.Unlock()
	if isNeedRefund {
		player.ChangeMoneyAndLog(
			rule.RequirementMoney, rule.MoneyType, false, "",
			ACTION_STOP_FINDING_MATCH, mm.game.GetGameCode(), "")
	}
	return err
}

func (mm *MatchMaker) createLobby(playerObj *player.Player, rule Rule) error {
	mm.Mutex.Lock()
	lobby := &Lobby{
		isPlaying: false,
		Rule:      rule,
		nPlayers:  mm.nPlayersToStartPlaying,
		MapPidToPlayers: map[int64]*player.Player{
			playerObj.Id(): playerObj,
		},
	}
	mm.mapPidToLobby[playerObj.Id()] = lobby
	mm.mapRuleToLobbies[rule] = append(mm.mapRuleToLobbies[rule], lobby)
	mm.Mutex.Unlock()
	//
	mm.Mutex.Lock()
	// fmt.Println("createLobby mm.mapRuleToLobbies", mm.mapRuleToLobbies)
	//	fmt.Println("createLobby mm.mapRuleToLobbies")
	//	for k, v := range mm.mapRuleToLobbies {
	//		fmt.Println("k, len(v)", k, len(v))
	//	}
	//fmt.Println("createLobby mm.mapPidToLobby", mm.mapPidToLobby)
	//fmt.Println("createLobby mm.MapPidToMatch", mm.MapPidToMatch)
	mm.Mutex.Unlock()
	return nil
}

// include start match
func (mm *MatchMaker) joinLobby(player *player.Player, lobby *Lobby) error {
	mm.Mutex.Lock()
	//	fmt.Println("(mm *MatchMaker) joinLobby", player.Id(), lobby)
	lobby.MapPidToPlayers[player.Id()] = player
	mm.mapPidToLobby[player.Id()] = lobby
	if len(lobby.MapPidToPlayers) == lobby.nPlayers {
		lobby.isPlaying = true
		mm.Mutex.Unlock()
		match := mm.game.StartMatch(lobby)
		mm.Mutex.Lock()
		for pid, _ := range lobby.MapPidToPlayers {
			mm.MapPidToMatch[pid] = match
		}
		mm.Mutex.Unlock()
	} else {
		mm.Mutex.Unlock()
	}

	//
	mm.Mutex.Lock()
	// fmt.Println("createLobby mm.mapRuleToLobbies", mm.mapRuleToLobbies)
	//	fmt.Println("createLobby mm.mapRuleToLobbies")
	//	for k, v := range mm.mapRuleToLobbies {
	//		fmt.Println("k, len(v)", k, len(v))
	//	}
	//fmt.Println("createLobby mm.mapPidToLobby", mm.mapPidToLobby)
	//fmt.Println("createLobby mm.MapPidToMatch", mm.MapPidToMatch)
	mm.Mutex.Unlock()
	return nil
}

// call this func when finish the match,
func (mm *MatchMaker) ClearLobby(lobby *Lobby) error {
	mm.Mutex.Lock()
	for pid, _ := range lobby.MapPidToPlayers {
		mm.MapPidToIsQueuing[pid] = false
		delete(mm.MapPidToMatch, pid)
		delete(mm.mapPidToLobby, pid)
		mm.mapPidToLastMatchPids[pid] = make(map[int64]*player.Player)
		for k, v := range lobby.MapPidToPlayers {
			mm.mapPidToLastMatchPids[pid][k] = v
		}
	}

	var i int
	for i = 0; i < len(mm.mapRuleToLobbies[lobby.Rule]); i++ {
		if lobby == mm.mapRuleToLobbies[lobby.Rule][i] {
			break
		}
	}
	//
	copy(mm.mapRuleToLobbies[lobby.Rule][i:], mm.mapRuleToLobbies[lobby.Rule][i+1:])
	mm.mapRuleToLobbies[lobby.Rule][len(mm.mapRuleToLobbies[lobby.Rule])-1] = nil
	mm.mapRuleToLobbies[lobby.Rule] = mm.mapRuleToLobbies[lobby.Rule][:len(mm.mapRuleToLobbies[lobby.Rule])-1]
	//
	mm.Mutex.Unlock()
	return nil
}

// need to lock mm when use this func
func (mm *MatchMaker) checkIsRematch(player *player.Player, lobby *Lobby) bool {
	isRematch := false
	lastMatchPids := mm.mapPidToLastMatchPids[player.Id()]
	for pidInLobby, _ := range lobby.MapPidToPlayers {
		if _, isIn := lastMatchPids[pidInLobby]; isIn == true {
			isRematch = true
			break
		}
	}
	return isRematch
}

func LoopReceiveActions(matchMaker *MatchMaker) {
	for {
		action := <-matchMaker.ChanAction
		go func() {
			defer func() {
				if r := recover(); r != nil {
					bytes := debug.Stack()
					fmt.Println("ERROR ERROR ERROR: ", r, string(bytes))
				}
			}()

			playerObj, err := player.GetPlayer(action.PlayerId)
			if err != nil {
				action.ChanResponse <- &ActionResponse{Err: errors.New("Cant find player for this id")}
			} else {
				if action.ActionName == ACTION_FIND_MATCH {
					rule := matchMaker.game.GetRuleForPlayer(action.PlayerId)
					err := matchMaker.FindMatch(playerObj, rule)
					action.ChanResponse <- &ActionResponse{Err: err}
				} else if action.ActionName == ACTION_STOP_FINDING_MATCH {
					err := matchMaker.StopFindingMatch(playerObj)
					action.ChanResponse <- &ActionResponse{Err: err}
				} else { // các hành động của riêng từng game
					matchMaker.game.ReceiveAction(action)
				}
			}
		}()
	}
}
