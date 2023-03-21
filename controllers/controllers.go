package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	database "github.com/amanraghuvanshi/ecombackend/databases"
	"github.com/amanraghuvanshi/ecombackend/models"
	"github.com/amanraghuvanshi/ecombackend/tokens"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")

var Validate = validator.New()

func Hashpassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Invalid Credentials"
		valid = false
	}
	return valid, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		// marhsalling
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		// validation for all the feild in the structs
		Validate := validator.New()
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": validationErr})
		}

		// check for the email
		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "User already Exist"})
			return
		}

		// check for phone number
		countPhn, err := UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()
		if countPhn > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Phone already Exist"})
			return
		}

		// updating other fields and generating token
		password := Hashpassword(*user.Password)
		user.Password = &password
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refresh_token, _ := tokens.TokenGenerator(*user.Email, *user.First_name, *user.Last_name, user.User_ID)
		user.Token = &token
		user.Refresh_token = &refresh_token
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_status = make([]models.Order, 0)

		// inserting the user in the database
		_, insertErr := UserCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error, user can't be created"})
			return
		}
		defer cancel()

		c.JSON(http.StatusCreated, "Sign-in Successful")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		// Marhsalling the data
		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		// checking if user exists
		if err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser); err != nil {
			c.JSON(http.StatusConflict, gin.H{"Error": "User doesn't exists!"})
			return
		}

		defer cancel()

		// Verifying the user
		passValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !passValid {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			fmt.Println(msg)
			return
		}

		token, refresh_token, _ := tokens.TokenGenerator(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_ID)

		defer cancel()

		tokens.UpdateAllToken(token, refresh_token, foundUser.User_ID)

		c.JSON(http.StatusFound, foundUser)

	}
}

// For products
func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product

		defer cancel()

		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
			return
		}

		products.Product_ID = primitive.NewObjectID()
		_, anyErr := ProductCollection.InsertOne(ctx, products)
		if anyErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added the Product Admin")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productlist []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong")
			return
		}
		if err = cursor.All(ctx, &productlist); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}

		defer cancel()
		c.IndentedJSON(200, productlist)
	}
}

func SearchProductbyQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var SearchProducts []models.Product
		queryParams := c.Query("name")

		if queryParams == "" {
			log.Println("Query Empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"Error": "Invalid search index"})
			c.Abort()
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQuery, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParams}})

		if err != nil {
			c.IndentedJSON(404, "Something went wrong while fetching the data")
			return
		}
		if err = searchQuery.All(ctx, SearchProducts); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid data")
			return
		}
		defer searchQuery.Close(ctx)

		if err := searchQuery.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid requests")
			return
		}
		defer cancel()
		c.IndentedJSON(200, SearchProducts)
	}
}
