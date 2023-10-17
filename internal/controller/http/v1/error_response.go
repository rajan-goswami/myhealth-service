package v1

import "github.com/gin-gonic/gin"

func ErrorResponse(c *gin.Context, httpStatus int, err error) {
	c.AbortWithStatusJSON(httpStatus, err)
}
