package models

// InitCache 는 서버 시작 때 한 번 불린다. 코드 테이블 등 메모리에 올려 둘 데이터가 있으면 여기서 읽는다.
func InitCache() {
	conn := GetConnection()
	if conn == nil {
		return
	}
	defer conn.Close()
}
