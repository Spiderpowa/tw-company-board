package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/go-resty/resty/v2"
)

var client *resty.Client

// Board --
type Board struct {
	Position     string `json:"Person_Position_Name"`
	Name         string `json:"Person_Name"`
	JuristicName string `json:"Juristic_Person_Name"`
	Shares       int64  `json:"Person_Shareholding"`
}

// Boards --
type Boards []Board

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func fetch(account string) (Boards, error) {
	resp, err := client.
		R().
		SetResult(&Boards{}).
		SetPathParams(map[string]string{
			"filter": fmt.Sprintf("Business_Accounting_NO eq %s", account),
		}).
		Get("http://data.gcis.nat.gov.tw/od/data/api/4E5F7653-1B91-4DDC-99D5-468530FAE396?$format=json&$filter={filter}&$skip=0&$top=50")
	if err != nil {
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			return nil, nil
		}
		return nil, err
	}
	board, ok := resp.Result().(*Boards)
	if !ok {
		return nil, fmt.Errorf("unable to cast to type Board")
	}
	return *board, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:", os.Args[0], "target_name", "csv_file")
		os.Exit(1)
	}
	target := []byte(os.Args[1])
	targetFirst, _ := utf8.DecodeRune(target)
	targetLast, _ := utf8.DecodeLastRune(target)
	client = resty.New()
	fd, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	reader := csv.NewReader(fd)
	// Skip header.
	_, err = reader.Read()
	if err != nil {
		panic(err)
	}
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		no := string(record[1])
		if len(no) == 0 {
			continue
		}
		storageFile := fmt.Sprintf("output/%s.json", no)
		if exists(storageFile) {
			continue
		}
		func() {
			fmt.Println("Fetching", no)
			boards, err := fetch(no)
			if err != nil {
				fmt.Println(err)
				return
			}
			out, err := os.Create(storageFile)
			if err != nil {
				fmt.Println("Error creating file", storageFile, err)
				return
			}
			defer out.Close()
			enc := json.NewEncoder(out)
			if err := enc.Encode(boards); err != nil {
				panic(err)
			}
			for _, board := range boards {
				first, _ := utf8.DecodeRune([]byte(board.Name))
				last, _ := utf8.DecodeLastRune([]byte(board.Name))
				if first == targetFirst && last == targetLast {
					fmt.Println(record)
					break
				}
			}
		}()
		// break
	}
}
