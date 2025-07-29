package weex

import (
	"math/rand/v2"
)

// GenerateRandomString 生成指定长度的随机字符串，并在特定位置插入固定字符
func GenerateRandomString(length int) string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.IntN(62)]
	}

	resultStr := string(result)
	resultStr = replaceCharAt(resultStr, 1, "5")  // 第1位（索引1）替换为 "5"
	resultStr = replaceCharAt(resultStr, 3, "7")  // 第3位（索引3）替换为 "7"
	resultStr = replaceCharAt(resultStr, 7, "8")  // 第7位（索引7）替换为 "8"
	resultStr = replaceCharAt(resultStr, 14, "9") // 第14位（索引14）替换为 "9"
	resultStr = replaceCharAt(resultStr, 20, "7") // 第20位（索引20）替换为 "7"
	resultStr = replaceCharAt(resultStr, 29, "6") // 第29位（索引29）替换为 "6"

	return resultStr
}

// replaceCharAt 在指定位置替换字符
func replaceCharAt(str string, index int, newChar string) string {
	if index < 0 || index >= len(str) {
		return str
	}
	return str[:index] + newChar + str[index+1:]
}
