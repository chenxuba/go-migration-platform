/**
 * 计算年龄，返回格式化的年龄字符串
 * @param birthDay - 生日日期字符串 (YYYY-MM-DD)
 * @returns 格式化的年龄字符串 (例如: "24个月" 或 "3周岁")
 */
export function calculateAge(birthDay: string): string {
  if (!birthDay)
    return '-'

  const birth = new Date(birthDay)
  const now = new Date()

  // 计算总月份差
  let months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())

  // 如果当前日期小于出生日期的日，则减去一个月
  if (now.getDate() < birth.getDate()) {
    months--
  }

  // 如果不满一个月，显示1个月
  if (months <= 0) {
    return '1个月'
  }

  // 如果超过35个月，显示周岁
  if (months > 35) {
    const years = Math.floor(months / 12)
    return `${years}周岁`
  }

  // 35个月及以下，显示具体月份
  return `${months}个月`
}
