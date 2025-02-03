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
	Types     []ReportType
	LogLevel  DEImp
	Directory string
}

func (r *Report) Report(p *Pipeline) error {

	for _, t := range r.Types {

		switch t {

		case JSON:
			dirPath := filepath.Join(r.Directory, p.Name)
			_, err := os.Stat(dirPath)
			if err != nil {
				err := os.MkdirAll(dirPath, os.ModePerm)

				if err != nil {
					fmt.Printf("err: %v\n", err)
					return err
				}
			}
			fileName := fmt.Sprintf("%s-%s.json", p.StartTime.Format(FILE_DATE_TIME_LAYOUT), p.GetId())
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
