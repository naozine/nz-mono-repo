package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
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

type Pattern struct {
	Name        string
	Description string
	Check       func(string) bool
	Examples    []string
}

func main() {
	db, err := sql.Open("duckdb", "data/app.duckdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 全データを取得
	query := `
		SELECT "都道府県", "市区町村"
		FROM excel_import
		WHERE "都道府県" IS NOT NULL AND "市区町村" IS NOT NULL
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// データを収集
	type Record struct {
		Pref string
		City string
	}
	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Pref, &r.City); err != nil {
			log.Fatal(err)
		}
		records = append(records, r)
	}

	fmt.Printf("総件数: %d\n\n", len(records))

	// パターン定義
	patterns := []Pattern{
		{
			Name:        "前後に空白",
			Description: "前後に全角または半角スペースがある",
			Check: func(s string) bool {
				return s != strings.TrimSpace(s)
			},
		},
		{
			Name:        "23区名のみ（完全一致）",
			Description: "東京23区の区名と完全一致（前後空白除去後）",
			Check: func(s string) bool {
				trimmed := strings.TrimSpace(s)
				for _, ward := range tokyo23Wards {
					if trimmed == ward {
						return true
					}
				}
				return false
			},
		},
		{
			Name:        "23区名+町名等（前方一致）",
			Description: "23区名で始まるが完全一致ではない",
			Check: func(s string) bool {
				trimmed := strings.TrimSpace(s)
				for _, ward := range tokyo23Wards {
					if strings.HasPrefix(trimmed, ward) && trimmed != ward {
						return true
					}
				}
				return false
			},
		},
		{
			Name:        "番地を含む",
			Description: "数字とハイフン・丁目・番地等を含む",
			Check: func(s string) bool {
				// 数字とハイフンまたは「丁目」「番地」「番」を含む
				hasNumber := regexp.MustCompile(`[0-9０-９]`).MatchString(s)
				hasAddress := regexp.MustCompile(`[-－ー丁目番地番]`).MatchString(s)
				return hasNumber && hasAddress
			},
		},
		{
			Name:        "ひらがなのみ",
			Description: "すべてひらがな（空白除く）",
			Check: func(s string) bool {
				trimmed := strings.TrimSpace(s)
				if trimmed == "" {
					return false
				}
				for _, r := range trimmed {
					if !((r >= 'ぁ' && r <= 'ん') || r == ' ' || r == '　') {
						return false
					}
				}
				return true
			},
		},
		{
			Name:        "カタカナのみ",
			Description: "すべてカタカナ（空白除く）",
			Check: func(s string) bool {
				trimmed := strings.TrimSpace(s)
				if trimmed == "" {
					return false
				}
				for _, r := range trimmed {
					if !((r >= 'ァ' && r <= 'ヴ') || r == ' ' || r == '　' || r == 'ー') {
						return false
					}
				}
				return true
			},
		},
		{
			Name:        "異常に短い（3文字以下）",
			Description: "空白除去後3文字以下",
			Check: func(s string) bool {
				return len([]rune(strings.TrimSpace(s))) <= 3
			},
		},
	}

	// 東京都のデータで分析
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("東京都のデータパターン分析")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	tokyoRecords := []string{}
	for _, r := range records {
		if r.Pref == "東京都" {
			tokyoRecords = append(tokyoRecords, r.City)
		}
	}

	fmt.Printf("東京都の件数: %d\n\n", len(tokyoRecords))

	for _, pattern := range patterns {
		count := 0
		var examples []string
		for _, city := range tokyoRecords {
			if pattern.Check(city) {
				count++
				if len(examples) < 5 {
					examples = append(examples, city)
				}
			}
		}

		fmt.Printf("【%s】\n", pattern.Name)
		fmt.Printf("  説明: %s\n", pattern.Description)
		fmt.Printf("  件数: %d件 (%.1f%%)\n", count, float64(count)*100/float64(len(tokyoRecords)))
		if len(examples) > 0 {
			fmt.Printf("  例:\n")
			for _, ex := range examples {
				fmt.Printf("    %q\n", ex)
			}
		}
		fmt.Println()
	}

	// 23区パターンの詳細分析
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("東京23区の詳細パターン")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	wardPatterns := make(map[string]map[string]int) // ward -> pattern -> count
	for _, ward := range tokyo23Wards {
		wardPatterns[ward] = map[string]int{
			"完全一致":    0,
			"町名のみ追加":  0,
			"番地あり":    0,
			"前後空白あり":  0,
			"その他前方一致": 0,
		}
	}

	for _, city := range tokyoRecords {
		trimmed := strings.TrimSpace(city)
		hasSpace := city != trimmed

		for _, ward := range tokyo23Wards {
			if strings.HasPrefix(trimmed, ward) {
				if trimmed == ward {
					wardPatterns[ward]["完全一致"]++
					if hasSpace {
						wardPatterns[ward]["前後空白あり"]++
					}
				} else {
					// 番地を含むか
					hasAddress := regexp.MustCompile(`[0-9０-９][-－ー丁目番地番]`).MatchString(trimmed)
					if hasAddress {
						wardPatterns[ward]["番地あり"]++
					} else {
						// 町名のみか
						remaining := strings.TrimPrefix(trimmed, ward)
						if len([]rune(remaining)) <= 10 { // 町名は通常短い
							wardPatterns[ward]["町名のみ追加"]++
						} else {
							wardPatterns[ward]["その他前方一致"]++
						}
					}
					if hasSpace {
						wardPatterns[ward]["前後空白あり"]++
					}
				}
				break
			}
		}
	}

	// 件数が多い区を表示
	fmt.Println("件数が多い区（上位10件）:\n")
	type wardCount struct {
		Ward  string
		Total int
	}
	var wardCounts []wardCount
	for ward, patterns := range wardPatterns {
		total := 0
		for _, count := range patterns {
			total += count
		}
		wardCounts = append(wardCounts, wardCount{ward, total})
	}
	// ソート（簡易版）
	for i := 0; i < len(wardCounts); i++ {
		for j := i + 1; j < len(wardCounts); j++ {
			if wardCounts[j].Total > wardCounts[i].Total {
				wardCounts[i], wardCounts[j] = wardCounts[j], wardCounts[i]
			}
		}
	}

	for i := 0; i < 10 && i < len(wardCounts); i++ {
		ward := wardCounts[i].Ward
		total := wardCounts[i].Total
		if total == 0 {
			continue
		}
		fmt.Printf("%s: %d件\n", ward, total)
		for pattern, count := range wardPatterns[ward] {
			if count > 0 {
				fmt.Printf("  - %s: %d件\n", pattern, count)
			}
		}
	}

	// その他のパターン（23区にも他のパターンにも当てはまらないもの）
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("その他のパターン（東京都で23区以外）")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	otherCities := make(map[string]int)
	for _, city := range tokyoRecords {
		trimmed := strings.TrimSpace(city)
		is23Ward := false
		for _, ward := range tokyo23Wards {
			if strings.HasPrefix(trimmed, ward) {
				is23Ward = true
				break
			}
		}
		if !is23Ward {
			// 市のパターンを抽出
			cityName := extractCityName(trimmed)
			otherCities[cityName]++
		}
	}

	// 頻度順にソート
	type cityCount struct {
		City  string
		Count int
	}
	var otherCityCounts []cityCount
	for city, count := range otherCities {
		otherCityCounts = append(otherCityCounts, cityCount{city, count})
	}
	for i := 0; i < len(otherCityCounts); i++ {
		for j := i + 1; j < len(otherCityCounts); j++ {
			if otherCityCounts[j].Count > otherCityCounts[i].Count {
				otherCityCounts[i], otherCityCounts[j] = otherCityCounts[j], otherCityCounts[i]
			}
		}
	}

	fmt.Println("三多摩・島しょ地域の市町村（上位20件）:\n")
	for i := 0; i < 20 && i < len(otherCityCounts); i++ {
		fmt.Printf("%2d. %s: %d件\n", i+1, otherCityCounts[i].City, otherCityCounts[i].Count)
	}
}

// extractCityName は市区町村名から基本的な市名を抽出
func extractCityName(s string) string {
	// 「市」「町」「村」で切る
	for _, suffix := range []string{"市", "町", "村"} {
		if idx := strings.Index(s, suffix); idx >= 0 {
			return s[:idx+len(suffix)]
		}
	}
	// 見つからない場合は最初の5文字（町名が続く場合を想定）
	runes := []rune(s)
	if len(runes) > 5 {
		return string(runes[:5])
	}
	return s
}
