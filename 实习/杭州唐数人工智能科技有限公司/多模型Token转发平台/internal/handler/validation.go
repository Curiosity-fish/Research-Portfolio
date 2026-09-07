package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/server/respond"
)

// bindAndValidate binds the JSON request body and maps validation errors to a
// domain.ValidationError. It returns false when binding or validation fails;
// the caller should return immediately because the response has already been
// written.
func bindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		details := map[string]string{"body": "请求体格式错误或缺少必填字段"}
		if ve, ok := err.(validator.ValidationErrors); ok {
			details = validationDetails(ve)
		}
		respond.Error(c, domain.NewValidationError(details))
		return false
	}
	return true
}

// validationDetails converts go-playground validator errors into a map of
// field -> message. Only the first error per field is kept to keep responses
// concise.
func validationDetails(ve validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(ve))
	for _, fe := range ve {
		if _, ok := details[fe.Field()]; ok {
			continue
		}
		details[fe.Field()] = validationMessage(fe)
	}
	return details
}

// validationMessage returns a Chinese message for a validator tag.
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "该字段为必填项"
	case "min":
		return "长度或值过小"
	case "max":
		return "长度或值过大"
	default:
		return "字段校验失败"
	}
}
