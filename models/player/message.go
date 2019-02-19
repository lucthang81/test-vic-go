package player

import (
	"encoding/json"
	"fmt"
	"github.com/vic/vic_go/utils"
	"time"
)

const (
	MESSAGE_TYPE_2 = "type2"
)

type Message struct {
	id          int64
	toPlayer    *Player
	messageType string
	data        map[string]interface{}
	status      string

	money int64

	createdAt time.Time
}

const MessageCacheKey string = "message"
const MessageDatabaseTableName string = "message"
const MessageClassName string = "Message"

func (message *Message) CacheKey() string {
	return MessageCacheKey
}

func (message *Message) DatabaseTableName() string {
	return MessageDatabaseTableName
}

func (message *Message) ClassName() string {
	return MessageClassName
}

func (message *Message) Id() int64 {
	return message.id
}

func (message *Message) SetId(id int64) {
	message.id = id
}

func (message *Message) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["id"] = message.Id()
	data["message_type"] = message.messageType
	data["data"] = message.data
	data["created_at"] = utils.FormatTime(message.createdAt)
	data["status"] = message.status
	return data
}

type MessageManager struct {
	playerId int64
}

func NewMessageManager() (manager *MessageManager) {
	return &MessageManager{}
}

func (manager *MessageManager) getData(limit int64, offset int64) (
	results []map[string]interface{}, total int64, err error) {
	// limit the limit to 30
	limit = utils.MinInt64(5000, limit)

	queryString := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE to_id = $1", MessageDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, manager.playerId)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return nil, 0, err
	}
	queryString = fmt.Sprintf("SELECT id,to_id, message_type, data, status, created_at"+
		" FROM %s"+
		" WHERE to_id = $1"+
		" ORDER BY -id LIMIT $2 OFFSET $3", MessageDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString, manager.playerId, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var toId int64
		var messageType string
		var status string
		var dataString []byte
		var createdAt time.Time
		err = rows.Scan(&id, &toId, &messageType, &dataString, &status, &createdAt)
		if err != nil {
			return nil, 0, err
		}
		var data map[string]interface{}
		err := json.Unmarshal(dataString, &data)
		if err != nil {
			return nil, 0, err
		}
		message := &Message{}
		message.id = id
		message.messageType = messageType
		message.toPlayer = player
		message.data = data
		message.status = status
		message.createdAt = createdAt

		results = append(results, message.SerializedData())
	}
	rows.Close()

	return results, total, nil
}

func (manager *MessageManager) getDataByType(
	limit int64, offset int64, msgType string) (
	results []map[string]interface{}, total int64, err error) {
	limit = utils.MinInt64(5000, limit)

	queryString := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE to_id = $1", MessageDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, manager.playerId)
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return nil, 0, err
	}
	queryString = fmt.Sprintf("SELECT id,to_id, message_type, data, status, created_at"+
		" FROM %s"+
		" WHERE to_id = $1 AND message_type = $4 "+
		" ORDER BY -id LIMIT $2 OFFSET $3", MessageDatabaseTableName)
	rows, err := dataCenter.Db().Query(queryString,
		manager.playerId, limit, offset, msgType)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results = make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var toId int64
		var messageType string
		var status string
		var dataString []byte
		var createdAt time.Time
		err = rows.Scan(&id, &toId, &messageType, &dataString, &status, &createdAt)
		if err != nil {
			return nil, 0, err
		}
		var data map[string]interface{}
		err := json.Unmarshal(dataString, &data)
		if err != nil {
			return nil, 0, err
		}
		message := &Message{}
		message.id = id
		message.messageType = messageType
		message.toPlayer = player
		message.data = data
		message.status = status
		message.createdAt = createdAt

		results = append(results, message.SerializedData())
	}
	rows.Close()

	return results, total, nil
}

func (manager *MessageManager) createMessage(messageType string, additionalData map[string]interface{}) (message *Message, err error) {
	player, err := GetPlayer(manager.playerId)
	if err != nil {
		return nil, err
	}
	message = &Message{
		toPlayer:    player,
		messageType: messageType,
		data:        additionalData,
		createdAt:   time.Now(),
	}

	dataString, _ := json.Marshal(additionalData)

	_, err = dataCenter.InsertObject(message,
		[]string{"to_id", "message_type", "data"},
		[]interface{}{player.id, messageType, dataString}, true)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (manager *MessageManager) markReadAllMessages() (err error) {
	queryString := fmt.Sprintf("UPDATE %s SET status = $1 WHERE to_id = $2", MessageDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, "read", manager.playerId)
	return err
}

func (manager *MessageManager) markRead1Message(msgId int64) (err error) {
	queryString := fmt.Sprintf("UPDATE %s SET status = $1 WHERE to_id = $2 and id = $3", MessageDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, "read", manager.playerId, msgId)
	return err
}

func (manager *MessageManager) delete1Message(msgId int64) (err error) {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE id = $1", MessageDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, msgId)
	return err
}

func (manager *MessageManager) getUnreadCount() (count int64, err error) {
	queryString := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE to_id = $1 AND status = $2", MessageDatabaseTableName)
	row := dataCenter.Db().QueryRow(queryString, manager.playerId, "unread")
	err = row.Scan(&count)
	return count, err
}

/*
create some default message
*/

func (manager *MessageManager) createPaymentAcceptedMessage(id int64, cardCode string, serialCode string, cardNumber string) (err error) {
	data := make(map[string]interface{})
	data["id"] = id
	data["serial_code"] = serialCode
	data["card_number"] = cardNumber
	data["card_code"] = cardCode
	_, err = manager.createMessage("payment_accepted", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

func (manager *MessageManager) createPaymentDeclinedMessage(id int64, cardCode string) (err error) {
	data := make(map[string]interface{})
	data["card_code"] = cardCode
	data["id"] = id
	_, err = manager.createMessage("payment_declined", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

func (manager *MessageManager) createPurchaseMessage(serialCode string, cardNumber string, addedMoney int64, currentMoney int64) (err error) {
	data := make(map[string]interface{})
	data["serial_code"] = serialCode
	data["card_number"] = cardNumber
	data["added_money"] = addedMoney
	data["money"] = currentMoney
	_, err = manager.createMessage("purchase_message", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

// server send to player
func (manager *MessageManager) createRawMessage(title string, content string) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["content"] = content
	_, err = manager.createMessage("raw", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

// tin nhắn đổi thưởng
func (manager *MessageManager) createType2Message(title string, content string) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["content"] = content
	_, err = manager.createMessage(MESSAGE_TYPE_2, data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

// sender (other player) send to player
func (manager *MessageManager) createRawMessage2(
	title string, content string, sender *Player) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["content"] = content
	if sender != nil {
		data["senderId"] = sender.Id()
		data["senderDisplayName"] = sender.DisplayName()
	}
	_, err = manager.createMessage("raw", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

// message include some commands,
// user can click to execute the commands,
// reactingData = map "methodName" to mapParams
func (manager *MessageManager) createReactingMessage(
	title string, content string, reactingData map[string]interface{}) (
	err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["content"] = content
	data["isReactingMessage"] = true
	data["reactingData"] = reactingData
	_, err = manager.createMessage("raw", data)
	if err != nil {
		// log.LogSerious("create register message error %v", err)
		return err
	}
	return nil
}

func createRawMessageToAllPlayers(title string, content string) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["content"] = content
	dataString, _ := json.Marshal(data)
	queryString := fmt.Sprintf("INSERT INTO %s (to_id, message_type, data) "+
		"SELECT id as to_id, $1, $2 FROM %s", MessageDatabaseTableName, PlayerDatabaseTableName)
	_, err = dataCenter.Db().Exec(queryString, "raw", dataString)
	return err
}
