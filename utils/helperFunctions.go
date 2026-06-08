package utils

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetPrimitiveUserID(c *gin.Context) primitive.ObjectID {
	StringuserID := c.MustGet("userId").(string)
	userId, err := primitive.ObjectIDFromHex(StringuserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return primitive.NilObjectID
	}
	return userId
}

func GenerateRoomCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	roomCode := make([]byte, length)

	for i := range roomCode {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		roomCode[i] = charset[num.Int64()]
	}

	return string(roomCode), nil
}
