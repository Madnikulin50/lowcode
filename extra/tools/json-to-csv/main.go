package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/madnikulin50/lowcode/server/pkg/cli"
)

var (
	verbose      bool
	showHelp     bool
	jsonFilePath string
	outputBase   string
)

func init() {
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&verbose, "v", false, "be verbose")
	flag.StringVar(&jsonFilePath, "in", "./data.json", "location of json data file")
	flag.StringVar(&outputBase, "out", ".", "base dir for output")
	flag.Parse()
}

// Takes JSON input with codegen tasks and definitions and generates files
func main() {
	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	reader, err := os.Open(jsonFilePath)
	if err != nil {
		cli.HandleError(fmt.Errorf("failed to decode input from: %v", err))
		os.Exit(1)
		return
	}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		cli.HandleError(fmt.Errorf("failed to decode input from: %v", err))
		os.Exit(1)
		return
	}
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		cli.HandleError(fmt.Errorf("failed to decode input from: %v", err))
		os.Exit(1)
		return
	}
	for k, v := range data {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		columnsMap := map[string]bool{}

		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			for k, _ := range m {
				columnsMap[k] = true
			}
		}
		columns := make([]string, 0, len(columnsMap))

		for column := range columnsMap {
			columns = append(columns, column)
		}
		if len(columns) == 0 {
			continue
		}

		fn := filepath.Join(outputBase, k) + ".csv"
		os.Remove(fn)
		file, err := os.Create(fn)
		if err != nil {
			cli.HandleError(fmt.Errorf("Could not create file %v:%v", fn, err))
			continue
		}
		defer file.Close()

		// 2. Initialize the CSV writer
		writer := csv.NewWriter(file)
		defer writer.Flush() // Ensure buffered data is written
		if err := writer.Write(columns); err != nil {
			cli.HandleError(fmt.Errorf("Could not create file %v", err))
			continue
		}
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			row := make([]string, len(columnsMap))
			for i, column := range columns {
				v, ok := m[column]
				if !ok {
					v = ""
				}
				row[i] = fmt.Sprintf("%v", v)
			}
			if err := writer.Write(row); err != nil {
				cli.HandleError(fmt.Errorf("Could not create file %v", err))
				continue
			}
		}

	}
}
