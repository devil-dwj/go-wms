package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	red "github.com/go-redis/redis/v8"
)

var (
	luaExtend    = red.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`)
	deleteScript = red.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)
	luaPTTL      = red.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pttl", KEYS[1]) else return -3 end`)
)

var (
	// 加锁失败
	ErrLockFailed = errors.New("lock failed")
	// 未获得锁
	ErrLockObtained = errors.New("not obtained")
	// 未锁
	ErrLockNotHeld = errors.New("lock not held")
	// 自旋超时
	ErrSpinLockTimeOut = errors.New("spin lock time out")
)

type rediser interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *red.BoolCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *red.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *red.Cmd
	ScriptExists(ctx context.Context, scripts ...string) *red.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *red.StringCmd
}

type Mutex interface {
	Lock() error
	SpinLock() error
	Unlock() (bool, error)
	TTL() (time.Duration, error)
	Extend() error
}

type Sync interface {
	NewMutex(name string) Mutex
}

type SyncRedis struct {
	*red.Client
}

func NewSync(r *red.Client) Sync {
	return &SyncRedis{Client: r}
}

type MutexRedis struct {
	redis  rediser
	key    string
	value  string
	expiry time.Duration
}

func (r *SyncRedis) NewMutex(key string) Mutex {
	return &MutexRedis{
		redis:  r.Client,
		key:    key,
		value:  genValue(),
		expiry: time.Second * 10,
	}
}

func (m *MutexRedis) Lock() error {
	ok, err := m.obtain(m.key, m.value, m.expiry)
	if err != nil {
		return nil
	} else if ok {
		return nil
	}

	return nil
}

func (m *MutexRedis) SpinLock() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	var timer *time.Timer
	var retry time.Duration = time.Duration(100) * time.Millisecond
	for {
		ok, err := m.obtain(m.key, m.value, m.expiry)
		if err != nil {
			return nil
		} else if ok {
			return nil
		}

		if timer == nil {
			timer = time.NewTimer(retry)
			defer timer.Stop()
		} else {
			timer.Reset(retry)
		}

		select {
		case <-ctx.Done():
			return ErrSpinLockTimeOut
		case <-timer.C:
		}
	}
}

func (m *MutexRedis) Unlock() (bool, error) {
	res, err := deleteScript.Run(context.Background(), m.redis, []string{m.key}, m.value).Result()
	if err == red.Nil {
		return false, ErrLockNotHeld
	}
	if err != nil {
		return false, err
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return false, ErrLockNotHeld
	}

	return true, nil
}

func (m *MutexRedis) TTL() (time.Duration, error) {
	res, err := luaPTTL.Run(context.Background(), m.redis, []string{m.key}, m.value).Result()
	if err == red.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	if num := res.(int64); num > 0 {
		return time.Duration(num) * time.Millisecond, nil
	}
	return 0, nil
}

func (m *MutexRedis) Extend() error {
	ttlVal := strconv.FormatInt(int64(m.expiry/time.Millisecond), 10)
	status, err := luaExtend.Run(context.Background(), m.redis, []string{m.key}, m.value, ttlVal).Result()
	if err != nil {
		return err
	} else if status == int64(1) {
		return nil
	}
	return ErrLockObtained
}

func (m *MutexRedis) obtain(key string, value string, ttl time.Duration) (bool, error) {
	return m.redis.SetNX(context.Background(), key, value, ttl).Result()
}

func genValue() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "default-value"
	}
	return base64.StdEncoding.EncodeToString(b)
}
