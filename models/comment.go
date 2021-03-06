package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	CommentID   primitive.ObjectID `json:"commentID" bson:"_id"`
	Username    string             `json:"username" bson:"username"`
	CommentToID primitive.ObjectID `json:"commentToID" bson:"commentToID"`
	CreatedTime time.Time          `json:"commentCreatedTime" bson:"commentCreatedTime"`
	NumOfThumb  int64              `json:"numOfThumb" bson:"numOfThumb"`
	Content     string             `json:"commentContent" bson:"commentContent"`
}
