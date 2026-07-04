// cmd/migrate_users/main.go
// 用法: go run ./cmd/migrate_users
// 默认源库：catering_recruitment；默认目标库：项目 config/local.yml 中的 cyxx_test
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randAlphanumeric(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func genUserCode(dstDB *sql.DB) string {
	for i := 0; i < 10; i++ {
		code := randAlphanumeric(8)
		var count int
		dstDB.QueryRow("SELECT COUNT(*) FROM user WHERE user_code=?", code).Scan(&count)
		if count == 0 {
			return code
		}
	}
	log.Fatal("failed to generate unique user_code")
	return ""
}

type OldUser struct {
	Avatar        string
	Sex           int
	Age           int
	Birthday      sql.NullString
	Phone         string
	WechartCode   sql.NullString
	WechatOpenID  string
	FirstAreaID   int
	SecondAreaID  int
	ThirdAreaID   int
	Address       string
	Longitude     float64
	Latitude      float64
	Integral      uint64
	CollectNum    uint64
	BuyNum        uint64
	InviteNum     uint64
	FirstRecharge sql.NullString
	TotalRecharge float64
	DeviceModel   string
	IP            string
	CreateAt      time.Time
}

func main() {
	src := flag.String("src", "producter:20220123Zzy!@tcp(114.115.153.27:3306)/catering_recruitment?charset=utf8mb4&parseTime=True&loc=Local", "source DSN (old DB)")
	dst := flag.String("dst", "cyxx_test:cyxx_test123!@#@tcp(rm-2zey437xt5595ehf67o.mysql.rds.aliyuncs.com:3306)/cyxx_test?charset=utf8mb4&parseTime=True&loc=Local", "destination DSN (new DB)")
	batchSize := flag.Int("batch", 500, "rows per batch")
	flag.Parse()

	srcDB, err := sql.Open("mysql", *src)
	if err != nil {
		log.Fatalf("open src: %v", err)
	}
	defer srcDB.Close()

	dstDB, err := sql.Open("mysql", *dst)
	if err != nil {
		log.Fatalf("open dst: %v", err)
	}
	defer dstDB.Close()

	var total int
	srcDB.QueryRow("SELECT COUNT(*) FROM user").Scan(&total)
	log.Printf("source rows: %d", total)

	offset := 0
	inserted := 0

	for {
		rows, err := srcDB.Query(`
			SELECT avatar, sex, age, birthday, phone, wechart_code, wechat_open_id,
			       first_area_id, second_area_id, third_area_id, address,
			       longitude, latitude, integral, collect_num, buy_num,
			       invite_num, first_recharge, total_recharge, device_model, ip,
			       create_at
			FROM user ORDER BY id LIMIT ? OFFSET ?`, *batchSize, offset)
		if err != nil {
			log.Fatalf("query src: %v", err)
		}

		var batch []OldUser
		for rows.Next() {
			var u OldUser
			if err := rows.Scan(
				&u.Avatar, &u.Sex, &u.Age, &u.Birthday, &u.Phone,
				&u.WechartCode, &u.WechatOpenID,
				&u.FirstAreaID, &u.SecondAreaID, &u.ThirdAreaID, &u.Address,
				&u.Longitude, &u.Latitude,
				&u.Integral, &u.CollectNum, &u.BuyNum,
				&u.InviteNum, &u.FirstRecharge, &u.TotalRecharge,
				&u.DeviceModel, &u.IP, &u.CreateAt,
			); err != nil {
				log.Fatalf("scan: %v", err)
			}
			batch = append(batch, u)
		}
		rows.Close()

		if len(batch) == 0 {
			break
		}

		tx, err := dstDB.Begin()
		if err != nil {
			log.Fatalf("begin tx: %v", err)
		}

		stmt, err := tx.Prepare(`
			INSERT IGNORE INTO user
			  (avatar, name, user_code, sex, age, birthday, phone, wechart_code, wechat_open_id,
			   first_area_id, second_area_id, third_area_id, address,
			   longitude, latitude, integral, collect_num, buy_num,
			   invite_num, first_recharge, total_recharge, device_model, ip,
			   contact_voucher_num, first_top_status, new_customer_status, profile_complete_status, old_user_status,
			   create_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
		if err != nil {
			tx.Rollback()
			log.Fatalf("prepare: %v", err)
		}

		for _, u := range batch {
			name := "餐饮人" + randAlphanumeric(6)
			userCode := genUserCode(dstDB)
			_, err := stmt.Exec(
				u.Avatar, name, userCode, u.Sex, u.Age, u.Birthday, u.Phone,
				u.WechartCode, u.WechatOpenID,
				u.FirstAreaID, u.SecondAreaID, u.ThirdAreaID, u.Address,
				u.Longitude, u.Latitude,
				u.Integral, u.CollectNum, u.BuyNum,
				u.InviteNum, u.FirstRecharge, u.TotalRecharge,
				u.DeviceModel, u.IP,
				0, 0, 0, 0, 1, // old_user_status=1
				u.CreateAt,
			)
			if err != nil {
				stmt.Close()
				tx.Rollback()
				log.Fatalf("insert phone=%s: %v", u.Phone, err)
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Fatalf("commit: %v", err)
		}

		inserted += len(batch)
		fmt.Printf("\rmigrated %d / %d", inserted, total)
		offset += *batchSize
	}

	fmt.Printf("\ndone. total inserted: %d\n", inserted)
}
