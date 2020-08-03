package pubsub

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSubscriber struct {
	received []string
	receiveCount int
	receiveLimit int
	wg *sync.WaitGroup
	t *testing.T
}

func (s *testSubscriber) OnData(data string) bool {
	s.received = append(s.received, data)
	s.receiveCount++
	if s.receiveCount >= s.receiveLimit {
		s.wg.Done()
		return true
	}
	return false
}

type testdata struct {
	pubSub PubSub
	data []string
	topic string
	receiver1Limit int
	receiver2Limit int
	receiver1 *testSubscriber
	receiver2 *testSubscriber
}

func setup(t *testing.T, wg *sync.WaitGroup) testdata {
	td := testdata{
		pubSub: NewPubSub(),
		data: []string{"data 1", "data 2", "data 3", "data 4", "data 5", "data 6", "data 7", "date 8"},
		topic: "test-topic",
		receiver1Limit: 3,
		receiver2Limit: 4,
		receiver1: &testSubscriber{receiveLimit: 3, wg: wg},
		receiver2: &testSubscriber{receiveLimit: 4, wg: wg},
	}
	return td
}

func TestPublish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var wg sync.WaitGroup
		td := setup(t, &wg)
		wg.Add(2)
		td.pubSub.Subscribe(td.topic, td.receiver1)
		td.pubSub.Subscribe(td.topic, td.receiver2)

		for i := 0; i < 6; i++ {
			td.pubSub.Publish(td.topic, td.data[i])
		}

		wg.Wait()
		assert.Equal(t, td.receiver1Limit, td.receiver1.receiveCount, "rcvr 1 item count check")
		assert.EqualValues(t, td.data[:td.receiver1Limit], td.receiver1.received, "rcvr 1 item content check")
		assert.Equal(t, td.receiver2Limit, td.receiver2.receiveCount, "rcvr 2 item count check")
		assert.EqualValues(t, td.data[:td.receiver2Limit], td.receiver2.received, "rcfr 2 item content check")
	})
}
