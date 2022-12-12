package usecase

import (
	"errors"
	"fmt"
	"time"

	"nekonoshiri/go-echo-sample/domain"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
)

// GetUser ユースケースのリクエスト。
type GetUserRequest struct {
	// ユーザー ID。必須で、1 文字以上 100 文字以下です。
	UserID string `param:"userID"`
}

func (request *GetUserRequest) validate() error {
	return validation.ValidateStruct(request,
		validation.Field(&request.UserID,
			validation.Required.Error("ユーザー ID は必須です"),
			validation.RuneLength(1, 100).Error("ユーザー ID は 1 文字以上 100 文字以下です"),
		),
	)
}

// GetUser ユースケースのレスポンス。
type GetUserResponse struct {
	// 名前。必須で、1 文字以上 100 文字以下です。
	Name string `json:"name"`
	// ステータス。必須で、normal か frozen です。
	Status string `json:"status"`
	// 登録日時。必須です。
	RegisteredAt time.Time `json:"registeredAt"`
}

func (response *GetUserResponse) validate() error {
	return validation.ValidateStruct(response,
		validation.Field(&response.Name,
			validation.Required.Error("名前は必須です"),
			validation.RuneLength(1, 100).Error("名前は 1 文字以上 100 文字以下です"),
		),
		validation.Field(&response.Status,
			validation.Required.Error("ステータスは必須です"),
			validation.In("normal", "frozen").Error("ステータスは normal か frozen です"),
		),
		validation.Field(&response.RegisteredAt,
			validation.Required.Error("登録日時は必須です"),
		),
	)
}

// GetUser ユースケース。ユーザーを取得します。
//   - リクエスト: [GetUserRequest]
//   - レスポンス: [GetUserResponse]
//
// このユースケースは、以下のエラーコードを返します。
//   - BadRequest: リクエストが不正な場合。
//   - UserNotFound: ユーザーが見つからなかった場合。
//   - InternalServerError: サーバーエラーが発生した場合。
func GetUser(c echo.Context, userRepository domain.UserRepository) error {
	ctx := c.Request().Context()

	var request GetUserRequest
	if err := c.Bind(&request); err != nil {
		return badRequest(c, "リクエストが不正です", err)
	}
	if err := request.validate(); err != nil {
		if errs, ok := err.(validation.Errors); ok {
			return badRequest(c, fmt.Sprintf("リクエストが不正です: %v", errs), err)
		}
		return internalServerError(c, "リクエストのバリデーションに失敗しました", err)
	}

	user, err := userRepository.Get(ctx, request.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return newErrorResponse(c, 400, "UserNotFound", "ユーザーが見つかりませんでした", err)
		}
		return internalServerError(c, "ユーザーの取得に失敗しました", err)
	}

	response := GetUserResponse{
		Name:         user.Name,
		Status:       string(user.Status),
		RegisteredAt: user.RegisteredAt,
	}
	if err := response.validate(); err != nil {
		return internalServerError(c, "レスポンスのバリデーションに失敗しました", fmt.Errorf("%+v: %w", response, err))
	}

	return c.JSON(200, response)
}
