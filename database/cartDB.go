package database

import (
	"errors"
	"log"
	"time"

	"github.com/amanraghuvanshi/ecombackend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

var (
	ErrorCantFindProduct    = errors.New("can't find the product")
	ErrorCantDecodeProducts = errors.New("can't find the product")
	ErrorUserNotvalid       = errors.New("user is not valid")
	ErrorCantUpdateUser     = errors.New("can't add the product to cart")
	ErrorCantRemoveItem     = errors.New("can't remove the product from cart")
	ErrorCantGetItem        = errors.New("can't get the item from cart")
	ErrorCantBuyCartItems   = errors.New("can't update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchfromDB, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrorCantFindProduct
	}
	var productCart []models.ProductUser
	if err = searchfromDB.All(ctx, &productCart); err != nil {
		log.Println(err)
		return ErrorCantDecodeProducts
	}
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrorUserNotvalid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrorUserNotvalid
	}
	return nil
}

func RemoveItemFromCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	// we can update the item with empty or zero data
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrorUserNotvalid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrorCantRemoveItem
	}
	return nil
}

func BuyItemfromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	// fetch the cart
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrorUserNotvalid
	}

	// find the cart value

	var getCartItems models.User
	var orderCart models.Order

	// since we are checking out, so there will be a order id to be generated for that order
	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_at = time.Now()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Payment_method.COD = true

	// here, we are unwinding our usercart in order to get all the details on the user cart
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "usercart"}}}}

	// here we are trying to get the price of the cart by grouping them all together
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "usercart.price"}}}}}}

	curRes, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	if err != nil {
		panic(err)
	}
	ctx.Done()

	// creating an order with the items,
	//  here we are trying to convert all the item to something that golang can understand
	var getUserCart []bson.M
	if err = curRes.All(ctx, &getUserCart); err != nil {
		panic(err)
	}

	var total_Price int32
	for _, user_item := range getUserCart {
		price := user_item["total"]
		total_Price = price.(int32)
	}

	// empty the cart
}

func InstantBuyer() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
