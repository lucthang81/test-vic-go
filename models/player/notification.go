package player

import (
	"github.com/vic/vic_go/utils"
	"sort"
	"time"
)

type Notification struct {
	toPlayer         *Player
	notificationType string
	notificationId   int64 // id of the core object of this notification
	data             map[string]interface{}
	createdAt        time.Time
	expiredAt        time.Time
}

func (notification *Notification) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	data["notification_type"] = notification.notificationType
	data["created_at"] = utils.FormatTime(notification.createdAt)
	if !notification.expiredAt.IsZero() {
		data["expired_at"] = utils.FormatTime(notification.expiredAt)
	}
	data["data"] = notification.data
	data["player_id"] = notification.toPlayer.Id()
	return data
}

type ByCreatedAt []*Notification

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return a[i].createdAt.Before(a[j].createdAt) }

type NotificationManager struct {
	alreadyFetched bool
	notifications  []*Notification
	playerId       int64
}

func NewNotificationManager() (manager *NotificationManager) {
	return &NotificationManager{
		alreadyFetched: false,
		notifications:  make([]*Notification, 0),
	}
}

func (manager *NotificationManager) fetchData() (err error) {
	if manager.alreadyFetched || manager.playerId == 0 {
		return nil
	}
	player, err := GetPlayer(manager.playerId)
	manager.notifications = make([]*Notification, 0)
	for _, friendRequest := range player.relationshipManager.getFriendRequests() {
		notification := &Notification{
			toPlayer:         player,
			notificationType: "friend_request",
			notificationId:   friendRequest.Id(),
			data:             friendRequest.SerializedData(),
			createdAt:        friendRequest.createdAt,
		}
		manager.notifications = append(manager.notifications, notification)
	}

	for _, gift := range player.giftManager.getGifts() {
		notification := &Notification{
			toPlayer:         player,
			notificationType: "gift",
			notificationId:   gift.Id(),
			data:             gift.SerializedData(),
			createdAt:        gift.createdAt,
			expiredAt:        gift.expiredAt,
		}
		manager.notifications = append(manager.notifications, notification)
	}

	sort.Sort(ByCreatedAt(manager.notifications))
	manager.alreadyFetched = true
	return nil
}

func (manager *NotificationManager) addNotificationForFriendRequest(friendRequest *FriendRequest) {
	manager.fetchData()
	notification := manager.getNotificationForFriendRequest(friendRequest)
	if notification == nil {
		notification = &Notification{
			toPlayer:         friendRequest.toPlayer,
			notificationType: "friend_request",
			notificationId:   friendRequest.Id(),
			data:             friendRequest.SerializedData(),
			createdAt:        friendRequest.createdAt,
		}
		manager.addNotification(notification)
		player, _ := GetPlayer(manager.playerId)
		if player != nil {
			player.notifyReceiveNotification(notification)
		}
	}
}

func (manager *NotificationManager) removeNotificationForFriendRequest(friendRequest *FriendRequest) {
	manager.fetchData()
	notification := manager.getNotificationForFriendRequest(friendRequest)
	if notification != nil {
		manager.removeNotification(notification)
	}
}

func (manager *NotificationManager) addNotificationForGift(gift *Gift) {
	manager.fetchData()
	notification := &Notification{
		toPlayer:         gift.toPlayer,
		notificationType: "gift",
		notificationId:   gift.Id(),
		data:             gift.SerializedData(),
		createdAt:        gift.createdAt,
		expiredAt:        gift.expiredAt,
	}
	manager.addNotification(notification)
	player, _ := GetPlayer(manager.playerId)
	if player != nil {
		player.notifyReceiveNotification(notification)
	}
}

func (manager *NotificationManager) removeNotificationForGift(gift *Gift) {
	manager.fetchData()
	notification := manager.getNotificationForGift(gift)
	if notification != nil {
		manager.removeNotification(notification)
	}
}

func (manager *NotificationManager) getNotificationListData() (data []map[string]interface{}) {
	data = make([]map[string]interface{}, 0)
	for _, notification := range manager.notifications {
		notificationData := notification.SerializedData()
		data = append(data, notificationData)
	}
	return data
}

func (manager *NotificationManager) getFriendRequestNotificationListData() (data []map[string]interface{}) {
	data = make([]map[string]interface{}, 0)
	for _, notification := range manager.notifications {
		if notification.notificationType == "friend_request" {
			notificationData := notification.SerializedData()
			data = append(data, notificationData)
		}
	}
	return data
}

func (manager *NotificationManager) getNotFriendRequestNotificationListData() (data []map[string]interface{}) {
	data = make([]map[string]interface{}, 0)
	for _, notification := range manager.notifications {
		if notification.notificationType != "friend_request" {
			notificationData := notification.SerializedData()
			data = append(data, notificationData)
		}
	}
	return data
}

func (manager *NotificationManager) getTotalNumberOfNotifications() int {
	manager.fetchData()
	return len(manager.notifications)
}

func (manager *NotificationManager) getNumberOfFriendRequestNotifications() int {
	manager.fetchData()
	counter := 0
	for _, notification := range manager.notifications {
		if notification.notificationType == "friend_request" {
			counter++

		}
	}
	return counter
}

func (manager *NotificationManager) getNumberOfOtherNotifications() int {
	manager.fetchData()
	counter := 0
	for _, notification := range manager.notifications {
		if notification.notificationType != "friend_request" {
			counter++

		}
	}
	return counter
}

func (manager *NotificationManager) getNotificationForFriendRequest(friendRequest *FriendRequest) (notification *Notification) {
	manager.fetchData()
	for _, notification := range manager.notifications {
		if notification.notificationType == "friend_request" {
			if notification.notificationId == friendRequest.Id() {
				return notification
			}
		}
	}
	return nil
}

func (manager *NotificationManager) getNotificationForGift(gift *Gift) (notification *Notification) {
	manager.fetchData()
	for _, notification := range manager.notifications {
		if notification.notificationType == "gift" {
			if notification.notificationId == gift.Id() {
				return notification
			}
		}
	}
	return nil
}

func (manager *NotificationManager) addNotification(notification *Notification) {
	manager.notifications = append(manager.notifications, notification)
}

func (manager *NotificationManager) removeNotification(notificationToRemove *Notification) {
	newNotifications := make([]*Notification, 0)
	for _, notification := range manager.notifications {
		if notification != notificationToRemove {
			newNotifications = append(newNotifications, notification)
		}
	}
	manager.notifications = newNotifications
}
