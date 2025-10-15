package fhe

import (
	"fmt"
	"log"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/multiparty"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
	"github.com/tuneinsight/lattigo/v6/utils/sampling"
)

// party 表示多方计算中的一个参与方。
type party struct {
	multiparty.Combiner

	i        int             // 参与方索引
	sk       *rlwe.SecretKey // 参与方的私钥
	tsk      multiparty.ShamirSecretShare
	shamirPt multiparty.ShamirPublicPoint
	input    []uint64 // 参与方的输入值
}

func exampleMultiBfv(N, t int) {
	fmt.Printf("开始多方BFV示例，%d个参与方计算加法和乘法...\n", N)

	// 设置BFV参数（使用带scale-invariant标志的BGV）
	// 128位安全参数，支持深度为3的电路。
	// LogN:13, LogQP: 218。
	params, err := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
		LogN:             13,                // log2(环度)
		LogQ:             []int{55, 54, 54}, // log2(素数Q) (密文模数)
		LogP:             []int{55},         // log2(素数P) (辅助模数)
		PlaintextModulus: 0x10001,           // 明文模数T
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("参数创建成功。\n")
	fmt.Printf("参与方数量: %d\n", N)
	fmt.Printf("阈值: %d\n", t)

	// 创建N个参与方并生成它们的密钥和输入
	P := genParties(params, N, t)
	fmt.Printf("参与方生成成功。\n")

	// 步骤1：阈值密钥生成
	fmt.Println("\n========= 阈值密钥生成 =========")
	thresholdizer := multiparty.NewThresholdizer(params.Parameters)

	// 为协议中各参与方的份额分配内存
	tskShares := make([][]multiparty.ShamirSecretShare, N)
	for i := range P {
		tskShares[i] = make([]multiparty.ShamirSecretShare, N)
		for j := range P {
			tskShares[i][j] = thresholdizer.AllocateThresholdSecretShare()
		}
	}

	// 各参与方为其阈值密钥生成份额
	for i, pi := range P {
		pol, err := thresholdizer.GenShamirPolynomial(t, pi.sk)
		if err != nil {
			log.Fatal(err)
		}
		for j, pj := range P {
			thresholdizer.GenShamirSecretShare(pj.shamirPt, pol, &tskShares[i][j])
		}
	}

	// 每个参与方聚合它从其他参与方收到的份额
	for i := range P {
		P[i].tsk = thresholdizer.AllocateThresholdSecretShare()
		for j := range P {
			thresholdizer.AggregateShares(tskShares[j][i], P[i].tsk, &P[i].tsk)
		}
	}

	fmt.Println("阈值密钥生成完成。")

	// 步骤2：集体公钥生成
	fmt.Println("\n========= 集体公钥生成 =========")
	crs, err := sampling.NewKeyedPRNG([]byte{'g', 'o', 'l', 'a', 'n', 'g'})
	if err != nil {
		log.Fatal(err)
	}

	pk := execCKGProtocol(params, crs, P[:t]) // 仅使用t个参与方进行协议
	fmt.Println("集体公钥生成完成。")

	// 步骤3：集体重线性化密钥生成
	fmt.Println("\n========= 集体重线性化密钥生成 =========")
	rlk := execRKGProtocol(params, crs, P[:t]) // 仅使用t个参与方进行协议
	fmt.Println("集体重线性化密钥生成完成。")

	// 步骤4：输入加密
	fmt.Println("\n========= 输入加密 =========")
	encoder := bgv.NewEncoder(params)
	encryptor := rlwe.NewEncryptor(params, pk)

	// 每个参与方加密其输入（设置为参与方索引+1）
	encInputs := make([]*rlwe.Ciphertext, N)
	for i, pi := range P {
		pt := bgv.NewPlaintext(params, params.MaxLevel())
		if err := encoder.Encode(pi.input, pt); err != nil {
			log.Fatal(err)
		}
		encInputs[i], err = encryptor.EncryptNew(pt)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("参与方 %d 加密输入: %d\n", i, pi.input[0])
	}

	// 步骤5：同态计算（所有参与方输入值的加法和乘法）
	fmt.Println("\n========= 同态计算 =========")
	// 创建同态运算评估器（BFV模式，scale-invariant=true）
	eval := bgv.NewEvaluator(params, rlwe.NewMemEvaluationKeySet(rlk))

	// 计算加法：所有参与方输入值相加
	fmt.Printf("计算加法: ")
	addResult := bgv.NewCiphertext(params, 1, params.MaxLevel())
	if err := eval.Add(encInputs[0], encInputs[1], addResult); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d + %d", P[0].input[0], P[1].input[0])

	for i := 2; i < N; i++ {
		if err := eval.Add(addResult, encInputs[i], addResult); err != nil {
			log.Fatal(err)
		}
		fmt.Printf(" + %d", P[i].input[0])
	}
	fmt.Println("...")
	fmt.Println("同态加法完成。")

	// 计算乘法：所有参与方输入值相乘
	fmt.Printf("计算乘法: ")
	// 第一次乘法
	mulResult := bgv.NewCiphertext(params, 2, params.MaxLevel()) // 重线性化前次数为2
	if err := eval.Mul(encInputs[0], encInputs[1], mulResult); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d * %d", P[0].input[0], P[1].input[0])

	// 重线性化将次数从2降到1
	if err := eval.Relinearize(mulResult, mulResult); err != nil {
		log.Fatal(err)
	}

	// 继续与其他参与方的输入相乘
	for i := 2; i < N; i++ {
		tempResult := bgv.NewCiphertext(params, 2, params.MaxLevel()) // 重线性化前次数为2
		if err := eval.Mul(mulResult, encInputs[i], tempResult); err != nil {
			log.Fatal(err)
		}
		fmt.Printf(" * %d", P[i].input[0])

		// 重线性化将次数从2降到1
		if err := eval.Relinearize(tempResult, mulResult); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("...")
	fmt.Println("同态乘法完成。")

	// 步骤6：集体解密
	fmt.Println("\n========= 集体解密 =========")
	// 使用t个参与方进行解密
	onlineParties := P[:t]

	// 执行集体密钥切换以解密（朝向实际私钥解密）
	encOutAdd := execCKSProtocol(params, onlineParties, addResult)
	encOutMul := execCKSProtocol(params, onlineParties, mulResult)

	// 解密加法结果
	// 为了解密，我们需要组合所有t个参与方的份额来获得理想密钥
	idealSk := rlwe.NewSecretKey(params)
	for i := 0; i < t; i++ {
		params.RingQP().Add(idealSk.Value, onlineParties[i].sk.Value, idealSk.Value)
	}

	decryptor := rlwe.NewDecryptor(params, idealSk)
	ptResAdd := bgv.NewPlaintext(params, params.MaxLevel())
	decryptor.Decrypt(encOutAdd, ptResAdd)

	// 解码并打印加法结果
	resAdd := make([]uint64, params.MaxSlots())
	if err := encoder.Decode(ptResAdd, resAdd); err != nil {
		log.Fatal(err)
	}

	// 计算期望的加法结果
	expectedAdd := uint64(0)
	for i := 0; i < N; i++ {
		expectedAdd += P[i].input[0]
	}

	fmt.Printf("解密的加法结果: %d\n", resAdd[0])
	fmt.Printf("期望的加法结果: %d\n", expectedAdd)

	// 验证加法正确性
	if resAdd[0] == expectedAdd {
		fmt.Println("多方BFV加法计算: 正确")
	} else {
		fmt.Println("多方BFV加法计算: 错误")
	}

	// 解密乘法结果
	decryptorMul := rlwe.NewDecryptor(params, idealSk)
	ptResMul := bgv.NewPlaintext(params, params.MaxLevel())
	decryptorMul.Decrypt(encOutMul, ptResMul)

	// 解码并打印乘法结果
	resMul := make([]uint64, params.MaxSlots())
	if err := encoder.Decode(ptResMul, resMul); err != nil {
		log.Fatal(err)
	}

	// 计算期望的乘法结果
	expectedMul := uint64(1)
	for i := 0; i < N; i++ {
		expectedMul *= P[i].input[0]
	}

	fmt.Printf("解密的乘法结果: %d\n", resMul[0])
	fmt.Printf("期望的乘法结果: %d\n", expectedMul)

	// 验证乘法正确性
	if resMul[0] == expectedMul {
		fmt.Println("多方BFV乘法计算: 正确")
	} else {
		fmt.Println("多方BFV乘法计算: 错误")
	}
}

// genParties 创建N个参与方及其密钥和输入。
func genParties(params bgv.Parameters, N, t int) []party {
	P := make([]party, N) // 明确指定长度
	kgen := rlwe.NewKeyGenerator(params)

	for i := range P {
		/* #nosec G115 -- i cannot be negative */
		P[i].shamirPt = multiparty.ShamirPublicPoint(i + 1)
		P[i].i = i
		P[i].sk = kgen.GenSecretKeyNew()

		// 将参与方的输入设置为参与方索引+1（这样每个参与方有不同的输入）
		P[i].input = make([]uint64, params.MaxSlots())
		for j := range P[i].input {
			P[i].input[j] = uint64(i + 1)
		}
	}

	shamirPts := getShamirPoints(P)
	for i := range P {
		P[i].Combiner = multiparty.NewCombiner(params.Parameters, P[i].shamirPt, shamirPts, t)
	}

	return P
}

// getShamirPoints 返回一组参与方的Shamir公共点。
func getShamirPoints(parties []party) []multiparty.ShamirPublicPoint {
	shamirPoints := make([]multiparty.ShamirPublicPoint, len(parties))
	for i := range parties {
		shamirPoints[i] = parties[i].shamirPt
	}
	return shamirPoints
}

// execCKGProtocol 执行集体公钥生成协议。
func execCKGProtocol(params bgv.Parameters, crs sampling.PRNG, participants []party) *rlwe.PublicKey {
	ckg := multiparty.NewPublicKeyGenProtocol(params)

	// 为协议中的参与方份额分配内存
	ckgShares := make([]multiparty.PublicKeyGenShare, len(participants))
	tsks := make([]*rlwe.SecretKey, len(participants))
	for i := range ckgShares {
		ckgShares[i] = ckg.AllocateShare()  // 公共CKG份额
		tsks[i] = rlwe.NewSecretKey(params) // 参与方组内的t-out-of-t私钥
	}
	ckgCombined := ckg.AllocateShare() // 为组合份额分配内存

	// 从公共参考字符串(crs)中采样公共参考多项式(crp)
	crp := ckg.SampleCRP(crs)

	// 生成参与方的份额
	for i, pi := range participants {
		// 在参与方组内生成参与方的t-out-of-t私钥
		err := pi.Combiner.GenAdditiveShare(getShamirPoints(participants), pi.shamirPt, pi.tsk, tsks[i])
		if err != nil {
			log.Fatal(err)
		}

		// 从t-out-of-t私钥生成公钥份额
		ckg.GenShare(tsks[i], crp, &ckgShares[i])
	}

	// 将参与方的份额聚合成集体公钥
	pk := rlwe.NewPublicKey(params)
	// 将参与方的份额聚合成组合份额
	for i := range participants {
		ckg.AggregateShares(ckgShares[i], ckgCombined, &ckgCombined)
	}

	// 从组合份额生成公钥
	ckg.GenPublicKey(ckgCombined, crp, pk)

	return pk
}

// execRKGProtocol 执行集体重线性化密钥生成协议。
func execRKGProtocol(params bgv.Parameters, crs sampling.PRNG, participants []party) *rlwe.RelinearizationKey {
	// 创建集体重线性化密钥生成的协议类型。
	rkg := multiparty.NewRelinearizationKeyGenProtocol(params)

	// 为协议中的参与方份额分配内存
	rkgSharesRoundOne := make([]multiparty.RelinearizationKeyGenShare, len(participants))
	rkgSharesRoundTwo := make([]multiparty.RelinearizationKeyGenShare, len(participants))

	// RKG协议中每个参与方的私有临时私钥
	rlkEphemSk := make([]*rlwe.SecretKey, len(participants))

	tsks := make([]*rlwe.SecretKey, len(participants))
	for i := range participants {
		// 参与方在RKGen协议中有一个私有临时私钥
		rlkEphemSk[i], rkgSharesRoundOne[i], rkgSharesRoundTwo[i] = rkg.AllocateShare()
		tsks[i] = rlwe.NewSecretKey(params)
	}

	// 为组合的公共份额分配内存
	_, rkgCombined1, rkgCombined2 := rkg.AllocateShare()

	// 从公共参考字符串(crs)中采样公共参考多项式(crp)
	crp := rkg.SampleCRP(crs)

	// 参与方生成第一轮的份额
	for i, pi := range participants {
		// 在参与方组内生成参与方的t-out-of-t私钥
		err := pi.Combiner.GenAdditiveShare(getShamirPoints(participants), pi.shamirPt, pi.tsk, tsks[i])
		if err != nil {
			log.Fatal(err)
		}

		// 从t-out-of-t私钥生成第一轮的份额
		rkg.GenShareRoundOne(tsks[i], crp, rlkEphemSk[i], &rkgSharesRoundOne[i])
	}

	// 助手聚合参与方第一轮的份额
	for i := range participants {
		rkg.AggregateShares(rkgSharesRoundOne[i], rkgCombined1, &rkgCombined1)
	}

	// 参与方生成第二轮的份额
	for i := range participants {
		// 从t-out-of-t私钥生成第二轮的份额
		rkg.GenShareRoundTwo(rlkEphemSk[i], tsks[i], rkgCombined1, &rkgSharesRoundTwo[i])
	}

	// 助手聚合参与方第二轮的份额并生成重线性化密钥
	rlk := rlwe.NewRelinearizationKey(params)
	for i := range participants {
		rkg.AggregateShares(rkgSharesRoundTwo[i], rkgCombined2, &rkgCombined2)
	}
	rkg.GenRelinearizationKey(rkgCombined1, rkgCombined2, rlk)

	return rlk
}

// execCKSProtocol 执行用于解密的集体密钥切换协议。
func execCKSProtocol(params bgv.Parameters, participants []party, ctIn *rlwe.Ciphertext) *rlwe.Ciphertext {
	// 创建集体密钥切换协议的协议类型
	cks, err := multiparty.NewKeySwitchProtocol(params, params.Xe())
	if err != nil {
		log.Fatal(err)
	}

	// 为协议中的参与方份额分配内存
	cksShares := make([]multiparty.KeySwitchShare, len(participants))
	tsks := make([]*rlwe.SecretKey, len(participants))
	for i := range participants {
		cksShares[i] = cks.AllocateShare(params.MaxLevel()) // 为公共份额分配内存
		tsks[i] = rlwe.NewSecretKey(params)                 // 为t-out-of-t私钥分配内存
	}
	cksCombined := cks.AllocateShare(params.MaxLevel()) // 为组合份额分配内存

	// 参与方生成密钥切换协议的份额
	for i := range participants {
		// 生成t-out-of-t私钥
		err := participants[i].Combiner.GenAdditiveShare(getShamirPoints(participants), participants[i].shamirPt, participants[i].tsk, tsks[i])
		if err != nil {
			log.Fatal(err)
		}

		// 使用t-out-of-t私钥生成密钥切换份额
		cks.GenShare(tsks[i], participants[i].sk, ctIn, &cksShares[i])
	}

	// 助手聚合参与方的份额并生成密钥切换密钥
	ctOut := bgv.NewCiphertext(params, 1, params.MaxLevel())
	// 将参与方的份额聚合成组合份额
	for i := range participants {
		err := cks.AggregateShares(cksShares[i], cksCombined, &cksCombined)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 从组合份额生成重加密
	cks.KeySwitch(ctIn, cksCombined, ctOut)

	return ctOut
}
