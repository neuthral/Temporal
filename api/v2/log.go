package v2

import (
	"github.com/gin-gonic/gin"
)

// LogError is a wrapper used by the API to handle logging of errors. Returns a
// callback to also fail a gin context with an optional status code, which
// defaults to http.StatusInternalServerError. Fields is an optional set of
// params provided in pairs, where the first of a pair is the key, and the second
// is the value
func (api *API) LogError(err error, message string, fields ...interface{}) func(c *gin.Context, code ...int) {
	// create base entry
	logger := api.l

	// write log
	if fields != nil && len(fields)%2 == 0 {
		logger = logger.With(fields...)
	}
	if fields != nil && len(fields)%2 == 0 {
		fields = append(fields, "error", err.Error())
		logger.Errorw(message, fields...)
	} else {
		logger.Errorw(message, "error", err.Error())
	}

	// return utility callback
	if message == "" && err != nil {
		return func(c *gin.Context, code ...int) { Fail(c, err, code...) }
	}
	return func(c *gin.Context, code ...int) { FailWithMessage(c, message, code...) }
}
