package vo

import (
	"encoding/json"
	"errors"
)

type FlowNode struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Position NodePosition    `json:"position"`
	Data     json.RawMessage `json:"data"`
}

type FlowEdge struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Target       string  `json:"target"`
	SourceHandle *string `json:"sourceHandle,omitempty"`
	TargetHandle *string `json:"targetHandle,omitempty"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FlowViewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

type FlowGraph struct {
	Nodes    []FlowNode   `json:"nodes"`
	Edges    []FlowEdge   `json:"edges"`
	Viewport FlowViewport `json:"viewport"`
}

func (g *FlowGraph) Validate() error {
	if len(g.Nodes) == 0 {
		return errors.New("empty graph")
	}
	return nil
}
