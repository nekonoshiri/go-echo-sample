package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ユーザーのステータス。
type UserStatus string

const (
	UserStatusNormal UserStatus = "normal" // 通常。
	UserStatusFrozen UserStatus = "frozen" // 凍結状態。
)

// ユーザー。システムの利用者です。
type User struct {
	UserID       string     // ユーザー ID。
	Name         string     // 名前。
	Status       UserStatus // ステータス。
	RegisteredAt time.Time  // 登録日時 (UTC)。
}

// 新しいユーザーを作成します。
func NewUser(name string) User {
	return User{
		uuid.New().String(),
		name,
		UserStatusNormal,
		time.Now().UTC(),
	}
}

// テスト用に、有効な適当な値を持つユーザーを返します。
func DummyUser(*testing.T) User {
	return User{
		"",
		"",
		UserStatusNormal,
		time.Unix(0, 0).UTC(),
	}
}

// ユーザーが凍結状態であれば true、そうでなければ false を返します。
func (user *User) IsFrozen() bool {
	return user.Status == UserStatusFrozen
}

// ユーザーを凍結状態にします。ユーザーが既に凍結状態の場合は何もしません。
func (user *User) Freeze() {
	user.Status = UserStatusFrozen
}

// ユーザーの凍結状態を解除します。ユーザーが凍結状態でない場合は何もしません。
func (user *User) Unfreeze() {
	user.Status = UserStatusNormal
}

// ユーザーの名前を変更します。凍結状態のユーザーは名前を変更できません。
// 凍結状態のユーザーの名前を変更しようとした場合、エラーを返します。
// ユーザーが凍結状態でない場合は、エラーを返しません。
func (user *User) ChangeName(name string) error {
	if user.IsFrozen() {
		return errors.New("凍結状態のユーザーは名前を変更できません。")
	}

	user.Name = name
	return nil
}

// ユーザーのリポジトリ。
type UserRepository interface {
	// ユーザーを取得します。
	// ユーザーが見つからない場合は ErrUserNotFound を返します。
	Get(ctx context.Context, userID string) (*User, error)

	// ユーザー一覧を取得します。
	//
	// 初回の呼び出しでは exclusiveStartKey に空文字を指定してください。
	// 戻り値の lastEvaluatedKey が空文字列でない場合、
	// その値を exclusiveStartKey に指定して呼び出すと、続きから取得できます。
	//
	// もう続きがない場合、戻り値の lastEvaluatedKey は空文字列です。
	// このときその値（空文字列）を exclusiveStartKey に指定してしまうと再び最初から取得を始めてしまうため、
	// 呼び出しが無限に繰り返されないように、必ず lastEvaluatedKey が空文字列かどうかを確認してください。
	//
	// limit は最大取得件数です（実際に取得される件数は、この値未満になる可能性があります）。
	// limit に 0 または負数を指定すると、制限なし（最大取得件数が無限大）となります。この場合は、全件取得されます。
	List(ctx context.Context, exclusiveStartKey string, limit int) (users []User, lastEvaluatedKey string, err error)

	// ユーザーを保存します。
	Put(ctx context.Context, user *User) error

	// ユーザーを削除します。
	// この操作は冪等です。つまり、ユーザーが見つからない場合は何もしません（この場合、エラーは返しません）。
	Delete(ctx context.Context, userID string) error
}

var (
	// ErrUserNotFound は、ユーザーが見つからなかったことを表します。
	ErrUserNotFound = errors.New("ユーザーが見つかりません。")
)
