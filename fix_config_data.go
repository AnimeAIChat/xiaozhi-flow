package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite3", "data/xiaozhi.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// 查询所有配置记录
	rows, err := db.Query("SELECT id, key, value FROM config_records")
	if err != nil {
		log.Fatal("Failed to query config records:", err)
	}
	defer rows.Close()

	var updates []struct {
		id    int
		key   string
		value string
	}

	for rows.Next() {
		var id int
		var key, value string
		if err := rows.Scan(&id, &key, &value); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		updates = append(updates, struct {
			id    int
			key   string
			value string
		}{id, key, value})
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}

	fmt.Printf("Found %d config records to migrate\n", len(updates))

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	// 更新每条记录
	for _, record := range updates {
		// 将值转换为JSON
		var jsonValue []byte
		var err error

		// 尝试解析为整数
		var intValue int64
		if _, err := fmt.Sscanf(record.value, "%d", &intValue); err == nil {
			jsonValue, err = json.Marshal(intValue)
			if err != nil {
				log.Printf("Failed to marshal int value for key %s: %v", record.key, err)
				continue
			}
		} else if record.value == "true" || record.value == "false" {
			// 布尔值
			boolValue := record.value == "true"
			jsonValue, err = json.Marshal(boolValue)
			if err != nil {
				log.Printf("Failed to marshal bool value for key %s: %v", record.key, err)
				continue
			}
		} else {
			// 默认作为字符串处理
			jsonValue, err = json.Marshal(record.value)
			if err != nil {
				log.Printf("Failed to marshal string value for key %s: %v", record.key, err)
				continue
			}
		}

		// 更新数据库
		_, err = tx.Exec("UPDATE config_records SET value = ? WHERE id = ?", string(jsonValue), record.id)
		if err != nil {
			log.Printf("Failed to update record %s: %v", record.key, err)
			continue
		}

		fmt.Printf("Migrated: %s = %s -> %s\n", record.key, record.value, string(jsonValue))
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("Migration completed successfully!")
}