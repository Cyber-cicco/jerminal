package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/Cyber-cicco/jerminal/server/rpc"
	"github.com/google/uuid"
)

// unMarshallFileFromReq uses a custom unMarshalling process to ommit non
// wanted fields
func unMarshallFileFromReq(req rpc.GetReportsReq, directory string, id string) (map[string]interface{}, error) {
	if uuid.Validate(id) != nil {
        return nil, errors.New("Invalid identifier")
	}

	file, err := os.Open(filepath.Join(directory, id+".json"))
	if err != nil {
        return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	wantedFields := make(map[string]bool)
	if len(req.Params.Fields) > 0 {
		for _, field := range req.Params.Fields {
			wantedFields[field] = true
		}
	} else if len(req.Params.OmittedFields) > 0 {

		allFields := [9]string{"name", "agent", "id", "parent", "time-ran", "in-error", "start-time", "diagnostics", "elapsed-time"}

		omitMap := make(map[string]bool)
		for _, field := range req.Params.OmittedFields {
			omitMap[field] = true
		}
		for _, field := range allFields {
			if _, ok := omitMap[field]; !ok {
				wantedFields[field] = true
			}
		}
	}

	// Selective decoding
	result := make(map[string]interface{})

	// Read opening bracket
	_, err = decoder.Token()
	if err != nil {
        return nil, err
	}

	// Read object contents
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
            return nil, err
		}

		if keyStr, ok := key.(string); ok {
            wanted, ok := wantedFields[keyStr]
			if len(wantedFields) == 0 || (ok && wanted)  {
				// Decode only if we want this field
				var value interface{}
				if err := decoder.Decode(&value); err != nil {
                    return nil, err
				}
				result[keyStr] = value
			} else {
				// Skip this value
				if _, err := decoder.Token(); err != nil {
                    return nil, err
				}
			}
		}
	}

	return result, nil
}
