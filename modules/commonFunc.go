package modules

func padLeft(str string, count int, c byte) string {
	if len(str) >= count {
		return str
	}
	tmp := []byte(str)
	l := len(tmp)
	buf := make([]byte, count)
	for i := 0; i < count-l; i++ {
		buf[i] = c
	}
	for i := 0; i < l; i++ {
		buf[count-l+i] = tmp[i]
	}
	return string(buf[:])
}

func padRight(str string, count int, c byte) string {
	if len(str) > count {
		return str
	}
	tmp := []byte(str)
	l := len(tmp)
	buf := make([]byte, count)
	for i := 0; i < l; i++ {
		buf[i] = tmp[i]
	}
	for i := count - l; i < count; i++ {
		buf[i] = c
	}
	return string(buf[:])
}
