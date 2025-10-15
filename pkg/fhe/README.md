# FHE模块说明

本目录包含基于BGV方案的同态加密实现，支持单方和多方计算场景。

## 文件说明

- [multibfv.go](file:///home/user/codes/go/src/github.com/jianlu8023/golang-example/pkg/fhe/multibfv.go) - 多方BFV同态加密实现
- [singlebfv.go](file:///home/user/codes/go/src/github.com/jianlu8023/golang-example/pkg/fhe/singlebfv.go) - 单方BFV同态加密实现
- [threshold.go](file:///home/user/codes/go/src/github.com/jianlu8023/golang-example/pkg/fhe/threshold.go) - 阈值方案分析和实现
- [parties_threshold.go](file:///home/user/codes/go/src/github.com/jianlu8023/golang-example/pkg/fhe/parties_threshold.go) - 多方参与的阈值方案示例

## 详细说明文档

- [多方BFV实现说明](./README_multibfv.md) - 多方同态加密的详细步骤
- [单方BFV实现说明](./README_singlebfv.md) - 单方同态加密的详细步骤
- [阈值方案说明](./README_threshold.md) - 阈值同态加密方案的详细说明

# 参数设置

```text
params, err := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
		LogN:             13,                // log2(环度)
		LogQ:             []int{55, 54, 54}, // log2(素数Q) (密文模数)
		LogP:             []int{55},         // log2(素数P) (辅助模数)
		PlaintextModulus: 0x10001,           // 明文模数T
})

LogN (环度): 这是多项式环的度数的对数，即N=2^LogN。它决定了可以并行处理的数据槽位数（slots）。
            推荐值：对于大多数应用，LogN=12到16是常见的选择
            LogN=13意味着有8192个槽位
LogQ (密文模数): 这是主要的密文模数，由一系列素数组成。每个素数代表一个"level"，决定了可以执行的计算深度。
                第一个素数通常较大（如55位），用于容纳初始噪声
                后续素数通常较小且大小相似（如54位）
                推荐值：素数大小通常在30-60位之间
LogP (辅助模数): 这是用于密钥切换操作的辅助模数。
                推荐值：通常选择50-61位的素数
                数量通常为sqrt(密文素数数量)
PlaintextModulus (明文模数T): 这是明文空间的模数，决定了可以表示的数值范围。
                0x10001 = 65537，这是一个常用的素数


LogN=14, LogQP≈438位: 这是标准的128位安全参数
LogN=16, LogQP≈1550位: 用于引导的参数，确保128位安全性
```