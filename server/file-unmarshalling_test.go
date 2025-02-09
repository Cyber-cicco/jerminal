package server

import (
	"fmt"
	"testing"

	"github.com/Cyber-cicco/jerminal/server/rpc"
)

func TestUnmarshall(t *testing.T) {

	pipelineId := "0b91fe0e-9ad6-4b37-bdfe-b22f2e8cd248"
    pipelineName := "test3"
	req := rpc.GetReportsReq{
		JRPCRequest: rpc.JRPCRequest{
			JsonRpcVersion: "2.0",
			Id:             0,
			Method:         "get-reports",
		},
		Params: rpc.GetReportsParams{
			PipelineId:    &pipelineId,
			PipelineName:  &pipelineName,
			Type:          "json",
			Fields:        []string{},
			OmittedFields: []string{"id"},
		},
	}
    res, err := unMarshallFileFromReq(&req, "../reports/test3", *req.Params.PipelineId)
    if err != nil {
        t.Fatalf("Expected no error got %v", err)
    }

    if _, ok := res["id"]; ok {
        t.Fatalf("Field id should have been omitted, got %v", res["id"])
    }

    fmt.Println("Results : ")
    for key, val := range res {
        fmt.Printf("%v : %s\n", key, val)
    }

    if name, ok := res["name"]; !ok || name != "test3" {
        t.Fatalf("Field name should exist")
    }
}
