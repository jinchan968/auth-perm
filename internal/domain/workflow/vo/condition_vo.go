package vo

import (
	"auth-perm/internal/domain/workflow/constant"
	"errors"
)

type RuleGroup struct {
	Logic string     `json:"logic"`
	Rules []RuleItem `json:"rules"`
}

type RuleItem struct {
	Field    string     `json:"field,omitempty"`
	Operator string     `json:"operator,omitempty"`
	Value    string     `json:"value,omitempty"`
	Negate   bool       `json:"negate,omitempty"`
	SubGroup *RuleGroup `json:"sub_group,omitempty"`
}

type BranchConfig struct {
	Handle string    `json:"handle"`
	Rule   RuleGroup `json:"rule"`
}

type ConditionNodeData struct {
	Branches      []BranchConfig `json:"branches"`
	DefaultHandle string         `json:"default_handle"`
}

func (r *RuleGroup) Validate() error {
	if r.Logic != "AND" && r.Logic != "OR" {
		return errors.New("invalid logic")
	}
	if len(r.Rules) == 0 {
		return errors.New("empty rules")
	}
	for _, rule := range r.Rules {
		if rule.SubGroup != nil {
			if err := rule.SubGroup.Validate(); err != nil {
				return err
			}
			continue
		}
		if !constant.IsValidOperator(rule.Operator) {
			return errors.New("invalid operator: " + rule.Operator)
		}
		if rule.Field == "" {
			return errors.New("field required")
		}
	}
	return nil
}
