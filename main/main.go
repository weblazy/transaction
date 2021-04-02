package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/redis.v5"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cast"
	gormx "github.com/sunmi-OS/gocore/gorm"
	"github.com/weblazy/gocore/utils"
	"github.com/weblazy/transaction/config"
	"github.com/weblazy/transaction/model/order"
	"github.com/weblazy/transaction/model/tx"
	"github.com/weblazy/transaction/model/wallet"
)

func main() {
	Init()
	// 链接redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewClient(&redis.Options{Addr: redisHost, Password: redisPassword, DB: 0})
	go TxCron(redisClient)
	// for i := 0; i < 1000; i++ {
	// 	RunTx(redisClient)
	// }
	select {}

}
func Init() {
	// Initialize the configuration center
	config.InitNacos(utils.GetRunTime())
	// 链接数据库
	gormx.NewDB("DbOrder")
	gormx.NewDB("DbWallet")
	gormx.NewDB("DbTx")
	// 创建表
	tx.CreateTable()
	order.CreateTable()
	wallet.CreateTable()
}

type Id struct {
	Id int64 `json:"id" gorm:"column:id"`
}

type TrxId struct {
	TrxId int64 `json:"trx_id" gorm:"column:trx_id"`
}

func RunTx(redisClient *redis.Client) {
	// 对uid为1的用户进行加锁

	uid := "1"
	ok, err := redisClient.SetNX("tx_"+uid, 1, 0).Result()
	if !ok {
		return
	}
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// 创建事务
	transaction := tx.Tx{Uid: cast.ToInt64(uid)}
	tx.TxHandler.Insert(nil, &transaction)

	// 开启钱包事务
	walletTx := wallet.Orm().Begin()
	defer walletTx.RollbackUnlessCommitted()
	// now := time.Now().Unix()

	// 开启订单事务
	orderTx := order.Orm().Begin()
	defer orderTx.RollbackUnlessCommitted()

	// check
	money := "money - 1"
	err = wallet.WalletHandler.Update(walletTx, map[string]interface{}{
		"money": gorm.Expr(money),
	}, "id = 1")
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	err = order.OrderHandler.Update(orderTx, map[string]interface{}{
		"status": 1,
	}, "id = 1")
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// preper
	contentByte, err := json.Marshal(map[string]interface{}{
		"money": "money - 1",
	})
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	var mysql1 Id
	// 获取当前链接事物Id
	err = walletTx.Raw("SELECT TRX_ID as id FROM INFORMATION_SCHEMA.INNODB_TRX  WHERE TRX_MYSQL_THREAD_ID = CONNECTION_ID();").Scan(&mysql1).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	wTx1 := wallet.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql1.Id,
		Content: string(contentByte),
	}
	err = wallet.TxRecordHandler.Insert(nil, &wTx1)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}

	contentByte, err = json.Marshal(map[string]interface{}{
		"status": 1,
	})
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	var mysql2 Id
	// 获取当前链接事物Id
	err = orderTx.Raw("SELECT TRX_ID as id FROM INFORMATION_SCHEMA.INNODB_TRX  WHERE TRX_MYSQL_THREAD_ID = CONNECTION_ID();").Scan(&mysql2).Error
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}

	wTx2 := order.TxRecord{
		TxId:    transaction.Id,
		MysqlId: mysql2.Id,
		Content: string(contentByte),
	}
	err = order.TxRecordHandler.Insert(nil, &wTx2)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
	}
	tx.TxHandler.Update(nil, map[string]interface{}{
		"status": 1,
	}, "id = ?", transaction.Id)
	// commit

	err = wallet.TxRecordHandler.Update(walletTx, map[string]interface{}{
		"status": 2,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}

	err = order.TxRecordHandler.Update(orderTx, map[string]interface{}{
		"status": 2,
	}, "tx_id = ?", transaction.Id)
	if err != nil {
		fmt.Printf("%#v\n", err.Error())
		return
	}
	// time.Sleep(time.Hour)
	orderTx.Commit()
	walletTx.Commit()
	redisClient.Del("tx_" + uid)
	tx.TxHandler.Update(nil, map[string]interface{}{
		"status": 2,
	}, "id = ?", transaction.Id)

}

func TxCron(redisClient *redis.Client) {
	for {
		unPreperList, _ := tx.TxHandler.GetList("status = 0")
		preperList, _ := tx.TxHandler.GetList("status = 1")
		// 获取当前所有的事物
		var walletTrxId []*TrxId
		err := wallet.Orm().Raw("select trx_id from information_schema.INNODB_TRX;").Scan(&walletTrxId).Error
		if err != nil {
			fmt.Printf("%#v\n", err.Error())
		}
		walletTrxMap := make(map[int64]bool)
		for k1 := range walletTrxId {
			walletTrxMap[walletTrxId[k1].TrxId] = true
		}

		var orderTrxId []*TrxId
		err = order.Orm().Raw("select trx_id from information_schema.INNODB_TRX;").Scan(&orderTrxId).Error
		if err != nil {
			fmt.Printf("%#v\n", err.Error())
		}
		orderTrxMap := make(map[int64]bool)
		for k1 := range orderTrxId {
			orderTrxMap[orderTrxId[k1].TrxId] = true
		}
		for k1 := range unPreperList {
			obj := unPreperList[k1]
			walletObj, _ := wallet.TxRecordHandler.GetOne("tx_id = ?", obj.Id)
			orderObj, _ := wallet.TxRecordHandler.GetOne("tx_id = ?", obj.Id)
			_, ok1 := walletTrxMap[walletObj.MysqlId]
			_, ok2 := orderTrxMap[orderObj.MysqlId]
			if (walletObj.MysqlId == 0 || !ok1) && (orderObj.MysqlId == 0 || !ok2) {
				row, _ := tx.TxHandler.Update(nil, map[string]interface{}{
					"status": 3,
				}, "id = ? and status = 0", obj.Id)
				if row > 0 {
					redisClient.Del("tx_" + cast.ToString(obj.Uid))
				}
			} else {

			}
			fmt.Printf("%#v\n", obj)
		}

		for k1 := range preperList {
			obj := preperList[k1]
			walletObj, _ := wallet.TxRecordHandler.GetOne("tx_id = ?", obj.Id)
			orderObj, _ := wallet.TxRecordHandler.GetOne("tx_id = ?", obj.Id)
			_, ok1 := walletTrxMap[walletObj.MysqlId]
			_, ok2 := orderTrxMap[orderObj.MysqlId]
			if !ok1 && !ok2 {
				if walletObj.Status == 1 {
					// 开启钱包事务
					walletTx := wallet.Orm().Begin()
					defer walletTx.RollbackUnlessCommitted()

					money := "money - 1"
					err = wallet.WalletHandler.Update(walletTx, map[string]interface{}{
						"money": gorm.Expr(money),
					}, "id = 1")
					if err != nil {
						fmt.Printf("%#v\n", err.Error())
						continue
					}
					err = wallet.TxRecordHandler.Update(walletTx, map[string]interface{}{
						"status": 2,
					}, "tx_id = ? and status = 1", obj.Id)
					if err != nil {
						fmt.Printf("%#v\n", err.Error())
						continue
					}
					walletTx.Commit()
				}
				if orderObj.Status == 1 {
					// 开启订单事务
					orderTx := order.Orm().Begin()
					defer orderTx.RollbackUnlessCommitted()
					err = order.OrderHandler.Update(orderTx, map[string]interface{}{
						"status": 1,
					}, "id = 1")
					if err != nil {
						fmt.Printf("%#v\n", err.Error())
						continue
					}
					err = order.TxRecordHandler.Update(orderTx, map[string]interface{}{
						"status": 2,
					}, "tx_id = ?  and status = 1", obj.Id)
					if err != nil {
						fmt.Printf("%#v\n", err.Error())
						continue
					}
					orderTx.Commit()
				}
				row, _ := tx.TxHandler.Update(nil, map[string]interface{}{
					"status": 2,
				}, "id = ? and status = 1", obj.Id)
				if row > 0 {
					redisClient.Del("tx_" + cast.ToString(obj.Uid))
				}
			}
			fmt.Printf("%#v\n", obj)
		}
		time.Sleep(2 * time.Second)
	}
}
