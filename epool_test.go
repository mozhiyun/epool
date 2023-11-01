package epool

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewDefault(t *testing.T) {
	c := NewDefault[int]()
	if c == nil {
		t.Fatal("NewDefault failed")
	}
	mockCount := 100000
	for i := 0; i < mockCount; i++ {
		i2 := i
		go c.AddJob(i2, func(id int64, data int) {
			// 随机1到1000毫秒
			s := time.Millisecond * time.Duration(rand.Intn(1000))
			time.Sleep(s)
			data = data + 1
			//t.Logf("id:%d, data:%d", id, data)
		})
	}
	for {
		time.Sleep(time.Millisecond)
		t.Logf("Total:%d", c.jobS.Total)
		t.Logf("Executing:%d", c.jobS.Executing)
		if c.jobS.Total == int64(mockCount) && c.jobS.Executing == 0 {
			break
		}
	}
}
