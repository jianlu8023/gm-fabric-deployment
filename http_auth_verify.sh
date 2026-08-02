#!/bin/bash
# =====================================================================================
# HTTP 认证流程验证脚本
# 目标：验证 jwt + session + auth 三层中间件重构后的认证流程
# 前提：server 已启动（./server.bin -c configs/server.yaml -t win），监听 8080 端口
# 使用：bash http_auth_verify.txt
# =====================================================================================

set -u

# 基础变量
BASE_URL="https://localhost:8080/example"
CURL_OPTS="-sk"  # -s 静默 -k 跳过自签证书校验
USERNAME="testuser"
PASSWORD="test123"
EMAIL="test@example.com"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
sep()  { echo "──────────────────────────────────────────────────────────────"; }

sep
echo "HTTP 认证流程验证"
echo "目标服务: $BASE_URL"
echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
sep

# --------------------------------------------------------------------------------------
# 1. 注册用户（若已存在会返回业务错误，不影响后续流程）
# --------------------------------------------------------------------------------------
echo ""
info "1. 注册用户 ($USERNAME)"
REGISTER_RESP=$(curl $CURL_OPTS -X POST "$BASE_URL/user/register" \
  -F "username=$USERNAME" \
  -F "password=$PASSWORD" \
  -F "email=$EMAIL")
echo "响应: $REGISTER_RESP"
REGISTER_SUCCESS=$(echo "$REGISTER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success', False))" 2>/dev/null)
if [ "$REGISTER_SUCCESS" = "True" ]; then
    pass "注册成功"
else
    info "注册返回非成功（用户可能已存在，属正常，继续后续测试）"
fi

# --------------------------------------------------------------------------------------
# 2. 登录获取 token
# --------------------------------------------------------------------------------------
echo ""
info "2. 登录获取 token"
LOGIN_RESP=$(curl $CURL_OPTS -X POST "$BASE_URL/user/login" \
  -F "username=$USERNAME" \
  -F "password=$PASSWORD")
echo "响应: $LOGIN_RESP"
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('token',''))" 2>/dev/null)
LOGIN_SUCCESS=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('success', False))" 2>/dev/null)
if [ "$LOGIN_SUCCESS" = "True" ] && [ -n "$TOKEN" ]; then
    pass "登录成功，token 已获取"
    info "token 前 60 字符: ${TOKEN:0:60}..."
else
    fail "登录失败，无法继续后续测试"
    exit 1
fi

# --------------------------------------------------------------------------------------
# 3. 认证中间件场景测试（受保护接口：GET /ws/connections，EnableJWtVerify=true）
# --------------------------------------------------------------------------------------
PROTECTED_URL="$BASE_URL/ws/connections"

# 3a. 不带 token
echo ""
info "3a. 不带 token 访问受保护接口 (期望 401 单一响应)"
RESP_3A=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL")
HTTP_3A=$(echo "$RESP_3A" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3A=$(echo "$RESP_3A" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3A"
echo "HTTP: $HTTP_3A"
JSON_COUNT_3A=$(echo "$BODY_3A" | grep -o '{' | wc -l)
if [ "$HTTP_3A" = "401" ] && [ "$JSON_COUNT_3A" = "1" ]; then
    pass "3a 通过：401 + 单一 JSON 响应"
else
    fail "3a 失败：期望 401 且单一 JSON，实际 HTTP=$HTTP_3A JSON数=$JSON_COUNT_3A"
fi

# 3b. 带正确 token
echo ""
info "3b. 带正确 token 访问受保护接口 (期望 200，认证通过)"
RESP_3B=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL" -H "Authorization: Bearer $TOKEN")
HTTP_3B=$(echo "$RESP_3B" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3B=$(echo "$RESP_3B" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3B"
echo "HTTP: $HTTP_3B"
if [ "$HTTP_3B" = "200" ]; then
    pass "3b 通过：200 认证通过（业务层报分页参数错误属正常，认证已放行）"
else
    fail "3b 失败：期望 200，实际 HTTP=$HTTP_3B"
fi

# 3c. 带错误 token
echo ""
info "3c. 带错误 token 访问受保护接口 (期望 401 认证信息无效或已过期)"
RESP_3C=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL" -H "Authorization: Bearer invalidtoken123")
HTTP_3C=$(echo "$RESP_3C" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3C=$(echo "$RESP_3C" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3C"
echo "HTTP: $HTTP_3C"
if [ "$HTTP_3C" = "401" ]; then
    pass "3c 通过：401 token 无效被拒"
else
    fail "3c 失败：期望 401，实际 HTTP=$HTTP_3C"
fi

# 3d. 格式错误 token（验证重复响应 bug 已修复）
echo ""
info "3d. 格式错误 token 访问 (期望 401 单一响应，验证重复响应 bug 已修复)"
RESP_3D=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL" -H "Authorization: InvalidFormat")
HTTP_3D=$(echo "$RESP_3D" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3D=$(echo "$RESP_3D" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3D"
echo "HTTP: $HTTP_3D"
JSON_COUNT_3D=$(echo "$BODY_3D" | grep -o '{' | wc -l)
if [ "$HTTP_3D" = "401" ] && [ "$JSON_COUNT_3D" = "1" ]; then
    pass "3d 通过：401 + 单一 JSON 响应（重复响应 bug 已修复）"
else
    fail "3d 失败：期望 401 且单一 JSON，实际 HTTP=$HTTP_3D JSON数=$JSON_COUNT_3D"
fi

# 3e. 缺少 Bearer 前缀（仅有 token 无前缀）
echo ""
info "3e. 仅 token 无 Bearer 前缀 (期望 401 格式错误)"
RESP_3E=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL" -H "Authorization: $TOKEN")
HTTP_3E=$(echo "$RESP_3E" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3E=$(echo "$RESP_3E" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3E"
echo "HTTP: $HTTP_3E"
if [ "$HTTP_3E" = "401" ]; then
    pass "3e 通过：401 缺少 Bearer 前缀被拒"
else
    fail "3e 失败：期望 401，实际 HTTP=$HTTP_3E"
fi

# 3f. Bearer 前缀但 token 为空
echo ""
info "3f. Bearer 前缀但 token 为空 (期望 401 格式错误)"
RESP_3F=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$PROTECTED_URL" -H "Authorization: Bearer ")
HTTP_3F=$(echo "$RESP_3F" | grep -oP 'HTTP_CODE:\K\d+')
BODY_3F=$(echo "$RESP_3F" | sed 's/HTTP_CODE:.*//')
echo "响应: $BODY_3F"
echo "HTTP: $HTTP_3F"
if [ "$HTTP_3F" = "401" ]; then
    pass "3f 通过：401 空 token 被拒"
else
    fail "3f 失败：期望 401，实际 HTTP=$HTTP_3F"
fi

# --------------------------------------------------------------------------------------
# 4. 公开接口测试（无需认证，EnableJWtVerify=false）
# --------------------------------------------------------------------------------------
echo ""
info "4. 公开接口测试 (期望 200，无需 token)"
RESP_4=$(curl $CURL_OPTS -w "\nHTTP_CODE:%{http_code}" "$BASE_URL/ping")
HTTP_4=$(echo "$RESP_4" | grep -oP 'HTTP_CODE:\K\d+')
echo "HTTP: $HTTP_4"
if [ "$HTTP_4" = "200" ]; then
    pass "4 通过：公开接口无需认证可访问"
else
    fail "4 失败：期望 200，实际 HTTP=$HTTP_4"
fi

# --------------------------------------------------------------------------------------
# 5. 连续请求验证滑动续期（同一 token 两次请求都应通过）
# --------------------------------------------------------------------------------------
echo ""
info "5. 连续请求验证 session 续期 (同一 token 两次请求都应 200)"
RESP_5A=$(curl $CURL_OPTS -o /dev/null -w "%{http_code}" "$PROTECTED_URL" -H "Authorization: Bearer $TOKEN")
RESP_5B=$(curl $CURL_OPTS -o /dev/null -w "%{http_code}" "$PROTECTED_URL" -H "Authorization: Bearer $TOKEN")
echo "第 1 次: $RESP_5A"
echo "第 2 次: $RESP_5B"
if [ "$RESP_5A" = "200" ] && [ "$RESP_5B" = "200" ]; then
    pass "5 通过：连续请求均通过，session 续期正常"
else
    fail "5 失败：第1次=$RESP_5A 第2次=$RESP_5B"
fi

# --------------------------------------------------------------------------------------
# 汇总
# --------------------------------------------------------------------------------------
echo ""
sep
echo "验证完成。请确认上方所有用例均为 [PASS]。"
echo "若需查看服务端日志：tail -f /tmp/server-test.log | grep -iE 'auth|session|jwt'"
sep

