package service

import (
	"encoding/json"
	"testing"

	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/vo"
)

func TestValidateFlowGraphRequiresConditionHandles(t *testing.T) {
	graph := flowGraphWithCondition(nil, ptr("default"))

	errs := ValidateFlowGraph(&graph)
	if !hasValidationError(errs, "condition 出边缺少分支 handle") {
		t.Fatalf("expected missing condition handle error, got %#v", errs)
	}
}

func TestValidateFlowGraphAcceptsConnectedConditionHandles(t *testing.T) {
	graph := flowGraphWithCondition(ptr("vip"), ptr("default"))

	errs := ValidateFlowGraph(&graph)
	for _, err := range errs {
		if err.Level == "error" {
			t.Fatalf("expected no validation errors, got %#v", errs)
		}
	}
}

func TestBuildConditionTargetsNormalizesLegacyDefaultHandle(t *testing.T) {
	graph := flowGraphWithCondition(ptr("vip"), ptr("__default"))
	nodeMap := map[string]vo.FlowNode{}
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	targets, err := buildConditionTargets(graph.Edges, nodeMap)
	if err != nil {
		t.Fatalf("buildConditionTargets returned error: %v", err)
	}
	if targets["condition"]["default"] != "output-default" {
		t.Fatalf("expected legacy default handle to map to output-default, got %#v", targets)
	}
}

func flowGraphWithCondition(branchHandle, defaultHandle *string) vo.FlowGraph {
	conditionData, _ := json.Marshal(vo.ConditionNodeData{
		DefaultHandle: "default",
		Branches: []vo.BranchConfig{
			{
				Handle: "vip",
				Rule: vo.RuleGroup{
					Logic: "AND",
					Rules: []vo.RuleItem{{Field: "content", Operator: constant.OpContains, Value: "vip"}},
				},
			},
		},
	})
	outputData := json.RawMessage(`{"format":"text","join_mode":"or"}`)

	return vo.FlowGraph{
		Nodes: []vo.FlowNode{
			{ID: "trigger", Type: constant.NodeTypeTrigger, Data: json.RawMessage(`{}`)},
			{ID: "condition", Type: constant.NodeTypeCondition, Data: conditionData},
			{ID: "output-vip", Type: constant.NodeTypeOutput, Data: outputData},
			{ID: "output-default", Type: constant.NodeTypeOutput, Data: outputData},
		},
		Edges: []vo.FlowEdge{
			{ID: "trigger-condition", Source: "trigger", Target: "condition"},
			{ID: "condition-vip", Source: "condition", Target: "output-vip", SourceHandle: branchHandle},
			{ID: "condition-default", Source: "condition", Target: "output-default", SourceHandle: defaultHandle},
		},
	}
}

func hasValidationError(errs []ValidationError, message string) bool {
	for _, err := range errs {
		if err.Level == "error" && err.Message == message {
			return true
		}
	}
	return false
}

func ptr(value string) *string {
	return &value
}
