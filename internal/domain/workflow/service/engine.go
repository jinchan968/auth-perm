package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/dm"
	"auth-perm/internal/domain/workflow/repo"
	"auth-perm/internal/domain/workflow/vo"
	"auth-perm/internal/infra/opencode"

	"github.com/cloudwego/eino/compose"
)

type NodeOutput struct {
	NodeID    string                 `json:"node_id"`
	NodeType  string                 `json:"node_type"`
	Content   string                 `json:"content"`
	ModelName string                 `json:"model_name,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type Engine struct {
	openCode *opencode.Client
	wsHub    *WSHub
	runRepo  *repo.WorkflowRunRepo
	nodeRepo *repo.WorkflowRunNodeRepo
}

func NewEngine(
	oc *opencode.Client,
	hub *WSHub,
	rr *repo.WorkflowRunRepo,
	nr *repo.WorkflowRunNodeRepo,
) *Engine {
	return &Engine{openCode: oc, wsHub: hub, runRepo: rr, nodeRepo: nr}
}

func (e *Engine) Execute(ctx context.Context, runID string, flowJSON string, inputText string) (*NodeOutput, error) {
	var graph vo.FlowGraph
	if err := json.Unmarshal([]byte(flowJSON), &graph); err != nil {
		return nil, fmt.Errorf("parse flow_json: %w", err)
	}

	errs := ValidateFlowGraph(&graph)
	var fatalErrs []ValidationError
	for _, err := range errs {
		if err.Level == "error" {
			fatalErrs = append(fatalErrs, err)
		}
	}
	if len(fatalErrs) > 0 {
		return nil, fmt.Errorf("validation failed: %d errors", len(fatalErrs))
	}

	g := compose.NewGraph[string, *NodeOutput]()
	nodeMap := make(map[string]vo.FlowNode)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	for _, node := range graph.Nodes {
		if err := e.registerNode(g, node, runID); err != nil {
			return nil, fmt.Errorf("register node %s: %w", node.ID, err)
		}
	}

	for _, edge := range graph.Edges {
		sourceNode := nodeMap[edge.Source]
		if sourceNode.Type == constant.NodeTypeCondition {
			continue
		}
		g.AddEdge(edge.Source, edge.Target)
	}

	runnable, err := g.Compile(ctx, compose.WithGraphName("workflow_"+runID))
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}

	result, err := runnable.Invoke(ctx, inputText)
	if err != nil {
		return nil, fmt.Errorf("invoke graph: %w", err)
	}

	return result, nil
}

func (e *Engine) registerNode(g *compose.Graph[string, *NodeOutput], node vo.FlowNode, runID string) error {
	switch node.Type {
	case constant.NodeTypeTrigger:
		g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, input string) (*NodeOutput, error) {
			return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: input}, nil
		}))

	case constant.NodeTypeLLM:
		var data struct {
			ModelID       string  `json:"model_id"`
			SystemPrompt  string  `json:"system_prompt"`
			Temperature   float64 `json:"temperature"`
			ReasoningMode string  `json:"reasoning_mode"`
		}
		if err := json.Unmarshal(node.Data, &data); err != nil {
			return fmt.Errorf("parse llm node data: %w", err)
		}

		g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
			e.writeNodeStart(runID, node.ID, node.Type, in.Content)
			start := time.Now()
			content, err := e.openCode.Chat(ctx, data.ModelID, data.SystemPrompt, in.Content, false, data.ReasoningMode)
			duration := time.Since(start).Milliseconds()
			if err != nil {
				e.writeNodeEnd(runID, node.ID, node.Type, "", err.Error(), duration)
				return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: err.Error()}, nil
			}
			if content == "" {
				e.writeNodeEnd(runID, node.ID, node.Type, "", "llm returned empty response", duration)
				return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: "llm returned empty response"}, nil
			}
			e.writeNodeEnd(runID, node.ID, node.Type, content, "", duration)
			return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: content, ModelName: data.ModelID}, nil
		}))

	case constant.NodeTypeCondition:
		var condData vo.ConditionNodeData
		if err := json.Unmarshal(node.Data, &condData); err != nil {
			return err
		}
		handleNames := make(map[string]bool)
		for _, branch := range condData.Branches {
			handleNames[branch.Handle] = true
		}
		handleNames[condData.DefaultHandle] = true

		branchFunc := compose.NewGraphBranch(
			func(ctx context.Context, in *NodeOutput) (string, error) {
				for _, branch := range condData.Branches {
					matched, err := evaluateRuleGroup(in.Content, branch.Rule)
					if err != nil {
						continue
					}
					if matched {
						return branch.Handle, nil
					}
				}
				return condData.DefaultHandle, nil
			},
			handleNames,
		)
		g.AddBranch(node.ID, branchFunc)

	case constant.NodeTypeTransform:
		var data struct {
			Operation string                 `json:"operation"`
			Params    map[string]interface{} `json:"params"`
		}
		if err := json.Unmarshal(node.Data, &data); err != nil {
			return fmt.Errorf("parse transform node data: %w", err)
		}

		g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
			result, err := e.executeTransform(in.Content, data.Operation, data.Params)
			if err != nil {
				return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: err.Error()}, nil
			}
			return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: result}, nil
		}))

	case constant.NodeTypeMerge:
		var data struct {
			Strategy  string `json:"strategy"`
			Delimiter string `json:"delimiter,omitempty"`
		}
		if err := json.Unmarshal(node.Data, &data); err != nil {
			return fmt.Errorf("parse merge node data: %w", err)
		}

		g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
			results := collectPredecessorResults(ctx)
			merged := e.executeMerge(results, data.Strategy, data.Delimiter)
			return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: merged}, nil
		}))

	case constant.NodeTypeOutput:
		var data struct {
			Format   string `json:"format"`
			JoinMode string `json:"join_mode"`
		}
		if err := json.Unmarshal(node.Data, &data); err != nil {
			return fmt.Errorf("parse output node data: %w", err)
		}

		if data.JoinMode == constant.JoinModeOr {
			g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
				results := collectPredecessorResults(ctx)
				for _, res := range results {
					if res.Error == "" {
						return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: res.Content}, nil
					}
				}
				return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Error: "all branches failed"}, nil
			}))
		} else {
			g.AddLambdaNode(node.ID, compose.InvokableLambda(func(ctx context.Context, in *NodeOutput) (*NodeOutput, error) {
				results := collectPredecessorResults(ctx)
				var contents []string
				for _, res := range results {
					if res.Error == "" {
						contents = append(contents, res.Content)
					}
				}
				merged := strings.Join(contents, "\n\n---\n\n")
				return &NodeOutput{NodeID: node.ID, NodeType: node.Type, Content: merged}, nil
			}))
		}

	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}

	return nil
}

func (e *Engine) executeTransform(content, operation string, params map[string]interface{}) (string, error) {
	switch operation {
	case constant.TransformRegexExtract:
		pattern, ok := params["pattern"].(string)
		if !ok {
			return "", fmt.Errorf("transform regex_extract: pattern must be string")
		}
		groupIndex := 0
		if gi, ok := params["group_index"]; ok {
			if gf, ok := gi.(float64); ok {
				groupIndex = int(gf)
			}
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", err
		}
		matches := re.FindStringSubmatch(content)
		if len(matches) > groupIndex {
			return matches[groupIndex], nil
		}
		return "", nil

	case constant.TransformRegexReplace:
		pattern, ok := params["pattern"].(string)
		if !ok {
			return "", fmt.Errorf("transform regex_replace: pattern must be string")
		}
		replacement, ok := params["replacement"].(string)
		if !ok {
			return "", fmt.Errorf("transform regex_replace: replacement must be string")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return "", err
		}
		return re.ReplaceAllString(content, replacement), nil

	case constant.TransformTrim:
		return strings.TrimSpace(content), nil

	case constant.TransformMarkdownToText:
		result := regexp.MustCompile(`\*\*|\*|__|_|#|\[|\]|\(|\)`).ReplaceAllString(content, "")
		return result, nil

	case constant.TransformExtractJSON:
		re := regexp.MustCompile(`(?s)\{.*\}`)
		match := re.FindString(content)
		return match, nil

	case constant.TransformTruncate:
		maxLen := 100
		if ml, ok := params["max_length"]; ok {
			if mf, ok := ml.(float64); ok {
				maxLen = int(mf)
			}
		}
		runes := []rune(content)
		if len(runes) > maxLen {
			return string(runes[:maxLen]) + "...", nil
		}
		return content, nil

	case constant.TransformToUppercase:
		return strings.ToUpper(content), nil

	case constant.TransformToLowercase:
		return strings.ToLower(content), nil

	default:
		return content, nil
	}
}

func (e *Engine) executeMerge(results []*NodeOutput, strategy, delimiter string) string {
	switch strategy {
	case constant.MergeStrategyConcat:
		var contents []string
		for _, res := range results {
			if res.Error == "" {
				contents = append(contents, res.Content)
			}
		}
		return strings.Join(contents, "\n\n---\n\n")

	case constant.MergeStrategyFirst:
		for _, res := range results {
			if res.Error == "" {
				return res.Content
			}
		}
		return ""

	case constant.MergeStrategyJoin:
		var contents []string
		for _, res := range results {
			if res.Error == "" {
				contents = append(contents, res.Content)
			}
		}
		if delimiter == "" {
			delimiter = "\n"
		}
		return strings.Join(contents, delimiter)

	default:
		return ""
	}
}

// collectPredecessorResults 收集当前节点所有前驱节点的执行结果
// TODO: Eino Graph 的 lambda 只接收单条边的输入，不暴露前驱上下文。
//       需要改造引擎为每个节点输出缓存到 run-level map，并在闭包中传入节点前驱列表。
//       当前实现下 merge/output 节点只能拿到单条边的输入，多路汇聚功能不完整。
func collectPredecessorResults(ctx context.Context) []*NodeOutput {
	return nil
}

func (e *Engine) writeNodeStart(runID, nodeID, nodeType, input string) {
	e.wsHub.Broadcast(runID, map[string]interface{}{
		"type":      "node_start",
		"node_id":   nodeID,
		"node_type": nodeType,
	})
	now := time.Now()
	inputJSON, err := json.Marshal(map[string]string{"content": input})
	if err != nil {
		inputJSON = []byte("{}")
	}
	e.nodeRepo.Create(&dm.WorkflowRunNodeDO{
		RunID:      runID,
		NodeID:     nodeID,
		NodeType:   nodeType,
		Status:     constant.StatusRunning,
		InputJSON:  string(inputJSON),
		StartedAt:  &now,
	})
}

func (e *Engine) writeNodeEnd(runID, nodeID, nodeType, output, errStr string, durationMs int64) {
	now := time.Now()
	status := constant.StatusSuccess
	if errStr != "" {
		status = constant.StatusFailed
	}
	msgType := "node_end"
	if errStr != "" {
		msgType = "node_error"
	}
	msg := map[string]interface{}{
		"type":        msgType,
		"node_id":     nodeID,
		"node_type":   nodeType,
		"duration_ms": durationMs,
	}
	if errStr != "" {
		msg["error"] = errStr
	}
	e.wsHub.Broadcast(runID, msg)

	outputJSON, err := json.Marshal(map[string]string{"content": output})
	if err != nil {
		outputJSON = []byte("{}")
	}
	e.nodeRepo.Update(&dm.WorkflowRunNodeDO{
		RunID:      runID,
		NodeID:     nodeID,
		NodeType:   nodeType,
		Status:     status,
		OutputJSON: string(outputJSON),
		Error:      errStr,
		FinishedAt: &now,
		DurationMs: int(durationMs),
	})
}


