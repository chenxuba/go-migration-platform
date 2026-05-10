package service

import (
	"math"
	"strings"
	"unicode/utf8"

	"go-migration-platform/services/education/internal/model"
)

const (
	deepSeekV4ProInputPricePer1MTokensCNY         = 3.0
	deepSeekV4ProInputCacheHitPricePer1MTokensCNY = 0.025
	deepSeekV4ProOutputPricePer1MTokensCNY        = 6.0
)

func toDeepSeekUsageVO(usage *deepSeekUsage) *model.DeepSeekUsageVO {
	if usage == nil {
		return nil
	}
	return &model.DeepSeekUsageVO{
		PromptTokens:          usage.PromptTokens,
		CompletionTokens:      usage.CompletionTokens,
		PromptCacheHitTokens:  usage.PromptCacheHitTokens,
		PromptCacheMissTokens: usage.PromptCacheMissTokens,
		TotalTokens:           usage.TotalTokens,
	}
}

func computeDeepSeekUsageCostCNY(usage *model.DeepSeekUsageVO, modelName string) float64 {
	if usage == nil {
		return 0
	}
	switch strings.TrimSpace(modelName) {
	case "", deepSeekIEPPlanModel:
		inputMiss := float64(deepSeekMaxInt(usage.PromptCacheMissTokens, usage.PromptTokens-usage.PromptCacheHitTokens))
		inputHit := float64(usage.PromptCacheHitTokens)
		output := float64(usage.CompletionTokens)
		total := inputMiss*deepSeekV4ProInputPricePer1MTokensCNY/1_000_000 +
			inputHit*deepSeekV4ProInputCacheHitPricePer1MTokensCNY/1_000_000 +
			output*deepSeekV4ProOutputPricePer1MTokensCNY/1_000_000
		return roundCurrency(total)
	default:
		inputMiss := float64(deepSeekMaxInt(usage.PromptCacheMissTokens, usage.PromptTokens-usage.PromptCacheHitTokens))
		inputHit := float64(usage.PromptCacheHitTokens)
		output := float64(usage.CompletionTokens)
		total := inputMiss*deepSeekV4ProInputPricePer1MTokensCNY/1_000_000 +
			inputHit*deepSeekV4ProInputCacheHitPricePer1MTokensCNY/1_000_000 +
			output*deepSeekV4ProOutputPricePer1MTokensCNY/1_000_000
		return roundCurrency(total)
	}
}

func estimateDeepSeekOutputCostCNY(text string) float64 {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	runeCount := utf8.RuneCountInString(text)
	estimatedTokens := math.Ceil(float64(runeCount) * 0.9)
	total := estimatedTokens * deepSeekV4ProOutputPricePer1MTokensCNY / 1_000_000
	return roundCurrency(total)
}

func ComputeDeepSeekUsageCostCNY(usage *model.DeepSeekUsageVO, modelName string) float64 {
	return computeDeepSeekUsageCostCNY(usage, modelName)
}

func EstimateDeepSeekOutputCostCNY(text string) float64 {
	return estimateDeepSeekOutputCostCNY(text)
}

func roundCurrency(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func deepSeekMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
