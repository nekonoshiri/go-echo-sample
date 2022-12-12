package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"nekonoshiri/go-echo-sample/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
)

// cmp.Diff() のオプションとして渡すと、文字列を JSON として比較します。
//
// 例: cmp.Diff(`{"key": "value"}`, `{"key": "value2"}`, UnmarshalJSON)
var UnmarshalJSON = cmp.FilterValues(
	func(x, y string) bool {
		return json.Valid([]byte(x)) && json.Valid([]byte(y))
	},
	cmpopts.AcyclicTransformer("UnmarshalJSON", func(in string) interface{} {
		var out interface{}
		if err := json.Unmarshal([]byte(in), &out); err != nil {
			panic(err)
		}
		return out
	}),
)

// ユースケースの返すエラーから HTTP ステータスコードとエラーレスポンスを取り出します。
func ParseErrorResponse(t *testing.T, err error) (int, *ErrorResponse) {
	t.Helper()

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("ユースケースが HTTPError 以外のエラーを返しました: %v", err)
	}

	response, ok := httpErr.Message.(*ErrorResponse)
	if !ok {
		t.Fatalf("ユースケースが ErrorResponse 以外のエラーレスポンスを返しました: %v", httpErr.Message)
	}

	return httpErr.Code, response
}

// テスト用の UserRepository。
type MockUserRepository struct {
	get    func(ctx context.Context, userID string) (*domain.User, error)
	list   func(ctx context.Context, exclusiveStartKey string, limit int) (users []domain.User, lastEvaluatedKey string, err error)
	put    func(ctx context.Context, user *domain.User) error
	delete func(ctx context.Context, userID string) error
}

func (repo *MockUserRepository) Get(ctx context.Context, userID string) (*domain.User, error) {
	if repo.get != nil {
		return repo.get(ctx, userID)
	}
	return nil, errors.New("実装されていません")
}

func (repo *MockUserRepository) List(ctx context.Context, exclusiveStartKey string, limit int) (users []domain.User, lastEvaluatedKey string, err error) {
	if repo.list != nil {
		return repo.list(ctx, exclusiveStartKey, limit)
	}
	return nil, "", errors.New("実装されていません")
}

func (repo *MockUserRepository) Put(ctx context.Context, user *domain.User) error {
	if repo.put != nil {
		return repo.put(ctx, user)
	}
	return errors.New("実装されていません")
}

func (repo *MockUserRepository) Delete(ctx context.Context, userID string) error {
	if repo.delete != nil {
		return repo.delete(ctx, userID)
	}
	return errors.New("実装されていません")
}
