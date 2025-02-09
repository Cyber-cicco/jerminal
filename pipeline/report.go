package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

)

type ReportType uint8

const (
	JSON = ReportType(iota)
	HTML
	SQLITE
)

// Allows for creating reports in specified format
type Report struct {
	Types     []ReportType `json:"types"`
	LogLevel  DEImp        `json:"-"`
}

func (r *Report) Report(p *Pipeline) error {

	for _, t := range r.Types {

		switch t {

		case JSON:
			dirPath := filepath.Join(p.globalState.ReportDir, p.Name)
			_, err := os.Stat(dirPath)
			if err != nil {
				err := os.MkdirAll(dirPath, os.ModePerm)

				if err != nil {
					fmt.Printf("err: %v\n", err)
					return err
				}
			}
			fileName := p.GetId() + ".json"
			clone := *p
			clone.Diagnostic = clone.Diagnostic.FilterBasedOnImportance(r.LogLevel)
			filePath := filepath.Join(dirPath, fileName)
			fileContent, err := json.MarshalIndent(
				clone,
				"",
				"  ",
			)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return err
			}

			err = os.WriteFile(filePath, fileContent, 0644)
			return err
		default:
			return fmt.Errorf("Not yet supported")

		}
	}
	return nil
}
