package usecase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nekonoshiri/go-echo-sample/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
)

// GetUser ユースケースの正常系のテスト。
func TestGetUserOK(t *testing.T) {
	testCases := []struct {
		userID           string      // 取得しようとするユーザーの ID
		repositoryUser   domain.User // リポジトリが返すユーザー
		wantResponseBody string      // 期待されるレスポンスボディ
	}{
		{
			userID: "U1",
			repositoryUser: domain.User{
				UserID:       "U1",
				Name:         "ユーザー１",
				Status:       domain.UserStatusNormal,
				RegisteredAt: time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			wantResponseBody: `{
				"name": "ユーザー１",
				"status": "normal",
				"registeredAt": "1000-01-01T00:00:00Z"
			}`,
		},
		{
			userID: "U2",
			repositoryUser: domain.User{
				UserID:       "U2",
				Name:         "ユーザー２",
				Status:       domain.UserStatusFrozen,
				RegisteredAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			wantResponseBody: `{
				"name": "ユーザー２",
				"status": "frozen",
				"registeredAt": "2000-01-01T00:00:00Z"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("userID=%s", tc.userID), func(t *testing.T) {
			userRepository := &MockUserRepository{
				get: func(ctx context.Context, userID string) (*domain.User, error) {
					if userID != tc.userID {
						t.Fatalf("ユーザー ID が %q ではなく %q のユーザーを取得しようとしました", tc.userID, userID)
					}
					return &tc.repositoryUser, nil
				},
			}

			e := echo.New()
			request := httptest.NewRequest(http.MethodGet, "/users/:userID", nil)
			recorder := httptest.NewRecorder()
			c := e.NewContext(request, recorder)
			c.SetParamNames("userID")
			c.SetParamValues(tc.userID)

			if err := GetUser(c, userRepository); err != nil {
				t.Fatalf("ユースケースがエラーを返しました: %v", err)
			}
			if recorder.Code != http.StatusOK {
				t.Errorf("期待される HTTP ステータスコードは %d ですが、%d が返りました", http.StatusOK, recorder.Code)
			}
			if diff := cmp.Diff(tc.wantResponseBody, recorder.Body.String(), UnmarshalJSON); diff != "" {
				t.Errorf("期待されるリクエストボディ (-) と実際のリクエストボディ (+) が一致しませんでした:\n%s", diff)
			}
		})
	}
}

// GetUser ユースケースのリクエストのバリデーションのテスト。
func TestGetUserBadRequest(t *testing.T) {
	userRepository := &MockUserRepository{}

	testCases := []struct {
		userID string // 取得しようとするユーザーの ID
	}{
		{userID: ""},
		{userID: strings.Repeat("", 101)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("userID=%s", tc.userID), func(t *testing.T) {
			e := echo.New()
			request := httptest.NewRequest(http.MethodGet, "/users/:userID", nil)
			c := e.NewContext(request, nil)
			c.SetParamNames("userID")
			c.SetParamValues(tc.userID)

			err := GetUser(c, userRepository)
			if err == nil {
				t.Fatalf("ユースケースがエラーを返すはずですが、返しませんでした")
			}

			statusCode, errorResponse := ParseErrorResponse(t, err)
			if statusCode != http.StatusBadRequest {
				t.Errorf("期待される HTTP ステータスコードは %d ですが、%d が返りました", http.StatusBadRequest, statusCode)
			}
			if errorResponse.Code != "BadRequest" {
				t.Errorf("期待されるエラーコードは %s ですが、%s が返りました", "BadRequest", errorResponse.Code)
			}
		})
	}
}

// GetUser ユースケースのユーザーが見つからない場合のテスト。
func TestGetUserUserNotFound(t *testing.T) {
	userRepository := &MockUserRepository{
		get: func(ctx context.Context, userID string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/users/:userID", nil)
	c := e.NewContext(request, nil)
	c.SetParamNames("userID")
	c.SetParamValues("U1")

	err := GetUser(c, userRepository)
	if err == nil {
		t.Fatalf("ユースケースがエラーを返すはずですが、返しませんでした")
	}

	statusCode, errorResponse := ParseErrorResponse(t, err)
	if statusCode != http.StatusBadRequest {
		t.Errorf("期待される HTTP ステータスコードは %d ですが、%d が返りました", http.StatusBadRequest, statusCode)
	}
	if errorResponse.Code != "UserNotFound" {
		t.Errorf("期待されるエラーコードは %s ですが、%s が返りました", "UserNotFound", errorResponse.Code)
	}
}
