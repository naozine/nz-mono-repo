package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
)

// 東京23区のリスト
var tokyo23Wards = []string{
	"千代田区", "中央区", "港区", "新宿区", "文京区", "台東区", "墨田区",
	"江東区", "品川区", "目黒区", "大田区", "世田谷区", "渋谷区", "中野区",
	"杉並区", "豊島区", "北区", "荒川区", "板橋区", "練馬区", "足立区",
	"葛飾区", "江戸川区",
}

func main() {
	db, err := sql.Open("duckdb", "data/app.duckdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 東京都の全データを取得
	query := `
		SELECT "市区町村"
		FROM excel_import
		WHERE "都道府県" = '東京都' AND "市区町村" IS NOT NULL
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			log.Fatal(err)
		}
		cities = append(cities, city)
	}

	total := len(cities)
	fmt.Printf("東京都の総件数: %d\n\n", total)

	// 前方一致でマッチするか判定
	matched := 0
	unmatched := 0
	var unmatchedExamples []string

	for _, city := range cities {
		trimmed := strings.TrimSpace(city)
		isMatched := false

		for _, ward := range tokyo23Wards {
			if strings.HasPrefix(trimmed, ward) {
				isMatched = true
				break
			}
		}

		if isMatched {
			matched++
		} else {
			unmatched++
			if len(unmatchedExamples) < 100 {
				unmatchedExamples = append(unmatchedExamples, city)
			}
		}
	}

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("前方一致（LIKE '区名%%'）のカバー率\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Printf("23区前方一致でマッチ: %d件 (%.1f%%)\n", matched, float64(matched)*100/float64(total))
	fmt.Printf("マッチしない: %d件 (%.1f%%)\n\n", unmatched, float64(unmatched)*100/float64(total))

	// マッチしないデータの内訳を分析
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("残り%.1f%%の内訳（マッチしないデータ）\n", float64(unmatched)*100/float64(total))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// パターン別に分類
	patterns := make(map[string][]string)
	patterns["市で終わる"] = []string{}
	patterns["町・村で終わる"] = []string{}
	patterns["ひらがな・カタカナ"] = []string{}
	patterns["空白のみ・特殊文字"] = []string{}
	patterns["その他"] = []string{}

	for _, city := range unmatchedExamples {
		trimmed := strings.TrimSpace(city)

		if strings.HasSuffix(trimmed, "市") {
			patterns["市で終わる"] = append(patterns["市で終わる"], city)
		} else if strings.HasSuffix(trimmed, "町") || strings.HasSuffix(trimmed, "村") {
			patterns["町・村で終わる"] = append(patterns["町・村で終わる"], city)
		} else if isHiraganaOrKatakana(trimmed) {
			patterns["ひらがな・カタカナ"] = append(patterns["ひらがな・カタカナ"], city)
		} else if len(trimmed) == 0 || trimmed == "\u3000" {
			patterns["空白のみ・特殊文字"] = append(patterns["空白のみ・特殊文字"], city)
		} else {
			patterns["その他"] = append(patterns["その他"], city)
		}
	}

	for patternName, examples := range patterns {
		if len(examples) > 0 {
			fmt.Printf("【%s】: %d件\n", patternName, len(examples))
			displayCount := 10
			if len(examples) < displayCount {
				displayCount = len(examples)
			}
			for i := 0; i < displayCount; i++ {
				fmt.Printf("  %q\n", examples[i])
			}
			if len(examples) > displayCount {
				fmt.Printf("  ... 他%d件\n", len(examples)-displayCount)
			}
			fmt.Println()
		}
	}

	// マッチしないデータの全件をユニーク値でカウント
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("マッチしないデータのユニーク値（頻度順）\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	uniqueCounts := make(map[string]int)
	for _, city := range cities {
		trimmed := strings.TrimSpace(city)
		isMatched := false
		for _, ward := range tokyo23Wards {
			if strings.HasPrefix(trimmed, ward) {
				isMatched = true
				break
			}
		}
		if !isMatched {
			uniqueCounts[city]++
		}
	}

	// 頻度順にソート
	type cityCount struct {
		City  string
		Count int
	}
	var sortedCounts []cityCount
	for city, count := range uniqueCounts {
		sortedCounts = append(sortedCounts, cityCount{city, count})
	}
	for i := 0; i < len(sortedCounts); i++ {
		for j := i + 1; j < len(sortedCounts); j++ {
			if sortedCounts[j].Count > sortedCounts[i].Count {
				sortedCounts[i], sortedCounts[j] = sortedCounts[j], sortedCounts[i]
			}
		}
	}

	fmt.Printf("上位30件:\n\n")
	for i := 0; i < 30 && i < len(sortedCounts); i++ {
		fmt.Printf("%2d. %q: %d件\n", i+1, sortedCounts[i].City, sortedCounts[i].Count)
	}

	// 23区の「誤入力」や「変形」を検出
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("23区の可能性がある誤入力・変形\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, city := range unmatchedExamples {
		trimmed := strings.TrimSpace(city)
		for _, ward := range tokyo23Wards {
			// 部分一致（順序違いなど）
			wardBase := strings.TrimSuffix(ward, "区")
			if strings.Contains(trimmed, wardBase) && trimmed != ward {
				fmt.Printf("%q <- %sの可能性\n", city, ward)
				break
			}
		}
	}
}

func isHiraganaOrKatakana(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'ぁ' && r <= 'ん') || (r >= 'ァ' && r <= 'ヴ') || r == ' ' || r == '　' || r == 'ー') {
			return false
		}
	}
	return true
}
