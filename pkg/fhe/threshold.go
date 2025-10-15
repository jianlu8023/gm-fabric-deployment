package fhe

import (
	"fmt"
	"math"
)

// ThresholdAnalysis 分析阈值方案的安全性和可用性
type ThresholdAnalysis struct {
	N int // 总参与方数
	t int // 阈值
}

// NewThresholdAnalysis 创建新的阈值分析实例
func NewThresholdAnalysis(N, t int) *ThresholdAnalysis {
	return &ThresholdAnalysis{
		N: N,
		t: t,
	}
}

// SecurityLevel 计算安全级别（可容忍的敌手数量）
func (ta *ThresholdAnalysis) SecurityLevel() int {
	return ta.t - 1
}

// AvailabilityLevel 计算可用性级别（可容忍的故障数量）
func (ta *ThresholdAnalysis) AvailabilityLevel() int {
	return ta.N - ta.t
}

// Efficiency 计算效率指标（参与操作的最小参与方比例）
func (ta *ThresholdAnalysis) Efficiency() float64 {
	return float64(ta.t) / float64(ta.N)
}

// Analyze 分析给定t值的特性
func (ta *ThresholdAnalysis) Analyze() {
	fmt.Printf("=== N=%d, t=%d 的阈值方案分析 ===\n", ta.N, ta.t)
	fmt.Printf("安全级别: 可容忍 %d 个敌手\n", ta.SecurityLevel())
	fmt.Printf("可用性级别: 可容忍 %d 个故障\n", ta.AvailabilityLevel())
	fmt.Printf("效率指标: %.2f%% 的参与方需要在线\n", ta.Efficiency()*100)
	fmt.Printf("安全边际: %.2f%%\n", math.Abs(ta.Efficiency()-0.5)*100)
	fmt.Println()
}

func exampleThreshold() {
	N := 100 // 总参与方数

	fmt.Println("100个参与方的阈值方案分析")
	fmt.Println("========================")

	// 分析几种典型的t值设置
	tValues := []int{2, 10, 25, 34, 50, 67, 75, 90, 99}

	for _, t := range tValues {
		if t >= 2 && t <= N-1 {
			analysis := NewThresholdAnalysis(N, t)
			analysis.Analyze()
		}
	}

	// 推荐设置
	fmt.Println("推荐的t值设置:")
	fmt.Println("1. t=50 (50%阈值) - 最佳平衡点")
	fmt.Println("2. t=34 (34%阈值) - 更高可用性")
	fmt.Println("3. t=67 (67%阈值) - 更高安全性")
}