package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"auth-perm/internal/domain/workflow/constant"
	"auth-perm/internal/domain/workflow/vo"
)

func evaluateRuleGroup(content string, group vo.RuleGroup) (bool, error) {
	results := make([]bool, len(group.Rules))
	for i, rule := range group.Rules {
		if rule.SubGroup != nil {
			r, err := evaluateRuleGroup(content, *rule.SubGroup)
			if err != nil {
				return false, err
			}
			results[i] = r
			continue
		}
		r, err := evaluateSingleRule(content, rule)
		if err != nil {
			return false, err
		}
		results[i] = r
	}
	return combineLogic(group.Logic, results), nil
}

func evaluateSingleRule(content string, rule vo.RuleItem) (bool, error) {
	var result bool
	switch rule.Operator {
	case constant.OpContains:
		result = strings.Contains(content, rule.Value)
	case constant.OpNotContains:
		result = !strings.Contains(content, rule.Value)
	case constant.OpEquals:
		result = content == rule.Value
	case constant.OpMatches:
		re, err := regexp.Compile(rule.Value)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern %q: %w", rule.Value, err)
		}
		result = re.MatchString(content)
	case constant.OpStartsWith:
		result = strings.HasPrefix(content, rule.Value)
	case constant.OpEndsWith:
		result = strings.HasSuffix(content, rule.Value)
	case constant.OpGT:
		length := len([]rune(content))
		val, err := strconv.Atoi(rule.Value)
		if err != nil {
			return false, fmt.Errorf("OpGT: invalid value %q: %w", rule.Value, err)
		}
		result = length > val
	case constant.OpGTE:
		length := len([]rune(content))
		val, err := strconv.Atoi(rule.Value)
		if err != nil {
			return false, fmt.Errorf("OpGTE: invalid value %q: %w", rule.Value, err)
		}
		result = length >= val
	case constant.OpLT:
		length := len([]rune(content))
		val, err := strconv.Atoi(rule.Value)
		if err != nil {
			return false, fmt.Errorf("OpLT: invalid value %q: %w", rule.Value, err)
		}
		result = length < val
	case constant.OpLTE:
		length := len([]rune(content))
		val, err := strconv.Atoi(rule.Value)
		if err != nil {
			return false, fmt.Errorf("OpLTE: invalid value %q: %w", rule.Value, err)
		}
		result = length <= val
	case constant.OpIsEmpty:
		result = strings.TrimSpace(content) == ""
	case constant.OpNotEmpty:
		result = strings.TrimSpace(content) != ""
	default:
		return false, nil
	}
	if rule.Negate {
		result = !result
	}
	return result, nil
}

func combineLogic(logic string, results []bool) bool {
	if len(results) == 0 {
		return false
	}
	if logic == "AND" {
		for _, r := range results {
			if !r {
				return false
			}
		}
		return true
	}
	for _, r := range results {
		if r {
			return true
		}
	}
	return false
}
