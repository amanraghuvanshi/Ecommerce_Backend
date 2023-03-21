package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	database "github.com/amanraghuvanshi/ecombackend/databases"
	"github.com/amanraghuvanshi/ecombackend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Instance for updation on the database
type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Param("id")
		if productQueryID == "" {
			log.Println("Product ID empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("empty product id"))
			return
		}

		userQueryId := c.Query("userID")
		if userQueryId == "" {
			log.Println("User ID is empty")

			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}

		pID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err = database.AddProductToCart(ctx, app.prodCollection, app.userCollection, pID, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Added to Cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Param("id")

		if productQueryID == "" {
			log.Println("product ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("empty product id"))
			return
		}

		userId := c.Param("user_id")
		if userId == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user Id is empty"))
			return
		}

		// getting the product ID
		pID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err = database.RemoveItemFromCart(ctx, app.prodCollection, app.userCollection, pID, userId)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}

		c.JSON(200, "Removed from your Cart")
	}
}

func GetItemsFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		// getting access to user id and handling error
		UID := c.Query("id")
		if UID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Errors": "INVALID OR NIL ID"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		user_TID, _ := primitive.ObjectIDFromHex(UID)

		// for the cart details
		var filledCart models.User

		if err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "id", Value: user_TID}}).Decode(&filledCart); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		// match stage, this will be used for getting the data from the user that matches the criteria
		filter_Match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_TID}}}}

		// unwind stage, this helps to work with the data that we get after the filter stage, since the data is in array from so, we will be unwinding it.
		unwind := bson.D{{Key: "$unwind", Value: primitive.D{primitive.E{Key: "_id", Value: "$usercart"}}}}

		//group stage, this will be the stage where we will be adding the data
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}

		// aggregation function
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_Match, unwind, grouping})

		if err != nil {
			log.Println(err)
		}

		// returning of data in readable format.
		var listing []bson.M
		if err = pointCursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledCart.UserCart)
		}
		ctx.Done()
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		UserQueryId := c.Query("id")
		if UserQueryId == "" {
			log.Panic("ID EMPTY")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("No ID found"))
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		err := database.BuyItemfromCart(ctx, app.userCollection, UserQueryId)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}

		c.IndentedJSON(http.StatusOK, "Order Placed Successfully")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryId := c.Param("id")
		if productQueryId == "" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		user_id := c.Param("user_id")
		if user_id == "" {
			log.Println("User ID is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user Id is empty"))
			return
		}
		pid, err := primitive.ObjectIDFromHex(productQueryId)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, pid, user_id)

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(200, "Purchase Successful")
	}
}
