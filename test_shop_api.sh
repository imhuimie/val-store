#!/bin/bash

# 测试商店API接口
echo "测试商店API接口"
echo "================"

# 1. 普通请求测试
echo "1. 普通请求测试"
curl -X GET http://localhost:8080/api/shop -H "Authorization: Bearer test_token" -v

echo ""
echo "--------------------"
echo ""

# 2. 测试不同区域
echo "2. 测试不同区域"
for region in "ap" "na" "eu" "kr"; do
  echo "测试区域: $region"
  curl -X POST http://localhost:8080/api/user/region -H "Content-Type: application/json" -H "Authorization: Bearer test_token" -d "{\"region\":\"$region\"}" -v
  echo ""
  echo "获取商店数据:"
  curl -X GET http://localhost:8080/api/shop -H "Authorization: Bearer test_token" -v
  echo ""
  echo "--------------------"
  echo ""
done

# 3. 测试错误处理
echo "3. 测试错误处理 - 无授权"
curl -X GET http://localhost:8080/api/shop -v

echo ""
echo "测试完成" 