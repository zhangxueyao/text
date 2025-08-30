package pkg

import (
	"sync"
	"time"
)

// SnowflakeIDGenerator 是 Snowflake 算法的实现
type SnowflakeIDGenerator struct {
	mu            sync.Mutex
	nodeID        int64
	epoch         int64
	lastTimestamp int64
	sequence      int64
}

// SnowflakeIDGenerator 构造函数
func NewSnowflakeIDGenerator(nodeID int64) *SnowflakeIDGenerator {
	// 定义开始时间戳（可以选择不同的起始时间）
	epoch := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6
	return &SnowflakeIDGenerator{
		nodeID:        nodeID, // 唯一节点ID
		epoch:         epoch,  // 起始时间
		sequence:      0,      // 序列号初始为0
		lastTimestamp: -1,     // 上次生成时间初始化为-1
	}
}

// GenerateID 生成一个唯一的ID
func (gen *SnowflakeIDGenerator) GenerateID() (int64, error) {
	gen.mu.Lock()
	defer gen.mu.Unlock()

	// 获取当前时间戳
	timestamp := time.Now().UnixNano() / 1e6 // 当前时间的毫秒数

	// 如果时间戳和上次生成的时间相同，则增加序列号
	if timestamp == gen.lastTimestamp {
		gen.sequence = (gen.sequence + 1) & 0xFFF // 保证序列号最大为 4095
		if gen.sequence == 0 {
			// 如果序列号已经用完，则等待下一毫秒
			for timestamp <= gen.lastTimestamp {
				timestamp = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		gen.sequence = 0 // 不同时间戳则重置序列号
	}

	gen.lastTimestamp = timestamp

	// 计算ID
	id := ((timestamp - gen.epoch) << 22) | (gen.nodeID << 12) | gen.sequence
	return id, nil
}
