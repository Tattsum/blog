package user

// Email はメールアドレス（値オブジェクト）。
type Email string

// String はメールアドレスの文字列表現を返す。
func (e Email) String() string { return string(e) }

// User は管理者ユーザのドメインエンティティ。
type User struct {
	ID          string
	Email       Email
	DisplayName string
}
