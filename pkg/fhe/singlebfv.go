package fhe

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func ExampleSingleBfv() {
	fmt.Println("开始BFV示例...")

	var err error
	var params bgv.Parameters

	// 使用BFV的示例参数（scale-invariant BGV）
	// 128位安全参数，支持深度为3的电路。
	// LogN:13, LogQP: 218。
	if params, err = bgv.NewParametersFromLiteral(
		bgv.ParametersLiteral{
			LogN:             13,                // log2(环度)
			LogQ:             []int{55, 54, 54}, // log2(素数Q) (密文模数)
			LogP:             []int{55},         // log2(素数P) (辅助模数)
			PlaintextModulus: 0x10001,           // 明文模数T
		}); err != nil {
		panic(err)
	}

	fmt.Println("参数创建成功。")

	// 密钥生成器
	kgen := rlwe.NewKeyGenerator(params)

	// 私钥
	sk := kgen.GenSecretKeyNew()

	// 生成重线性化密钥
	rlk := kgen.GenRelinearizationKeyNew(sk)

	// 编码器
	ecd := bgv.NewEncoder(params)

	// 加密器
	enc := rlwe.NewEncryptor(params, sk)

	// 解密器
	dec := rlwe.NewDecryptor(params, sk)

	// 用于同态运算的评估器
	// 对于BFV，我们需要将scale-invariant标志设置为true
	eval := bgv.NewEvaluator(params, rlwe.NewMemEvaluationKeySet(rlk), true)

	// 明文模数
	T := params.PlaintextModulus()

	// 创建两个简单值：1和1
	values1 := make([]uint64, params.MaxSlots())
	values2 := make([]uint64, params.MaxSlots())

	// 将所有槽位设置为1
	for i := range values1 {
		values1[i] = 100
		values2[i] = 200
	}

	fmt.Printf("明文模数T: %d\n", T)
	fmt.Printf("向量1的前几个值: ")
	for i := 0; i < 4; i++ {
		fmt.Printf("%d ", values1[i])
	}
	fmt.Printf("\n")

	fmt.Printf("向量2的前几个值: ")
	for i := 0; i < 4; i++ {
		fmt.Printf("%d ", values2[i])
	}
	fmt.Printf("\n")

	// 编码向量
	pt1 := bgv.NewPlaintext(params, params.MaxLevel())
	if err = ecd.Encode(values1, pt1); err != nil {
		panic(err)
	}

	pt2 := bgv.NewPlaintext(params, params.MaxLevel())
	if err = ecd.Encode(values2, pt2); err != nil {
		panic(err)
	}

	// 加密明文
	ct1, err := enc.EncryptNew(pt1)
	if err != nil {
		panic(err)
	}

	ct2, err := enc.EncryptNew(pt2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("密文创建成功。\n")

	// 执行同态加法：ct1 + ct2
	ctAdd := bgv.NewCiphertext(params, 1, params.MaxLevel())
	if err = eval.Add(ct1, ct2, ctAdd); err != nil {
		panic(err)
	}

	// 执行同态乘法：ct1 * ct2
	// 注意：对于BFV，我们在scale-invariant模式下使用Mul
	ctMul := bgv.NewCiphertext(params, 2, params.MaxLevel())
	if err = eval.Mul(ct1, ct2, ctMul); err != nil {
		panic(err)
	}

	// 重线性化结果将密文次数从2降到1
	if err = eval.Relinearize(ctMul, ctMul); err != nil {
		panic(err)
	}

	fmt.Printf("同态运算完成。\n")

	// 解密、解码并打印结果
	PrintPrecisionStats("加法", params, ctAdd, sk, ecd, dec, func(i int) uint64 {
		return (values1[i] + values2[i]) % T
	})

	PrintPrecisionStats("乘法", params, ctMul, sk, ecd, dec, func(i int) uint64 {
		return (values1[i] * values2[i]) % T
	})
}

// PrintPrecisionStats 解密、解码并打印密文的精度统计信息。
func PrintPrecisionStats(op string, params bgv.Parameters, ct *rlwe.Ciphertext, sk *rlwe.SecretKey, ecd *bgv.Encoder, dec *rlwe.Decryptor, expected func(int) uint64) {
	// 解密结果
	pt := dec.DecryptNew(ct)

	// 解码明文
	have := make([]uint64, params.MaxSlots())
	if err := ecd.Decode(pt, have); err != nil {
		panic(err)
	}

	// 计算期望结果
	want := make([]uint64, params.MaxSlots())
	for i := range want {
		want[i] = expected(i)
	}

	// 美观地打印一些值
	fmt.Printf("\n%s 结果:\n", op)
	fmt.Printf("实际: ")
	for i := 0; i < 4; i++ {
		fmt.Printf("%d ", have[i])
	}
	fmt.Printf("...\n")

	fmt.Printf("期望: ")
	for i := 0; i < 4; i++ {
		fmt.Printf("%d ", want[i])
	}
	fmt.Printf("...\n")

	// 验证正确性
	correct := true
	for i := range want {
		if want[i] != have[i] {
			correct = false
			break
		}
	}

	if correct {
		fmt.Printf("运算 %s: 正确\n", op)
	} else {
		fmt.Printf("运算 %s: 错误\n", op)
	}
}