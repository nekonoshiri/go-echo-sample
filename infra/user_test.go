//go:build !skipmongo

package infra

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"nekonoshiri/go-echo-sample/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Get と Put のテスト。
func TestGetAndPut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	repo := NewMongoUserRepository(client, client.Database(mongoDatabase+"-test").Collection(userCollection+"-"+t.Name()))
	if err := repo.collection.Drop(ctx); err != nil {
		t.Fatalf("テスト前にコレクション %q をドロップしようとしましたが、失敗しました: %v", repo.collection.Name(), err)
	}

	testCases := []domain.User{
		{
			UserID:       "U1",
			Name:         "ユーザー1",
			Status:       domain.UserStatusNormal,
			RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:       "U2",
			Name:         "ユーザー2",
			Status:       domain.UserStatusFrozen,
			RegisteredAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, user := range testCases {
		t.Run(fmt.Sprintf("ユーザーID=%q", user.UserID), func(t *testing.T) {
			// 保存前にユーザーを取得しようとすると domain.ErrUserNotFound エラーが発生することを確認
			_, err = repo.Get(ctx, user.UserID)
			if err == nil {
				t.Fatalf("保存前のユーザーを取得しようとしましたが、エラーが発生しませんでした")
			}
			if !errors.Is(err, domain.ErrUserNotFound) {
				t.Fatalf("保存前のユーザーを取得しようとしましたが、予期しないエラーが発生しました: %v", err)
			}

			// ユーザーを保存
			if err := repo.Put(ctx, &user); err != nil {
				t.Fatalf("ユーザーの保存に失敗しました: %v", err)
			}

			// 保存したユーザーが取得されることを確認
			gotUser, err := repo.Get(ctx, user.UserID)
			if err != nil {
				t.Fatalf("保存したユーザーを取得しようとしましたが、エラーが発生しました: %v", err)
			}
			if diff := cmp.Diff(user, *gotUser); diff != "" {
				t.Fatalf("保存したユーザー (-) と取得したユーザー (+) が一致しませんでした:\n%s", diff)
			}
		})
	}
}

// 同じ ID のユーザーを保存すると上書きされることのテスト。
func TestPutOverwritesExistingUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	repo := NewMongoUserRepository(client, client.Database(mongoDatabase+"-test").Collection(userCollection+"-"+t.Name()))
	if err := repo.collection.Drop(ctx); err != nil {
		t.Fatalf("テスト前にコレクション %q をドロップしようとしましたが、失敗しました: %v", repo.collection.Name(), err)
	}

	user := domain.User{
		UserID:       "U1",
		Name:         "ユーザーA",
		Status:       domain.UserStatusNormal,
		RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := repo.Put(ctx, &user); err != nil {
		t.Fatalf("ユーザーの保存に失敗しました: %v", err)
	}

	// 名前を変えて同じユーザーをもう一度保存
	user.ChangeName("ユーザーB")
	if err := repo.Put(ctx, &user); err != nil {
		t.Fatalf("ユーザーの保存に失敗しました: %v", err)
	}

	// ユーザーが１人しか取得されない（上書きされている）ことを確認
	gotUsers, _, err := repo.List(ctx, "", -1)
	if err != nil {
		t.Fatalf("ユーザーの一覧を取得しようとしましたが、エラーが発生しました: %v", err)
	}
	if len(gotUsers) == 0 {
		t.Fatalf("ユーザーの一覧を取得しようとしましたが、一人も取得できませんでした")
	}
	if len(gotUsers) != 1 {
		t.Fatalf("ユーザーは１人しか保存されていないはずなのに、ユーザーの一覧取得で以下のユーザーが取得されました: %+v", gotUsers)
	}

	// 取得されたユーザーの名前が、名前変更後に保存されたものである（上書きされている）ことを確認
	if gotUsers[0].Name != "ユーザーB" {
		t.Fatalf("ユーザーの名前を %q にして上書き保存したはずなのに、取得されたユーザーの名前は %q でした", "ユーザーB", gotUsers[0].Name)
	}
}

// List のテスト。
func TestList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	repo := NewMongoUserRepository(client, client.Database(mongoDatabase+"-test").Collection(userCollection+"-"+t.Name()))
	if err := repo.collection.Drop(ctx); err != nil {
		t.Fatalf("テスト前にコレクション %q をドロップしようとしましたが、失敗しました: %v", repo.collection.Name(), err)
	}

	t.Run("ユーザーが１人も保存されていない場合のテスト", func(t *testing.T) {
		gotUsers, lastEvaluatedKey, err := repo.List(ctx, "", -1)
		if err != nil {
			t.Fatalf("ユーザーの一覧を取得しようとしましたが、エラーが発生しました: %v", err)
		}
		if len(gotUsers) != 0 {
			t.Fatalf("ユーザーが１人も保存されていないはずなのに、ユーザーの一覧取得で以下のユーザーが取得されました: %+v", gotUsers)
		}
		if lastEvaluatedKey != "" {
			t.Fatalf("ユーザーが１人も保存されていないはずなのに、ユーザーの一覧取得で空でない lastEvaluatedKey (%q) が返却されました", lastEvaluatedKey)
		}
	})

	// ユーザーを 100 人保存
	users := []domain.User{}
	for i := 0; i < 100; i++ {
		user := domain.User{
			UserID:       fmt.Sprintf("U%d", i),
			Name:         "ユーザー",
			Status:       domain.UserStatusNormal,
			RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
		}
		users = append(users, user)
		if err := repo.Put(ctx, &user); err != nil {
			t.Fatalf("ユーザーの保存に失敗しました: %v", err)
		}
	}

	t.Run("ユーザーを全件取得するテスト", func(t *testing.T) {
		testCases := []struct {
			limit int
		}{
			{limit: 0},
			{limit: -1},
			{limit: -2},
		}
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("limit=%d", tc.limit), func(t *testing.T) {
				gotUsers, lastEvaluatedKey, err := repo.List(ctx, "", tc.limit)
				if err != nil {
					t.Fatalf("ユーザーの一覧を取得しようとしましたが、エラーが発生しました: %v", err)
				}
				// スライスをソートしてから比較
				lessFunc := func(x, y domain.User) bool { return x.UserID < y.UserID }
				if diff := cmp.Diff(users, gotUsers, cmpopts.SortSlices(lessFunc)); diff != "" {
					t.Fatalf("保存したユーザー群 (-) と取得したユーザー群 (+) が一致しませんでした:\n%s", diff)
				}
				if lastEvaluatedKey != "" {
					t.Fatalf("ユーザーを全件取得したはずなのに、空文字列でない lastEvaluatedKey (%q) が返却されました", lastEvaluatedKey)
				}
			})
		}
	})

	t.Run("ページネーションのテスト", func(t *testing.T) {
		gotAllUsers := []domain.User{}
		exclusiveStartKey := ""

		for {
			gotUsers, lastEvaluatedKey, err := repo.List(ctx, exclusiveStartKey, 10)
			if err != nil {
				t.Fatalf("ユーザーの一覧を取得しようとしましたが、エラーが発生しました: %v", err)
			}
			if len(gotUsers) > 10 {
				t.Fatalf("ユーザーを最大 10 件取得しようとしましたが、%d 件取得されました", len(gotUsers))
			}
			gotAllUsers = append(gotAllUsers, gotUsers...)

			if lastEvaluatedKey == "" {
				break
			} else {
				exclusiveStartKey = lastEvaluatedKey
			}
		}

		// スライスをソートしてから比較
		lessFunc := func(x, y domain.User) bool { return x.UserID < y.UserID }
		if diff := cmp.Diff(users, gotAllUsers, cmpopts.SortSlices(lessFunc)); diff != "" {
			t.Fatalf("保存したユーザー群 (-) と取得したユーザー群 (+) が一致しませんでした:\n%s", diff)
		}
	})
}

// Delete のテスト。
func TestDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	repo := NewMongoUserRepository(client, client.Database(mongoDatabase+"-test").Collection(userCollection+"-"+t.Name()))
	if err := repo.collection.Drop(ctx); err != nil {
		t.Fatalf("テスト前にコレクション %q をドロップしようとしましたが、失敗しました: %v", repo.collection.Name(), err)
	}

	user := domain.User{
		UserID:       "U1",
		Name:         "ユーザー1",
		Status:       domain.UserStatusNormal,
		RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// 保存済みのユーザーを用意
	if err := repo.Put(ctx, &user); err != nil {
		t.Fatalf("ユーザーの保存に失敗しました: %v", err)
	}
	if _, err = repo.Get(ctx, user.UserID); err != nil {
		t.Fatalf("保存したユーザーを取得しようとしましたが、エラーが発生しました: %v", err)
	}

	// ユーザーを削除
	if err := repo.Delete(ctx, user.UserID); err != nil {
		t.Fatalf("ユーザーの削除に失敗しました: %v", err)
	}

	// ユーザーが削除されたことを確認
	_, err = repo.Get(ctx, user.UserID)
	if err == nil {
		t.Fatalf("削除済みのユーザーを取得しようとしましたが、エラーが発生しませんでした")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("削除済みのユーザーを取得しようとしましたが、予期しないエラーが発生しました: %v", err)
	}
}

// Delete が冪等である、つまりすでに削除されているユーザーを削除しようとしても何もしないことのテスト。
func TestDeleteIdempotency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	repo := NewMongoUserRepository(client, client.Database(mongoDatabase+"-test").Collection(userCollection+"-"+t.Name()))
	if err := repo.collection.Drop(ctx); err != nil {
		t.Fatalf("テスト前にコレクション %q をドロップしようとしましたが、失敗しました: %v", repo.collection.Name(), err)
	}

	user := domain.User{
		UserID:       "U1",
		Name:         "ユーザー1",
		Status:       domain.UserStatusNormal,
		RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// 保存済みのユーザーを用意
	if err := repo.Put(ctx, &user); err != nil {
		t.Fatalf("ユーザーの保存に失敗しました: %v", err)
	}
	if _, err = repo.Get(ctx, user.UserID); err != nil {
		t.Fatalf("保存したユーザーを取得しようとしましたが、エラーが発生しました: %v", err)
	}

	// ユーザーを削除
	if err := repo.Delete(ctx, user.UserID); err != nil {
		t.Fatalf("ユーザーの削除に失敗しました: %v", err)
	}

	// ユーザーが削除されたことを確認
	if _, err = repo.Get(ctx, user.UserID); err == nil {
		t.Fatalf("削除済みのユーザーを取得しようとしましたが、エラーが発生しませんでした")
	}

	// 削除済みのユーザーを削除しようとしても問題ないことを確認
	if err := repo.Delete(ctx, user.UserID); err != nil {
		t.Fatalf("削除済みのユーザーを削除しようとしたところ、エラーが発生しました: %v", err)
	}
}
