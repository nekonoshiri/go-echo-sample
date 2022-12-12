package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nekonoshiri/go-echo-sample/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	userCollection = "users"
)

type userDocument struct {
	// Note: UserID は、読みやすさや表現の変更しやすさを優先して string にしています。
	// よりパフォーマンスを上げたい場合はバイナリとしても格納できます。
	UserID       string            `bson:"_id"`
	Name         string            `bson:"name"`
	Status       domain.UserStatus `bson:"status"`
	RegisteredAt time.Time         `bson:"registered_at"`
}

type mongoUserRepository struct {
	collection *mongo.Collection
}

// *mongoUserRepository が domain.UserRepository を実装していることの確認
var _ domain.UserRepository = (*mongoUserRepository)(nil)

// MongoDB を用いた UserRepository の実装を返します。
// 第２引数で、デフォルトで使用するデータベースやコレクションを変更できます（テスト時に有用です）。
// 第３引数以降は無視されます。
func NewMongoUserRepository(client *mongo.Client, collection ...*mongo.Collection) *mongoUserRepository {
	col := client.Database(mongoDatabase).Collection(userCollection)
	if len(collection) > 0 {
		col = collection[0]
	}
	return &mongoUserRepository{
		collection: col,
	}
}

func (repo *mongoUserRepository) Get(ctx context.Context, userID string) (*domain.User, error) {
	filter := bson.M{"_id": userID}

	var result *userDocument
	if err := repo.collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("ユーザーの取得に失敗しました: %w", err)
	}

	return &domain.User{
		UserID:       result.UserID,
		Name:         result.Name,
		Status:       result.Status,
		RegisteredAt: result.RegisteredAt,
	}, nil
}

func (repo *mongoUserRepository) List(ctx context.Context, exclusiveStartKey string, limit int) ([]domain.User, string, error) {
	filter := bson.M{}
	if exclusiveStartKey != "" {
		filter = bson.M{"_id": bson.M{"$gt": exclusiveStartKey}}
	}

	opts := options.Find().SetSort(bson.M{"_id": 1})
	// limit が 0 または負数の場合は制限なし
	if limit > 0 {
		// limit より１つ多く取得する（続きがあるかどうか確認するため）
		opts.SetLimit(int64(limit) + 1)
	}

	cursor, err := repo.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", fmt.Errorf("ユーザーの取得に失敗しました: %w", err)
	}
	defer cursor.Close(ctx)

	users := []domain.User{}
	var lastEvaluatedKey string = ""

	// limit より１つ多く取得している（かもしれない）ので、愚直にループする
	for i := 0; limit <= 0 || i < limit; i++ {
		if !cursor.Next(ctx) {
			break
		}

		var result *userDocument
		if err := cursor.Decode(&result); err != nil {
			return nil, "", fmt.Errorf("取得したユーザーデータのデコードに失敗しました: %w", err)
		}
		users = append(users, domain.User{
			UserID:       result.UserID,
			Name:         result.Name,
			Status:       result.Status,
			RegisteredAt: result.RegisteredAt,
		})
		lastEvaluatedKey = result.UserID
	}

	if err := cursor.Err(); err != nil {
		return nil, "", fmt.Errorf("ユーザーの取得中にエラーが発生しました: %w", err)
	}

	// 続きが取得できない場合 lastEvaluatedKey は空文字列になる
	if !cursor.Next(ctx) {
		lastEvaluatedKey = ""
	}

	return users, lastEvaluatedKey, nil
}

func (repo *mongoUserRepository) Put(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.UserID}
	update := bson.M{"$set": userDocument{
		UserID:       user.UserID,
		Name:         user.Name,
		Status:       user.Status,
		RegisteredAt: user.RegisteredAt,
	}}

	_, err := repo.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return nil
}

func (repo *mongoUserRepository) Delete(ctx context.Context, userID string) error {
	filter := bson.M{"_id": userID}

	_, err := repo.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("ユーザーの削除に失敗しました: %w", err)
	}

	return nil
}
