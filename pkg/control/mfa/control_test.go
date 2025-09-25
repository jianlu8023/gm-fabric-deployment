package mfa

// func TestNewMFAControl(t *testing.T) {
// 	// 创建日志控制器
// 	logCtrl := logger.NewLoggerControl(nil)
// 	logCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动日志控制器失败: %v", err)
// 		}
// 	})
// 	defer logCtrl.Shutdown()
//
// 	// 创建配置控制器
// 	cfgCtrl := config.NewConfigControl()
// 	cfgCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动配置控制器失败: %v", err)
// 		}
// 	})
// 	defer cfgCtrl.Shutdown()
//
// 	// 创建MFA控制器
// 	mfaCtrl := NewMFAControl(cfgCtrl, logCtrl)
// 	if mfaCtrl == nil {
// 		t.Fatal("创建MFA控制器失败")
// 	}
//
// 	// 启动MFA控制器
// 	mfaCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动MFA控制器失败: %v", err)
// 		}
// 	})
// 	defer mfaCtrl.Shutdown()
//
// 	t.Log("MFA控制器创建和启动成功")
// }
//
// func TestMFAControl_GenerateSecret(t *testing.T) {
// 	// 创建日志控制器
// 	logCtrl := logger.NewLoggerControl(config.GetDefaultLoggerConfig())
// 	logCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动日志控制器失败: %v", err)
// 		}
// 	})
// 	defer logCtrl.Shutdown()
//
// 	// 创建配置控制器
// 	cfgCtrl := config.NewConfigControl()
// 	cfgCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动配置控制器失败: %v", err)
// 		}
// 	})
// 	defer cfgCtrl.Shutdown()
//
// 	// 创建MFA控制器
// 	mfaCtrl := NewMFAControl(cfgCtrl, logCtrl)
// 	mfaCtrl.StartUp(func(err error) {
// 		if err != nil {
// 			t.Fatalf("启动MFA控制器失败: %v", err)
// 		}
// 	})
// 	defer mfaCtrl.Shutdown()
//
// 	// 生成Google MFA密钥
// 	settings, err := mfaCtrl.GenerateSecret(context.Background(), "google", "test@example.com")
// 	if err != nil {
// 		t.Fatalf("生成Google MFA密钥失败: %v", err)
// 	}
// 	if settings.Secret == "" {
// 		t.Fatal("生成的Google MFA密钥为空")
// 	}
// 	if settings.QRCodeURL == "" {
// 		t.Fatal("生成的Google MFA QR码URL为空")
// 	}
//
// 	t.Logf("Google MFA密钥生成成功: %s", settings.Secret)
// 	t.Logf("Google MFA QR码URL: %s", settings.QRCodeURL)
// }
//
// func TestMFAControl_VerifyCode(t *testing.T) {
// 	// 这里仅测试验证逻辑，实际测试需要使用有效的密钥和验证码
// 	t.Skip("跳过验证测试，需要实际的密钥和验证码")
// }
