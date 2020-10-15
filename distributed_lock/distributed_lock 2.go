package main

//const (
//	success int64  = 1
//	OK      string = "OK"
//)
//
//var redisConn redis.Conn
//
//func init() {
//	initRedis()
//}
//
//func initRedis() {
//	redisClient := &redis.Pool{
//		MaxIdle:     10, // 最大空闲连接
//		MaxActive:   10, // 最大激活连接
//		IdleTimeout: 5 * time.Second,
//		Dial: func() (redis.Conn, error) {
//			rc, err := redis.Dial("tcp", "127.0.0.1:6379")
//			if err != nil {
//				return nil, err
//			}
//			rc.Do("SELECT", 0) // 切换到指定数据库
//			fmt.Println("USE DB", 0)
//			return rc, nil
//		},
//	}
//	redisConn = redisClient.Get()
//}
//
//func main() {
//	//fmt.Println(tryReleaseLockWithLua("guowei", "brother"))
//	pfTest()
//}
//
//func tryLockWithLua(key, UniqueId string, seconds int) bool {
//
//	luaScript := redis.NewScript(1, "if redis.call('setnx',KEYS[1],ARGV[1]) == 1 "+
//		"then "+
//		"redis.call('expire',KEYS[1],ARGV[2]) return 1 else return 0 end")
//
//	result, err := luaScript.Do(redisConn, key, UniqueId, seconds)
//	return result == success && err == nil
//}
//
//func tryReleaseLockWithLua(key, value string) bool {
//	luaScript := redis.NewScript(1, "if redis.call('get',KEYS[1]) == ARGV[1]"+
//		" then "+
//		"return redis.call('del',KEYS[1]) else return 0 end")
//	result, err := luaScript.Do(redisConn, key, value)
//	return result == success && err == nil
//}
//
//func tryLock(key, request string, timeout int) bool {
//	result, err := redisConn.Do("SETNX", key, request)
//	if err == nil && result == success {
//		result, err := redisConn.Do("EXPIRE", key, timeout)
//		return err == nil && result == success
//	} else {
//		return false
//	}
//}
//
//func tryLockWithSet(key, UniqueId string, seconds int) bool {
//	result, err := redisConn.Do("SET", key, UniqueId, "NX", "EX", seconds)
//	return result == OK && err == nil
//
//}
//
//func pfTest() {
//	for i := 0; i < 100000; i++ {
//		redisConn.Do("PFADD", "codehole", "user"+strconv.Itoa(i))
//		//if total.(int64) != int64(i + 1) {
//		//	fmt.Println("total:", total)
//		//	fmt.Println("i + 1:", i + 1)
//		//	break
//		//}
//	}
//	total, _ := redisConn.Do("PFCOUNT", "codehole")
//	fmt.Println("total:", total)
//	fmt.Println("实际：:", 100000)
//
//}
