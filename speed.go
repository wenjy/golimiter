package limiter

import (
	"sync/atomic"
	"time"
)

// 计算速度逻辑：每次发送完数据，递增当前时间戳的nowUpBytes，下一秒到来时把nowUpBytes赋值给upSpeed
// upSpeed就是上一秒的上行速度，nowUpBytes置零后继续递增，下行逻辑一致
type Speed struct {
	upSpeed      uint32 // 上行速度 bytes/s
	downSpeed    uint32 // 下行速度 bytes/s
	nowUpAt      int64  // 当前记录上行的时间戳，需要根据这个时间戳来判断nowUpBytes是否是当前的时间发送量
	nowDownAt    int64  // 当前记录下行的时间戳
	nowUpBytes   uint32 // 当前时间戳上行bytes
	nowDownBytes uint32 // 当前时间戳下行bytes
}

func NewSpeed() *Speed {
	return &Speed{}
}

// 递增上行Bytes，并把统计完成的Bytes，写入到上行速度字段
func (s *Speed) IncrUpBytes(n int) {
	now := time.Now().Unix()
	// 首先交换新旧时间戳，把最新的时间戳写入
	nowAt := atomic.SwapInt64(&s.nowUpAt, now)

	if nowAt == now { // 还是当前的时间，直接递增
		atomic.AddUint32(&s.nowUpBytes, uint32(n))
	} else if (now - nowAt) == 1 { // 已经到了下一秒，把nowUpBytes保存到upSpeed作为上行速度，nowUpBytes置零递增
		atomic.StoreUint32(&s.upSpeed, s.nowUpBytes)
		atomic.SwapUint32(&s.nowUpBytes, uint32(n))
	} else { // nowUpBytes已经是过时的或者是第一次，表示前1s没有数据发送，upSpeed置零，nowUpBytes递增
		atomic.StoreUint32(&s.upSpeed, 0)
		atomic.StoreUint32(&s.nowUpBytes, uint32(n))
	}
}

// 递增下行Bytes，并把统计完成的Bytes，写入到下行速度字段
func (s *Speed) IncrDownBytes(n int) {
	now := time.Now().Unix()
	nowAt := atomic.SwapInt64(&s.nowDownAt, now)
	if nowAt == now {
		atomic.AddUint32(&s.nowDownBytes, uint32(n))
	} else if (now - nowAt) == 1 {
		atomic.StoreUint32(&s.downSpeed, s.nowDownBytes)
		atomic.SwapUint32(&s.nowDownBytes, uint32(n))
	} else {
		atomic.StoreUint32(&s.downSpeed, 0)
		atomic.StoreUint32(&s.nowDownBytes, uint32(n))
	}
}

// 获取上下行速度
func (s *Speed) UpDownSpeed() (upSpeed, downSpeed uint32) {
	now := time.Now().Unix()
	nowUpAt := atomic.LoadInt64(&s.nowUpAt)
	nowDownAt := atomic.LoadInt64(&s.nowDownAt)
	if nowUpAt == now {
		upSpeed = atomic.LoadUint32(&s.upSpeed)
	} else if (now - nowUpAt) == 1 {
		upSpeed = atomic.LoadUint32(&s.nowUpBytes)
	}

	if nowDownAt == now {
		downSpeed = atomic.LoadUint32(&s.downSpeed)
	} else if (now - nowDownAt) == 1 {
		downSpeed = atomic.LoadUint32(&s.nowDownBytes)
	}

	return
}
