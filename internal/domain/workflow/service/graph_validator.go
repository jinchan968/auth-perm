package service

import (
	"encoding/json"

	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/vo"
)

type ValidationError struct {
	NodeID  string `json:"node_id,omitempty"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

func ValidateFlowGraph(graph *vo.FlowGraph) []ValidationError {
	var errs []ValidationError

	triggerCount := 0
	outputCount := 0
	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeTrigger {
			triggerCount++
		}
		if node.Type == constant.NodeTypeOutput {
			outputCount++
		}
	}

	if triggerCount == 0 {
		errs = append(errs, ValidationError{Message: "缺少 trigger 节点", Level: "error"})
	} else if triggerCount > 1 {
		errs = append(errs, ValidationError{Message: "只能有一个 trigger 节点", Level: "error"})
	}

	if outputCount == 0 {
		errs = append(errs, ValidationError{Message: "缺少 output 节点", Level: "error"})
	}

	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeMap := make(map[string]vo.FlowNode)
	conditionHandles := make(map[string]map[string]bool)

	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
		inDegree[node.ID] = 0
	}

	for _, edge := range graph.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
		if sourceNode, ok := nodeMap[edge.Source]; ok && sourceNode.Type == constant.NodeTypeCondition {
			handle := conditionEdgeHandle(edge.SourceHandle)
			if handle == "" {
				errs = append(errs, ValidationError{
					NodeID:  edge.Source,
					Message: "condition 出边缺少分支 handle",
					Level:   "error",
				})
				continue
			}
			if _, ok := conditionHandles[edge.Source]; !ok {
				conditionHandles[edge.Source] = make(map[string]bool)
			}
			conditionHandles[edge.Source][handle] = true
		}
	}

	if triggerCount == 1 && outputCount > 0 {
		var triggerID string
		for _, node := range graph.Nodes {
			if node.Type == constant.NodeTypeTrigger {
				triggerID = node.ID
				break
			}
		}
		reachable := bfsReachable(triggerID, adj)
		for _, node := range graph.Nodes {
			if node.Type == constant.NodeTypeOutput {
				if !reachable[node.ID] {
					errs = append(errs, ValidationError{
						NodeID:  node.ID,
						Message: "output 不可达",
						Level:   "error",
					})
				}
			}
		}
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeTrigger {
			continue
		}
		if inDegree[node.ID] == 0 {
			errs = append(errs, ValidationError{
				NodeID:  node.ID,
				Message: "孤立节点",
				Level:   "warning",
			})
		}
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeCondition {
			outCount := len(adj[node.ID])
			if outCount < 2 {
				errs = append(errs, ValidationError{
					NodeID:  node.ID,
					Message: "condition 需至少 2 条出边",
					Level:   "error",
				})
			}
			validateConditionHandles(node, conditionHandles[node.ID], &errs)
		}
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeMerge {
			if inDegree[node.ID] < 2 {
				errs = append(errs, ValidationError{
					NodeID:  node.ID,
					Message: "merge 需至少 2 条入边",
					Level:   "error",
				})
			}
		}
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeLLM {
			var data map[string]interface{}
			if err := json.Unmarshal(node.Data, &data); err == nil {
				if modelID, ok := data["model_id"].(string); !ok || modelID == "" {
					errs = append(errs, ValidationError{
						NodeID:  node.ID,
						Message: "LLM 需指定 model_id",
						Level:   "error",
					})
				}
			}
		}
	}

	if hasCycle(graph.Nodes, graph.Edges) {
		errs = append(errs, ValidationError{Message: "存在环路", Level: "error"})
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeOutput && inDegree[node.ID] == 0 {
			errs = append(errs, ValidationError{
				NodeID:  node.ID,
				Message: "output 需至少 1 条入边",
				Level:   "error",
			})
		}
	}

	for _, node := range graph.Nodes {
		if node.Type == constant.NodeTypeTrigger && inDegree[node.ID] > 0 {
			errs = append(errs, ValidationError{
				NodeID:  node.ID,
				Message: "trigger 不应有入边",
				Level:   "error",
			})
		}
	}

	return errs
}

func validateConditionHandles(node vo.FlowNode, connectedHandles map[string]bool, errs *[]ValidationError) {
	var data vo.ConditionNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		*errs = append(*errs, ValidationError{
			NodeID:  node.ID,
			Message: "condition 配置解析失败",
			Level:   "error",
		})
		return
	}
	if data.DefaultHandle == "" {
		data.DefaultHandle = "default"
	}
	for _, branch := range data.Branches {
		if branch.Handle == "" {
			*errs = append(*errs, ValidationError{
				NodeID:  node.ID,
				Message: "condition 分支缺少 handle",
				Level:   "error",
			})
			continue
		}
		if !connectedHandles[branch.Handle] {
			*errs = append(*errs, ValidationError{
				NodeID:  node.ID,
				Message: "condition 分支未连接: " + branch.Handle,
				Level:   "error",
			})
		}
	}
	if !connectedHandles[data.DefaultHandle] {
		*errs = append(*errs, ValidationError{
			NodeID:  node.ID,
			Message: "condition 默认分支未连接: " + data.DefaultHandle,
			Level:   "error",
		})
	}
}

func bfsReachable(start string, adj map[string][]string) map[string]bool {
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, next := range adj[curr] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return visited
}

func hasCycle(nodes []vo.FlowNode, edges []vo.FlowEdge) bool {
	adj := make(map[string][]string)
	for _, edge := range edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
	}
	state := make(map[string]int)
	var dfs func(nodeID string) bool
	dfs = func(nodeID string) bool {
		state[nodeID] = 1
		for _, next := range adj[nodeID] {
			if state[next] == 1 {
				return true
			}
			if state[next] == 0 && dfs(next) {
				return true
			}
		}
		state[nodeID] = 2
		return false
	}
	for _, node := range nodes {
		if state[node.ID] == 0 {
			if dfs(node.ID) {
				return true
			}
		}
	}
	return false
}
