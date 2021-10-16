package main

import "github.com/google/uuid"

// 利用Mysql事务，将幂等ID与业务结合到一个事务中，当业务失败时，写幂等ID也会被回滚
// 规避掉业务操作失败，但幂等号却删不掉的情况

// 幂等ID存储到数据库后，再回填到Redis中

// 如果不是同一数据库，或者说不是同一类型组件，则需要使用分布式事务，实现同步或异步的统一，过于复杂，慎用

// GenerateId 幂等ID ：来源Resource + ID
func GenerateId() string {
	return uuid.New().String()
}



