package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite3", "data/xiaozhi.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// 查询所有纯数字类型的配置记录（没有引号的数字）
	rows, err := db.Query("SELECT id, key, value FROM config_records WHERE value NOT LIKE '\"%\"'")
	if err != nil {
		log.Fatal("Failed to query config records:", err)
	}
	defer rows.Close()

	type Record struct {
		id    int
		key   string
		value string
	}

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.id, &r.key, &r.value); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}

	fmt.Printf("Found %d numeric/type records to fix\n", len(records))

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	// 更新每条记录
	for _, record := range records {
		var jsonValue []byte
		var err error

		// 处理特殊值
		if record.value == "true" || record.value == "false" {
			jsonValue, err = json.Marshal(record.value == "true")
			if err != nil {
				log.Printf("Failed to marshal bool value for key %s: %v", record.key, err)
				continue
			}
		} else if record.value == "null" {
			jsonValue = []byte("null")
		} else {
			// 尝试解析为整数
			if intVal, err := strconv.ParseInt(record.value, 10, 64); err == nil {
				jsonValue, err = json.Marshal(intVal)
				if err != nil {
					log.Printf("Failed to marshal int value for key %s: %v", record.key, err)
					continue
				}
			} else if floatVal, err := strconv.ParseFloat(record.value, 64); err == nil {
				jsonValue, err = json.Marshal(floatVal)
				if err != nil {
					log.Printf("Failed to marshal float value for key %s: %v", record.key, err)
					continue
				}
			} else {
				// 如果都失败了，作为字符串处理
				jsonValue, err = json.Marshal(record.value)
				if err != nil {
					log.Printf("Failed to marshal string value for key %s: %v", record.key, err)
					continue
				}
			}
		}

		// 更新数据库
		_, err = tx.Exec("UPDATE config_records SET value = ? WHERE id = ?", string(jsonValue), record.id)
		if err != nil {
			log.Printf("Failed to update record %s: %v", record.key, err)
			continue
		}

		fmt.Printf("Fixed: %s = %s -> %s\n", record.key, record.value, string(jsonValue))
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("Numeric values fix completed successfully!")
}