package usecase

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// エラーレスポンス。
type ErrorResponse struct {
	// エラーメッセージ。
	Message string `json:"message"`
	// エラーコード。これは HTTP ステータスコードではなく、ユースケース毎に定義されるコードです。
	Code string `json:"code"`
	// ラップしているエラー。レスポンス JSON には含まれません。
	Err error `json:"-"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Code, e.Message, e.Err)
}

func (e *ErrorResponse) Unwrap() error {
	return e.Err
}

// クライアントにエラーレスポンスを返します。
func newErrorResponse(c echo.Context, httpStatusCode int, errorCode string, errorMessage string, err error) error {
	return echo.NewHTTPError(httpStatusCode, &ErrorResponse{
		Code:    errorCode,
		Message: errorMessage,
		Err:     err,
	})
}

// クライアントに不正なリクエストエラー（HTTP ステータスコード 400, エラーコード BadRequest）を返します。
func badRequest(c echo.Context, message string, err error) error {
	return newErrorResponse(c, http.StatusBadRequest, "BadRequest", message, err)
}

// クライアントにサーバーエラー（HTTP ステータスコード 500, エラーコード InternalServerError）を返します。
func internalServerError(c echo.Context, message string, err error) error {
	return newErrorResponse(c, http.StatusInternalServerError, "InternalServerError", message, err)
}
