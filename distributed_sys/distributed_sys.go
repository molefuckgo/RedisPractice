package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	go_redis "github.com/go-redis/redis/v8"
	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	success int64  = 1
	OK      string = "OK"
)

func main() {
	distributedV3()
}

func distributedV1() {
	router := gin.Default()
	router.GET("/reduce", func(c *gin.Context) {

		courseId := "1"
		//课程剩余数量
		remainingNumOfCourses, err := rdb.Get(ctx, "remaining_"+courseId).Int()
		var message string
		var code int

		if err != nil {
			log.Println("err is:", err)
			message = "系统发生错误，请稍后重试！"
			code = 500
		} else {
			if remainingNumOfCourses > 0 {
				afterRemainingNumOfCourses := remainingNumOfCourses - 1
				message = fmt.Sprintf("课程购买成功，库存数量%d! ", afterRemainingNumOfCourses)
				code = 200
				err := rdb.Set(ctx, "remaining_"+courseId, afterRemainingNumOfCourses, 0).Err()

				if err != nil {
					message = "系统发生错误，请稍后重试!"
					code = 500
				}
			} else {
				message = "当前课程已经售完！"
				code = 200
			}
			c.JSON(code, gin.H{
				"message": message,
			})
		}
		log.Println(message)
	})
	router.Run()
}

var lock sync.Mutex

func distributedV2() {

	router := gin.Default()
	router.GET("/reduce", func(c *gin.Context) {

		courseId := "1"
		//课程剩余数量
		lock.Lock()
		remainingNumOfCourses, err := rdb.Get(ctx, "remaining_"+courseId).Int()
		var message string
		var code int

		if err != nil {
			message = "系统发生错误，请稍后重试！"
			code = 500
		} else {
			if remainingNumOfCourses > 0 {
				afterRemainingNumOfCourses := remainingNumOfCourses - 1
				message = fmt.Sprintf("课程购买成功，库存数量%d! ", afterRemainingNumOfCourses)
				err := rdb.Set(ctx, "remaining_"+courseId, afterRemainingNumOfCourses, 0).Err()

				if err != nil {
					message = "系统发生错误，请稍后重试!"
					code = 500

				}

			} else {
				message = "当前课程已经售完！"
				code = 200
			}

			c.JSON(code, gin.H{
				"message": message,
			})
		}
		log.Println(message)
		lock.Unlock()

	})
	router.Run(":8081")
}

func distributedV3() {

	router := gin.Default()
	router.GET("/reduce", func(c *gin.Context) {
		var message string
		var code int
		requestUUID := uuid.NewV4().String()
		courseId := "1"
		var lockResult *go_redis.BoolCmd
		lockResult = rdb.SetNX(ctx, fmt.Sprintf("remaining_%s_lock", courseId), requestUUID, 10*time.Second)
		defer func() {
			nowLockId := rdb.Get(ctx, fmt.Sprintf("remaining_%s_lock", courseId)).Val()
			if requestUUID == nowLockId {
				rdb.Del(ctx, fmt.Sprintf("remaining_%s_lock", courseId))
			}
		}()
		for !lockResult.Val() {
			time.Sleep(10 * time.Millisecond)
			rdb.Do(ctx, "GET")
			lockResult = rdb.SetNX(ctx, fmt.Sprintf("remaining_%s_lock", courseId), requestUUID, 10*time.Second)
		}
		remainingNumOfCourses, err := rdb.Get(ctx, "remaining_"+courseId).Int()

		if err != nil {
			message = "系统发生错误，请稍后重试！"
			code = 500
		} else {
			if remainingNumOfCourses > 0 {
				afterRemainingNumOfCourses := remainingNumOfCourses - 1
				message = fmt.Sprintf("课程购买成功，库存数量%d!", afterRemainingNumOfCourses)
				//time.Sleep(100 * time.Millisecond)
				//redisConn.Do("SET", "remaining_"+courseId, remainingNumOfCourses-1)
				err := rdb.Set(ctx, "remaining_"+courseId, afterRemainingNumOfCourses, 0).Err()

				if err != nil {
					message = "系统发生错误，请稍后重试!"
					code = 500
				}
			} else {
				message = "当前课程已经售完！"
				code = 200
			}

			c.JSON(code, gin.H{
				"message": message,
			})
		}
		log.Println(message)
	})
	router.Run(":8081")
}

func ReduceNumOfCourses(w http.ResponseWriter, r *http.Request) {
	go func() {
		fmt.Println("进入goroutine")
		r.ParseForm()
		//if len(r.Form) > 0 {
		//	for k, v := range r.Form {
		//		fmt.Printf("%s=%s", k, v[0])
		//	}
		//}
		courseId := r.FormValue("course_id")
		//课程剩余数量
		remainingNumOfCourses, err := redis.Int(redisConn.Do("GET", "remaining_"+courseId))
		if err != nil {
			w.Write([]byte("系统发生错误，请稍后重试！"))
			log.Println("系统发生错误，请稍后重试！")
			return
		}
		if remainingNumOfCourses > 0 {
			w.Write([]byte(fmt.Sprintf("课程购买成功，库存数量%d! ", remainingNumOfCourses-1)))
			//time.Sleep(100 * time.Millisecond)
			redisConn.Do("SET", "remaining_"+courseId, remainingNumOfCourses-1)
			log.Println(fmt.Sprintf("课程购买成功，库存数量%d! ", remainingNumOfCourses-1))
		} else {
			w.Write([]byte("当前课程已经售完！"))
		}
	}()

}

//---------------------------------------------------------------------------------------------------------------------------------------------

var redisConn redis.Conn
var ctx = context.Background()
var rdb *go_redis.Client

func init() {
	RedisNewClient()
}

func RedisNewClient() {
	rdb = go_redis.NewClient(&go_redis.Options{
		Addr:         "localhost:6379",
		Password:     "", // no password set
		DB:           0,  // use default DB
		DialTimeout:  3 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolSize:     10,
		PoolTimeout:  2 * time.Second,
	})

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
}

func initRedis() {
	redisClient := &redis.Pool{
		MaxIdle:     500, // 最大空闲连接
		MaxActive:   500, // 最大激活连接
		IdleTimeout: 5 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			rc, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			rc.Do("SELECT", 0) // 切换到指定数据库
			fmt.Println("USE DB", 0)
			return rc, nil
		},
	}
	redisConn = redisClient.Get()
}

func tryLockWithLua(key, UniqueId string, seconds int) bool {

	luaScript := redis.NewScript(1, "if redis.call('setnx',KEYS[1],ARGV[1]) == 1 "+
		"then "+
		"redis.call('expire',KEYS[1],ARGV[2]) return 1 else return 0 end")

	result, err := luaScript.Do(redisConn, key, UniqueId, seconds)
	return result == success && err == nil
}

func tryReleaseLockWithLua(key, value string) bool {
	luaScript := redis.NewScript(1, "if redis.call('get',KEYS[1]) == ARGV[1]"+
		" then "+
		"return redis.call('del',KEYS[1]) else return 0 end")
	result, err := luaScript.Do(redisConn, key, value)
	return result == success && err == nil
}

//func tryLock(key, request string, timeout int) bool {
//	result, err := redisConn.Do("SETNX", key, request)
//	if err == nil && result == success {
//		result, err := redisConn.Do("EXPIRE", key, timeout)
//		return err == nil && result == success
//	} else {
//		return false
//	}
//}

//func tyrLock(key, request string, timeout int ) bool {
//
//}

func tryLockWithSet(key, UniqueId string, seconds int) bool {
	result, err := redisConn.Do("SET", key, UniqueId, "NX", "EX", seconds)
	return result == OK && err == nil
}

func pfTest() {
	for i := 0; i < 100000; i++ {
		redisConn.Do("PFADD", "codehole", "user"+strconv.Itoa(i))
		//if total.(int64) != int64(i + 1) {
		//	fmt.Println("total:", total)
		//	fmt.Println("i + 1:", i + 1)
		//	break
		//}

	}
	total, _ := redisConn.Do("PFCOUNT", "codehole")
	fmt.Println("total:", total)
	fmt.Println("实际：:", 100000)

}
