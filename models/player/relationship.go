package player

import (
	"errors"
	"fmt"
	"github.com/vic/vic_go/log"
	"github.com/vic/vic_go/utils"
	"time"
)

type RelationshipManager struct {
	alreadyFetched bool
	relationships  []*Relationship
	friendRequests []*FriendRequest
	playerId       int64
}

func NewRelationshipManager() (manager *RelationshipManager) {
	return &RelationshipManager{
		alreadyFetched: false,
		relationships:  make([]*Relationship, 0),
		friendRequests: make([]*FriendRequest, 0),
	}
}

func (manager *RelationshipManager) fetchData() (err error) {
	// get relationship
	if manager.alreadyFetched || manager.playerId == 0 {
		return nil
	}
	queryString := fmt.Sprintf("SELECT id, from_id,to_id,relationship_type FROM %s WHERE from_id = $1", RelationshipDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString, manager.playerId)
	if err != nil {
		return err
	}
	manager.relationships = make([]*Relationship, 0)
	for rows.Next() {
		var id int64
		var fromId int64
		var toId int64
		var relationshipType string
		err = rows.Scan(&id, &fromId, &toId, &relationshipType)
		if err != nil {
			rows.Close()
			return err
		}
		relationship := &Relationship{}
		relationship.id = id
		relationship.fromPlayer, err = GetPlayer(fromId)
		if err != nil {
			rows.Close()
			return err
		}
		relationship.toPlayer, err = GetPlayer(toId)
		if err != nil {
			rows.Close()
			return err
		}
		relationship.relationshipType = relationshipType
		manager.relationships = append(manager.relationships, relationship)
	}
	rows.Close()

	// get request
	queryString = fmt.Sprintf("SELECT id, from_id,to_id,created_at FROM %s WHERE to_id = $1 ORDER BY id DESC", FriendRequestDatabaseTableName)
	rows, err = dataCenter.Db().Query(queryString, manager.playerId)
	if err != nil {
		return err
	}
	manager.friendRequests = make([]*FriendRequest, 0)
	for rows.Next() {
		var id int64
		var fromId int64
		var toId int64
		var createdAt time.Time
		err = rows.Scan(&id, &fromId, &toId, &createdAt)
		if err != nil {
			rows.Close()
			return err
		}
		friendRequest := &FriendRequest{}
		friendRequest.id = id
		friendRequest.fromPlayer, err = GetPlayer(fromId)
		if err != nil {
			rows.Close()
			return err
		}
		friendRequest.toPlayer, err = GetPlayer(toId)
		if err != nil {
			rows.Close()
			return err
		}
		friendRequest.createdAt = createdAt
		manager.friendRequests = append(manager.friendRequests, friendRequest)
	}
	rows.Close()
	manager.alreadyFetched = true
	return nil
}

func (manager *RelationshipManager) addRelationship(relationship *Relationship) {
	manager.fetchData()
	manager.relationships = append(manager.relationships, relationship)
}

func (manager *RelationshipManager) removeRelationship(relationshipToRemove *Relationship) {
	manager.fetchData()
	newRelationships := make([]*Relationship, 0)
	for _, relationship := range manager.relationships {
		if relationship.Id() != relationshipToRemove.Id() {
			newRelationships = append(newRelationships, relationship)
		}
	}
	manager.relationships = newRelationships
}

func (manager *RelationshipManager) addFriendRequest(request *FriendRequest) {
	manager.fetchData()

	// have to do this first before add notification, since it can fetch the first time and got the
	// notification already, prevent it to send notify to the targeted person
	player, _ := GetPlayer(manager.playerId)
	if player != nil {
		player.notificationManager.fetchData()
	}
	manager.friendRequests = append(manager.friendRequests, request)
	if player != nil {
		player.notificationManager.addNotificationForFriendRequest(request)
	}
}

func (manager *RelationshipManager) removeFriendRequest(requestToRemove *FriendRequest) {
	manager.fetchData()
	newFriendRequests := make([]*FriendRequest, 0)
	for _, friendRequest := range manager.friendRequests {
		if friendRequest.Id() != requestToRemove.Id() {
			newFriendRequests = append(newFriendRequests, friendRequest)
		}
	}
	manager.friendRequests = newFriendRequests
	player, _ := GetPlayer(manager.playerId)
	if player != nil {
		player.notificationManager.removeNotificationForFriendRequest(requestToRemove)
	}
}

func (manager *RelationshipManager) getFriendRequestFromId(fromId int64) (friendRequest *FriendRequest) {
	manager.fetchData()
	for _, friendRequest := range manager.friendRequests {
		if friendRequest.fromPlayer.Id() == fromId {
			return friendRequest
		}
	}
	return nil
}

func (manager *RelationshipManager) getRelationship(toId int64) (relationship *Relationship) {
	manager.fetchData()
	for _, relationship := range manager.relationships {
		if relationship.toPlayer.Id() == toId {
			return relationship
		}
	}
	return nil
}

func (manager *RelationshipManager) getRelationshipDataWithPlayer(playerId int64) (data map[string]interface{}) {
	manager.fetchData()
	data = make(map[string]interface{})
	relationship := manager.getRelationship(playerId)
	if relationship == nil {
		friendRequest := manager.getFriendRequestFromId(playerId)
		if friendRequest != nil {
			data["friend_request"] = friendRequest.SerializedData()
			data["request_from_current_player"] = false
			return data
		}
		toPlayer, _ := GetPlayer(playerId)
		if toPlayer != nil {
			ourRequest := toPlayer.relationshipManager.getFriendRequestFromId(manager.playerId)
			if ourRequest != nil {
				data["friend_request"] = ourRequest.SerializedData()
				data["request_from_current_player"] = true
				return data
			}
		}
		return map[string]interface{}{}
	}
	data["relationship"] = relationship.SerializedData()
	return data
}

// friend relationship go 2 ways, so create or remove will both create/remove relationship in from and to player
func (manager *RelationshipManager) createFriendRelationship(toPlayer *Player) (err error) {
	manager.fetchData()
	relationship := manager.getRelationship(toPlayer.Id())
	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return err
	}
	if relationship != nil {
		if relationship.relationshipType == "friend" {
			return errors.New("err:already_friend")
		}
	}

	reverseRelationship := toPlayer.relationshipManager.getRelationship(manager.playerId)
	if reverseRelationship != nil {
		if reverseRelationship.relationshipType == "friend" {
			log.LogSerious("from player %d to %d, add friend error cause reverse relationship already exist, we will now just remove it", manager.playerId, toPlayer.Id())

			toPlayer.relationshipManager.removeFriendRelationship(player)
		}
	}

	relationship = &Relationship{}
	_, err = dataCenter.InsertObject(relationship,
		[]string{"from_id", "to_id", "relationship_type"},
		[]interface{}{manager.playerId, toPlayer.Id(), "friend"}, true)
	if err != nil {
		return err
	}
	relationship.fromPlayer = player
	relationship.toPlayer = toPlayer
	relationship.relationshipType = "friend"

	reverseRelationship = &Relationship{}
	_, err = dataCenter.InsertObject(reverseRelationship,
		[]string{"from_id", "to_id", "relationship_type"},
		[]interface{}{toPlayer.Id(), manager.playerId, "friend"}, true)
	if err != nil {
		return err
	}
	reverseRelationship.fromPlayer = toPlayer
	reverseRelationship.toPlayer = player
	reverseRelationship.relationshipType = "friend"

	manager.addRelationship(relationship)
	toPlayer.relationshipManager.addRelationship(reverseRelationship)
	player.notifyNumberOfFriendIncrease()
	toPlayer.notifyNumberOfFriendIncrease()
	fmt.Println("create relationship %d", len(manager.relationships))
	return nil
}

func (manager *RelationshipManager) removeFriendRelationship(toPlayer *Player) (err error) {
	manager.fetchData()
	relationship := manager.getRelationship(toPlayer.Id())
	if relationship == nil {
		return errors.New("err:not_friend_yet")
	} else {
		err = dataCenter.RemoveObject(relationship)
		if err != nil {
			return err
		}
		manager.removeRelationship(relationship)
	}

	reverseRelationship := toPlayer.relationshipManager.getRelationship(manager.playerId)
	if reverseRelationship == nil {
		log.LogSerious("from player %d to %d, remove friend error cause reverse relationship did not exist, we will now just ignore this...", manager.playerId, toPlayer.Id())
	} else {
		err = dataCenter.RemoveObject(reverseRelationship)
		if err != nil {
			return err
		}
		toPlayer.relationshipManager.removeRelationship(reverseRelationship)
	}
	return nil
}

func (manager *RelationshipManager) sendFriendRequest(toPlayerId int64) (becomeFriendInstantly bool, err error) {
	manager.fetchData()
	toPlayer, err := GetPlayer(toPlayerId)
	if err != nil {
		return false, err
	}
	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return false, err
	}

	relationship := manager.getRelationship(toPlayerId)
	if relationship != nil {
		if relationship.relationshipType == "friend" {
			return false, errors.New("err:already_friend")
		}
	}
	// check if the toPlayer already send request to current player
	reverseFriendRequest := manager.getFriendRequestFromId(toPlayer.Id())
	if reverseFriendRequest != nil {
		// just accept this and finish
		err = manager.acceptFriendRequest(toPlayer.Id())
		if err != nil {
			return false, err
		}
		return true, nil
	}

	friendRequest := toPlayer.relationshipManager.getFriendRequestFromId(manager.playerId)
	if friendRequest != nil {
		return false, errors.New("err:already_send_friend_request")
	}

	friendRequest = &FriendRequest{
		fromPlayer: player,
		toPlayer:   toPlayer,
	}
	_, err = dataCenter.InsertObject(friendRequest,
		[]string{"from_id", "to_id"},
		[]interface{}{manager.playerId, toPlayerId},
		false)
	if err != nil {
		return false, err
	}

	toPlayer.relationshipManager.addFriendRequest(friendRequest)
	return false, nil
}

func (manager *RelationshipManager) unfriend(toPlayerId int64) (err error) {
	manager.fetchData()
	toPlayer, err := GetPlayer(toPlayerId)
	if err != nil {
		return err
	}

	relationship := manager.getRelationship(toPlayerId)
	if relationship == nil {
		return errors.New("err:not_friend_yet")
	}
	return manager.removeFriendRelationship(toPlayer)
}

func (manager *RelationshipManager) acceptFriendRequest(fromPlayerId int64) (err error) {
	manager.fetchData()
	fromPlayer, err := GetPlayer(fromPlayerId)
	if err != nil {
		return err
	}
	friendRequest := manager.getFriendRequestFromId(fromPlayerId)
	if friendRequest == nil {
		return errors.New("err:friend_request_not_found")
	}

	// remove friend request
	err = dataCenter.RemoveObject(friendRequest)
	if err != nil {
		return err
	}
	manager.removeFriendRequest(friendRequest)

	err = manager.createFriendRelationship(fromPlayer)
	if err != nil {
		return err
	}
	return nil
}

func (manager *RelationshipManager) declineFriendRequest(fromPlayerId int64) (err error) {
	manager.fetchData()
	friendRequest := manager.getFriendRequestFromId(fromPlayerId)
	if friendRequest == nil {
		return errors.New("err:friend_request_not_found")
	}

	// remove friend request
	err = dataCenter.RemoveObject(friendRequest)
	if err != nil {
		return err
	}
	manager.removeFriendRequest(friendRequest)
	return nil
}

func (manager *RelationshipManager) getFriendListData() (data []map[string]interface{}) {
	manager.fetchData()
	data = make([]map[string]interface{}, 0)
	for _, relationship := range manager.relationships {
		relationshipData := relationship.SerializedData()
		data = append(data, relationshipData)
	}
	return data
}

func (manager *RelationshipManager) getNumberOfFriends() int {
	manager.fetchData()
	// currently only friend relationship
	return len(manager.relationships)
}

func (manager *RelationshipManager) getFriendRequests() (requests []*FriendRequest) {
	manager.fetchData()
	return manager.friendRequests
}

type Relationship struct {
	id               int64
	fromPlayer       *Player
	toPlayer         *Player
	relationshipType string
}

const RelationshipCacheKey string = "relationship"
const RelationshipDatabaseTableName string = "relationship"
const RelationshipClassName string = "Relationship"

func (relationship *Relationship) CacheKey() string {
	return RelationshipCacheKey
}

func (relationship *Relationship) DatabaseTableName() string {
	return RelationshipDatabaseTableName
}

func (relationship *Relationship) ClassName() string {
	return RelationshipClassName
}

func (relationship *Relationship) Id() int64 {
	return relationship.id
}

func (relationship *Relationship) SetId(id int64) {
	relationship.id = id
}

func (relationship *Relationship) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["to_player"] = relationship.toPlayer.SerializedDataWithFields([]string{"current_activity"})
	data["relationship_type"] = relationship.relationshipType
	return data
}

type FriendRequest struct {
	id         int64
	fromPlayer *Player
	toPlayer   *Player
	createdAt  time.Time
}

const FriendRequestCacheKey string = "friend_request"
const FriendRequestDatabaseTableName string = "friend_request"
const FriendRequestClassName string = "FriendRequest"

func (friendRequest *FriendRequest) CacheKey() string {
	return FriendRequestCacheKey
}

func (friendRequest *FriendRequest) DatabaseTableName() string {
	return FriendRequestDatabaseTableName
}

func (friendRequest *FriendRequest) ClassName() string {
	return FriendRequestClassName
}

func (friendRequest *FriendRequest) Id() int64 {
	return friendRequest.id
}

func (friendRequest *FriendRequest) SetId(id int64) {
	friendRequest.id = id
}

func (friendRequest *FriendRequest) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["from_player"] = friendRequest.fromPlayer.SerializedData()
	data["created_at"] = utils.FormatTime(friendRequest.createdAt)
	return data
}
