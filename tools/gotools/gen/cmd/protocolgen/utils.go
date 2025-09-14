package protocolgen

import "unicode"

// CapitalizeFirst 更完善的字符串首字母大写函数
func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	// 转换为rune切片处理Unicode
	runes := []rune(s)
	first := runes[0]

	// 如果首字符已经是大写字母，直接返回原字符串
	if unicode.IsUpper(first) {
		return s
	}

	// 如果首字符不是字母，直接返回原字符串
	if !unicode.IsLetter(first) {
		return s
	}

	// 首字母转为大写，拼接剩余部分
	runes[0] = unicode.ToUpper(first)
	return string(runes)
}
