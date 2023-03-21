package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/amanraghuvanshi/ecombackend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		UID := c.Query("id")
		if UID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Code"})
			c.Abort()
			return
		}
		address, err := primitive.ObjectIDFromHex(UID)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
		}

		var addresses models.Address

		addresses.Address_ID = primitive.NewObjectID()

		if err = c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, gin.H{"Error": err.Error()})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		// we are finding out the addresses that are already there.
		//matching stage
		match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}

		// unwind
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}

		// group stage
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})

		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Server ran into some error")
		}

		var addressinfo []bson.M
		if err = pointCursor.All(ctx, &addressinfo); err != nil {
			panic(err)
		}
		var size int32

		for _, address_no := range addressinfo {
			count := address_no["count"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, "Error while updating the data")
				return
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "Not allowed")
		}
		defer cancel()
		ctx.Done()

	}
}

func EditAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		UID := c.Query("id")
		if UID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "INVALID OR EMPTY ID"})
			c.Abort()
			return
		}
		user_TID, err := primitive.ObjectIDFromHex(UID)
		if err != nil {
			c.IndentedJSON(500, "Error while fetching!")
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}
		// filtering data
		filter := bson.D{primitive.E{Key: "_id", Value: user_TID}}

		// updation
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editAddress.House}, {Key: "address.0.street_name", Value: editAddress.Street}, {Key: "address.0.city_name", Value: editAddress.City}, {Key: "address.0.pincode", Value: editAddress.Pincode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
			return
		}

		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Updation Successful")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		UID := c.Query("id")
		if UID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "INVALID OR EMPTY ID"})
			c.Abort()
			return
		}
		user_TID, err := primitive.ObjectIDFromHex(UID)
		if err != nil {
			c.IndentedJSON(500, "Error while fetching!")
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// we use the struct so that we can have the access to the data structure.
		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}

		// filtering data
		filter := bson.D{primitive.E{Key: "_id", Value: user_TID}}

		// updation
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editAddress.House}, {Key: "address.1.street_name", Value: editAddress.Street}, {Key: "address.1.city_name", Value: editAddress.City}, {Key: "address.1.pincode", Value: editAddress.Pincode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Error while updating")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Updation Successful")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		UID := c.Query("id")

		if UID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{
				"Error": "Invalid Search Index"})
			c.Abort()
			return
		}
		// whenever a server is working with database, so we can't wait for the database response forever, so this is exactly why we use database.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		// addresses list.
		addresses := make([]models.Address, 0)
		user_TID, err := primitive.ObjectIDFromHex(UID)
		if err != nil {
			c.IndentedJSON(500, "Internal Server Error")
			return
		}
		//filter out the data.
		filter := bson.D{primitive.E{Key: "_id", Value: user_TID}}
		// this says that we are putting the value of address into the addresses.And using the filter we will be finding the respective values
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(404, "Error in Updation")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "Successfully Deleted")
	}
}
