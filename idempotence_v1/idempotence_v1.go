package idempotence_v1

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"time"
)

type Idempotence interface {
	// GenerateId 生成或拿取唯一识别ID
	GenerateId() string

	IdempotenceStorage
}

type IdempotenceStorage interface {
	// SaveIfAbsent 保存唯一ID
	SaveIfAbsent(idempotenceId string) bool

	// Delete 删除幂等ID
	Delete(idempotenceId string) (result bool)
}

type RedisIdempotenceImpl struct {
	conn *redis.Client
}

// NewRedisIdempotence 复用Redis链接，不在内部创建
func NewRedisIdempotence(conn *redis.Client) IdempotenceStorage {
	return &RedisIdempotenceImpl{conn: conn}
}

func (idem *RedisIdempotenceImpl) GenerateId() string {
	return uuid.New().String()
}

// SaveIfAbsent 根据返回值判断幂等ID是否有存在
func (idem *RedisIdempotenceImpl) SaveIfAbsent(idempotenceId string) bool {
	// setnx : 不加ExpireTime是原子的，加ExpireTime原子操作要用LUA脚本才是原子操作
	// 相当于执行了setnx后，再设置过期时间.
	// 故Redis在设置key后崩溃，ExpireTime是加不上的，
	// 不过后面的DEL操作肯定也报错了，加上打印，以及错误上报（Prometheus），让人工进行干预
	return idem.conn.SetNX(idempotenceId, 1, time.Second*60).Val()
}

// Delete 失败的情况，应将失败SQL和幂等ID打印出来，以预期人工干预来进行补偿
func (idem *RedisIdempotenceImpl) Delete(idempotenceId string) bool {
	err := idem.conn.Del(idempotenceId).Err()
	if err != nil {
		return false
	}
	return true
}
