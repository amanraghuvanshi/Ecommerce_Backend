package tokens

import (
	"context"
	"log"
	"os"
	"time"

	database "github.com/amanraghuvanshi/ecombackend/databases"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	UID        string
	jwt.StandardClaims
}

var Userdata *mongo.Collection = database.UserData(database.Client, "Users")

var SECRET_KEY = os.Getenv("SECRET_KEY")

func TokenGenerator(email string, first_name string, last_name string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: first_name,
		Last_name:  last_name,
		UID:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix()},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte("SECRET_KEY"))
	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, mesg string) {
	var msg string
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		msg = err.Error()
		// log.Panic(msg)
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "TOKEN INVALID"
		// log.Panic(msg)
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "TOKEN EXPIRED"
		// log.Panic(msg)
		return
	}
	return claims, msg
}

func UpdateAllToken(signedToken string, signedRefreshToken string, userID string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updatedObject primitive.D
	// we will keep appending the things that we are updating with this object

	updatedObject = append(updatedObject, bson.E{Key: "token", Value: signedToken})
	updatedObject = append(updatedObject, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updatedObject = append(updatedObject, bson.E{Key: "updatedat", Value: updated_at})

	upsert := true
	filter := bson.M{"user_id": userID}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := Userdata.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updatedObject},
	}, &opt)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}
}
