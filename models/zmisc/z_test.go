package zmisc

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	PushAll(fmt.Sprintf(
		"Đừng quên từ 8h đến 10h vào %v và nạp tiền "+
			"để nhận thêm 30%% Kim! Chúc may mắn!", CLIENT_NAME_PH))
	fmt.Println("hoho")
}
