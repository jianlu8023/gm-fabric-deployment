package fhe

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/tuneinsight/lattigo/v6/multiparty"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// simulateThresholdKeyGen 模拟阈值密钥生成
func simulateThresholdKeyGen(P []party, t int, params bgv.Parameters) {
	fmt.Printf("正在为%d个参与方生成阈值密钥 (t=%d)...\n", len(P), t)

	thresholdizer := multiparty.NewThresholdizer(params.Parameters)

	// 为每个参与方生成份额
	for i, pi := range P {
		pol, err := thresholdizer.GenShamirPolynomial(t, pi.sk)
		if err != nil {
			log.Fatal(err)
		}

		// 分发份额给其他参与方
		for j := range P {
			if i != j {
				share := thresholdizer.AllocateThresholdSecretShare()
				thresholdizer.GenShamirSecretShare(P[j].shamirPt, pol, &share)
				// 在实际应用中，这里会通过安全信道发送份额
			}
		}
	}

	fmt.Printf("阈值密钥生成完成\n\n")
}

// getOnlineParties 随机选择t个在线参与方
func getOnlineParties(P []party, t int) []party {
	if t > len(P) {
		t = len(P)
	}

	// 创建索引切片并随机打乱
	indices := make([]int, len(P))
	for i := range indices {
		indices[i] = i
	}

	// 随机打乱索引
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// 选择前t个
	online := make([]party, t)
	for i := 0; i < t; i++ {
		online[i] = P[indices[i]]
	}

	return online
}

func examplePartiesThreshold() {
	fmt.Println("100个参与方BFV阈值方案示例")
	fmt.Println("========================")

	// 设置参数
	params, err := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
		LogN:             12,        // 较小的参数以加快演示速度
		LogQ:             []int{50}, // 简化的参数
		LogP:             []int{50}, // 简化的参数
		PlaintextModulus: 65537,     // 常用的明文模数
	})
	if err != nil {
		log.Fatal(err)
	}

	N := 100 // 参与方总数

	// 分析不同的t值设置
	tValues := []int{2, 25, 50, 75, 99}

	for _, t := range tValues {
		fmt.Printf("场景: %d个参与方，阈值t=%d\n", N, t)
		fmt.Printf("--------------------------------\n")

		// 创建参与方
		P := genParties(params, N, t)
		fmt.Printf("已创建%d个参与方\n", len(P))

		// 模拟阈值密钥生成
		simulateThresholdKeyGen(P, t, params)

		// 模拟选择在线参与方
		onlineParties := getOnlineParties(P, t)
		fmt.Printf("选择了%d个在线参与方进行操作\n", len(onlineParties))

		// 显示安全性和可用性指标
		security := t - 1
		availability := N - t
		fmt.Printf("安全级别: 可容忍%d个敌手\n", security)
		fmt.Printf("可用性级别: 可容忍%d个故障\n", availability)
		fmt.Printf("操作效率: %.1f%%的参与方需要在线\n", float64(t)/float64(N)*100)

		fmt.Println()
	}

	fmt.Println("推荐的t值设置:")
	fmt.Println("1. t=50 (50%阈值) - 安全性与可用性的最佳平衡")
	fmt.Println("2. t=25 (25%阈值) - 更高可用性，适合对可用性要求高的场景")
	fmt.Println("3. t=75 (75%阈值) - 更高安全性，适合对安全性要求高的场景")
}
