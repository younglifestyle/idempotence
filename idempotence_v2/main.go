// $ docker run -p 6379:6379 --name my-redis -d redis:latest
// $ docker run -p 3306:3306 --name my-mysql -e MYSQL_ROOT_PASSWORD=123456 -d mysql:latest

package main

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"idempotence_v2/model"
	"sync"
	"time"
)

var cacheSingleFlight = &singleflight.Group{}

func getDbConn() (db *gorm.DB, err error) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true}})

	_ = db.AutoMigrate(&model.StudentMoney{})
	return
}

// 返回返回值给调用者使用，保存redis失败无很大问题，再次请求再传一次即可
func saveIdToRedis(rdb *redis.Client, idempotenceId string) bool {

	return rdb.SetNX(idempotenceId, 1, time.Second*20).Val()
}

//
func getIdFromRedis(rdb *redis.Client, idempotenceId string) string {

	return rdb.Get(idempotenceId).Val()
}

// 学生参与劳动，即可获得奖励，创建转账记录
func giveTipToStudent(db *gorm.DB, rdb *redis.Client, studentMoney *model.StudentMoney) error {

	// 并发量大时，这里是会同时找不到ID，且从DB中也捞不出DB的
	if getIdFromRedis(rdb, studentMoney.IdempotenceId) == "" {
		// 缓存查不到就查数据库
		rr, _, _ := cacheSingleFlight.Do(studentMoney.IdempotenceId, func() (r interface{}, e error) {
			log.Debug("cache miss : ", studentMoney.IdempotenceId)

			var stuMoneyExistInfo model.StudentMoney
			err := db.Select("idempotence_id").First(&stuMoneyExistInfo,
				"idempotence_id = ?", studentMoney.IdempotenceId).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				log.Errorf("select idempotence_id From student_money WHERE idempotence_id = %s, error = %s",
					studentMoney.IdempotenceId, err.Error())
			}

			return stuMoneyExistInfo, nil
		})

		stuMoneyExistInfo := rr.(model.StudentMoney)

		//log.Debug("cache miss : ", studentMoney.IdempotenceId)
		//
		//var stuMoneyExistInfo model.StudentMoney
		//err := db.Select("idempotence_id").First(&stuMoneyExistInfo,
		//	"idempotence_id = ?", studentMoney.IdempotenceId).Error
		//if err != nil && err != gorm.ErrRecordNotFound {
		//	log.Errorf("select idempotence_id From student_money WHERE idempotence_id = %s, error = %s",
		//		studentMoney.IdempotenceId, err.Error())
		//}

		// id exist in db
		if stuMoneyExistInfo.IdempotenceId != "" {
			log.Debug("get id from db : ", stuMoneyExistInfo.IdempotenceId)

			_ = saveIdToRedis(rdb, studentMoney.IdempotenceId)
			return nil
		}
	} else {
		// 缓存中存在幂等值，不执行
		log.Debug("id exist")
		return nil
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		log.Debug("store id : ", studentMoney.IdempotenceId)

		if err := tx.Create(studentMoney).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}

		// 返回 nil 提交事务
		return nil
	}); err != nil {

		return err
	}

	_ = saveIdToRedis(rdb, studentMoney.IdempotenceId)

	return nil
}

func main() {
	log.SetLevel(log.DebugLevel)

	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	dbConn, err := getDbConn()
	if err != nil {
		panic(err)
	}

	//err = dbConn.Transaction(func(tx *gorm.DB) error {
	//	for i := 0; i < 500000; i++ {
	//
	//		tx.Create(&model.StudentMoney{
	//			IdempotenceId: GenerateId(),
	//			Name:          "right",
	//			Age:           i,
	//			Money:         i,
	//		})
	//	}
	//
	//	// 返回 nil 提交事务
	//	return nil
	//})
	//if err != nil {
	//	panic(err)
	//}

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			//time.Sleep(time.Millisecond * 100)
			err = giveTipToStudent(dbConn, rdb, &model.StudentMoney{
				//IdempotenceId: GenerateId(),
				IdempotenceId: "b533516b-744c-41b6-b4d6-ba42bfc13fed",
				Name:          "right",
				Age:           18,
				Money:         100,
			})

			wg.Done()
		}()
	}

	wg.Wait()

	time.Sleep(time.Second * 10)
}
