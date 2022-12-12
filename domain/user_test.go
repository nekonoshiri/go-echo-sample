package domain

import "testing"

// 新しいユーザーを作成するテスト。
func TestNewUser(t *testing.T) {
	user := NewUser("name")

	anotherUser := NewUser("name")
	if user.UserID == anotherUser.UserID {
		t.Errorf("ユーザーの ID は一意のはずですが、新しく作成した２ユーザーの ID が等しくなっています。"+
			"１人目のユーザーの ID: %q, ２人目のユーザーの ID: %q", user.UserID, anotherUser.UserID)
	}

	if user.Name != "name" {
		t.Errorf(`新しいユーザーの名前に "name" を指定しましたが、設定された名前は %q です。`, user.Name)
	}

	if user.Status != UserStatusNormal {
		t.Errorf("新しいユーザーのステータスは %q のはずですが、設定されたステータスは %q です。", UserStatusNormal, user.Status)
	}
}

// ユーザーの凍結・凍結解除のテスト。
func TestFreezeUnfreeze(t *testing.T) {
	user := NewUser("")

	user.Freeze()
	if !user.IsFrozen() {
		t.Errorf("Freeze メソッドを実行しましたが、IsFrozen メソッドが false を返しました。")
	}

	user.Unfreeze()
	if user.IsFrozen() {
		t.Errorf("Unfreeze メソッドを実行しましたが、IsFrozen メソッドが true を返しました。")
	}
}

// ユーザーの名前変更のテスト。
func TestChangeName(t *testing.T) {
	t.Run("通常状態のユーザーの名前が変更できることのテスト。", func(t *testing.T) {
		user := NewUser("oldname")
		user.Unfreeze()

		err := user.ChangeName("newname")
		if err != nil {
			t.Fatalf(`ユーザーの名前を "oldname" から "newname" に変更しようとしましたが、エラーが発生しました: %v`, err)
		}

		if user.Name != "newname" {
			t.Errorf(`ユーザーの名前を "oldname" から "newname" に変更しようとしましたが、変更後の名前は %q となっています。`, user.Name)
		}
	})

	t.Run("凍結状態のユーザーの名前は変更できないことのテスト。", func(t *testing.T) {
		user := NewUser("oldname")
		user.Freeze()

		err := user.ChangeName("newname")
		if err == nil {
			t.Fatalf("凍結状態のユーザーの名前は変更できないはずですが、ChangeName メソッドがエラーを返しませんでした。")
		}

		if user.Name != "oldname" {
			t.Errorf("凍結状態のユーザーの名前は変更できないはずですが、%q に変更されています。", user.Name)
		}
	})
}
