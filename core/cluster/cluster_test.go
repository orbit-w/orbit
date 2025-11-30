package cluster

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNodeState(t *testing.T) {
	t.Run("节点状态常量", func(t *testing.T) {
		assert.Equal(t, NodeState(0), NodeStateUnknown)
		assert.Equal(t, NodeState(1), NodeStateOnline)
		assert.Equal(t, NodeState(2), NodeStateOffline)
		assert.Equal(t, NodeState(3), NodeStateMaintenance)
	})
}

func TestNode_SetState(t *testing.T) {
	t.Run("设置节点状态", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-1",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.SetState(NodeStateOnline)
		assert.Equal(t, NodeStateOnline, node.GetState())

		node.SetState(NodeStateOffline)
		assert.Equal(t, NodeStateOffline, node.GetState())

		node.SetState(NodeStateMaintenance)
		assert.Equal(t, NodeStateMaintenance, node.GetState())
	})

	t.Run("并发设置状态", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-2",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(state NodeState) {
				defer wg.Done()
				node.SetState(state)
			}(NodeState(i % 4))
		}
		wg.Wait()

		// 最终状态应该是最后一次设置的状态（虽然不确定是哪个）
		state := node.GetState()
		assert.True(t, state >= NodeStateUnknown && state <= NodeStateMaintenance)
	})
}

func TestNode_GetState(t *testing.T) {
	t.Run("获取初始状态", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-3",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		// 未设置状态时，应该返回零值
		assert.Equal(t, NodeState(0), node.GetState())
	})

	t.Run("获取设置后的状态", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-4",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.SetState(NodeStateOnline)
		assert.Equal(t, NodeStateOnline, node.GetState())
	})
}

func TestNode_SetLastHeartbeat(t *testing.T) {
	t.Run("设置最后心跳时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-5",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		now := time.Now()
		node.SetLastHeartbeat(now)
		heartbeat := node.GetLastHeartbeat()
		assert.True(t, heartbeat.Equal(now) || heartbeat.After(now.Add(-time.Millisecond)))
	})

	t.Run("并发设置心跳时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-6",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.SetLastHeartbeat(time.Now())
			}()
		}
		wg.Wait()

		// 应该能正常获取心跳时间
		heartbeat := node.GetLastHeartbeat()
		assert.False(t, heartbeat.IsZero())
	})
}

func TestNode_GetLastHeartbeat(t *testing.T) {
	t.Run("获取未设置的心跳时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-7",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		heartbeat := node.GetLastHeartbeat()
		assert.True(t, heartbeat.IsZero())
	})

	t.Run("获取设置后的心跳时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-8",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		expectedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		node.SetLastHeartbeat(expectedTime)
		heartbeat := node.GetLastHeartbeat()
		assert.True(t, heartbeat.Equal(expectedTime))
	})
}

func TestNode_SetUpdatedAt(t *testing.T) {
	t.Run("设置更新时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-9",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		now := time.Now()
		node.SetUpdatedAt(now)
		updatedAt := node.GetUpdatedAt()
		assert.True(t, updatedAt.Equal(now) || updatedAt.After(now.Add(-time.Millisecond)))
	})

	t.Run("并发设置更新时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-10",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.SetUpdatedAt(time.Now())
			}()
		}
		wg.Wait()

		updatedAt := node.GetUpdatedAt()
		assert.False(t, updatedAt.IsZero())
	})
}

func TestNode_GetUpdatedAt(t *testing.T) {
	t.Run("获取未设置的更新时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-11",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		updatedAt := node.GetUpdatedAt()
		assert.True(t, updatedAt.IsZero())
	})

	t.Run("获取设置后的更新时间", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-12",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		expectedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		node.SetUpdatedAt(expectedTime)
		updatedAt := node.GetUpdatedAt()
		assert.True(t, updatedAt.Equal(expectedTime))
	})
}

func TestNode_CCU(t *testing.T) {
	t.Run("增加CCU", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-13",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddCCU(10)
		assert.Equal(t, int32(10), node.GetCCU())

		node.AddCCU(5)
		assert.Equal(t, int32(15), node.GetCCU())
	})

	t.Run("减少CCU", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-14",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddCCU(20)
		node.SubCCU(5)
		assert.Equal(t, int32(15), node.GetCCU())

		node.SubCCU(10)
		assert.Equal(t, int32(5), node.GetCCU())
	})

	t.Run("并发操作CCU", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-15",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		operations := 1000
		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.AddCCU(1)
			}()
		}
		wg.Wait()

		assert.Equal(t, int32(operations), node.GetCCU())
	})

	t.Run("CCU可以为负数", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-16",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.SubCCU(10)
		assert.Equal(t, int32(-10), node.GetCCU())
	})
}

func TestNode_EntityCount(t *testing.T) {
	t.Run("增加实体数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-17",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddEntityCount(100)
		assert.Equal(t, int64(100), node.GetEntityCount())

		node.AddEntityCount(50)
		assert.Equal(t, int64(150), node.GetEntityCount())
	})

	t.Run("减少实体数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-18",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddEntityCount(200)
		node.SubEntityCount(50)
		assert.Equal(t, int64(150), node.GetEntityCount())
	})

	t.Run("实体数量加1", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-19",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.IncrEntityCount()
		assert.Equal(t, int64(1), node.GetEntityCount())

		node.IncrEntityCount()
		assert.Equal(t, int64(2), node.GetEntityCount())
	})

	t.Run("实体数量减1", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-20",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddEntityCount(10)
		node.DecrEntityCount()
		assert.Equal(t, int64(9), node.GetEntityCount())

		node.DecrEntityCount()
		assert.Equal(t, int64(8), node.GetEntityCount())
	})

	t.Run("并发操作实体数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-21",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		operations := 1000
		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.IncrEntityCount()
			}()
		}
		wg.Wait()

		assert.Equal(t, int64(operations), node.GetEntityCount())
	})

	t.Run("实体数量可以为负数", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-22",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.SubEntityCount(100)
		assert.Equal(t, int64(-100), node.GetEntityCount())
	})
}

func TestNode_ZoneCount(t *testing.T) {
	t.Run("增加服务区数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-23",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddZoneCount(5)
		assert.Equal(t, int32(5), node.GetZoneCount())

		node.AddZoneCount(3)
		assert.Equal(t, int32(8), node.GetZoneCount())
	})

	t.Run("减少服务区数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-24",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.AddZoneCount(10)
		node.SubZoneCount(3)
		assert.Equal(t, int32(7), node.GetZoneCount())

		node.SubZoneCount(2)
		assert.Equal(t, int32(5), node.GetZoneCount())
	})

	t.Run("并发操作服务区数量", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-25",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		var wg sync.WaitGroup
		operations := 500
		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.AddZoneCount(1)
			}()
		}
		wg.Wait()

		assert.Equal(t, int32(operations), node.GetZoneCount())
	})

	t.Run("服务区数量可以为负数", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-26",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		node.SubZoneCount(5)
		assert.Equal(t, int32(-5), node.GetZoneCount())
	})
}

func TestNode_Complete(t *testing.T) {
	t.Run("完整节点操作流程", func(t *testing.T) {
		node := &Node{
			ID:        "test-node-complete",
			Address:   "127.0.0.1:8080",
			CreatedAt: time.Now(),
		}

		// 设置状态
		node.SetState(NodeStateOnline)
		assert.Equal(t, NodeStateOnline, node.GetState())

		// 设置心跳和更新时间
		now := time.Now()
		node.SetLastHeartbeat(now)
		node.SetUpdatedAt(now)

		// 增加各种计数
		node.AddCCU(100)
		node.AddEntityCount(1000)
		node.AddZoneCount(10)

		// 验证所有值
		assert.Equal(t, NodeStateOnline, node.GetState())
		assert.False(t, node.GetLastHeartbeat().IsZero())
		assert.False(t, node.GetUpdatedAt().IsZero())
		assert.Equal(t, int32(100), node.GetCCU())
		assert.Equal(t, int64(1000), node.GetEntityCount())
		assert.Equal(t, int32(10), node.GetZoneCount())

		// 模拟一些操作
		node.SubCCU(20)
		node.DecrEntityCount()
		node.SubZoneCount(2)

		// 再次验证
		assert.Equal(t, int32(80), node.GetCCU())
		assert.Equal(t, int64(999), node.GetEntityCount())
		assert.Equal(t, int32(8), node.GetZoneCount())
	})
}
