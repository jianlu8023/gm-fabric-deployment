package main

import (
	mystr "_examples/pkg/str"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
)

func main() {

	path := "./clients.csv"
	csvFile, err := os.Open(path)
	if err != nil {
		log.Printf("open file error %v", err)
		return
	}
	defer func(csvFile *os.File) {
		if err := csvFile.Close(); err != nil {
			log.Printf("close file error %v", err)
		}
	}(csvFile)
	reader := csv.NewReader(csvFile)

	reader.Comment = '#'
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // 每行字段数，-1 表示不检查，其他值表示必须匹配

	var lineCount int64
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("read file error %v", err)
			continue
		}
		lineCount++

		if isBlankRow(record) {
			continue
		}
		log.Printf("line %d: %v", lineCount, record)
	}

}

func isBlankRow(strs []string) bool {
	for _, field := range strs {
		if !mystr.CompareIgnoreCase("", field) {
			return false // 只要有一个字段不为空，就不是空白行
		}
	}
	return true // 所有字段都为空，是空白行
}
