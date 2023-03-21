package database

import (
	"errors"
	"log"
	"time"

	"github.com/amanraghuvanshi/ecombackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

var (
	ErrorCantFindProduct    = errors.New("can't find the product")
	ErrorCantDecodeProducts = errors.New("can't find the product")
	ErrorUserNotvalid       = errors.New("user is not valid or userID is not valid")
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

	// since we are checking out, so there will be a order id to be generated for that order, so we are creating our order cart
	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_at = time.Now()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Payment_method.COD = true

	// here, we are unwinding our usercart in order to get all the details on the user cart
	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "usercart"}}}}

	// here we are trying to get the price of the cart by grouping them all together,
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

	// now we have the total/added price of the cart,
	orderCart.Price = int(total_Price)
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}

	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		panic(err)
	}

	if err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems); err != nil {
		log.Println(err)
	}

	// adding of data to the user's cart
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}

	// empty the cart
	userCart_empty := make([]models.ProductUser, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: id}}
	update3 := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: userCart_empty}}}}
	_, err = userCollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		return ErrorCantBuyCartItems
	}
	return nil
}

// this function, will take a product and just checkout it, without adding it to the cart.
func InstantBuyer(ctx context.Context, prodCollection *mongo.Collection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrorUserNotvalid
	}

	var prodDetails models.ProductUser
	var order_details models.Order

	order_details.Order_ID = primitive.NewObjectID()
	order_details.Ordered_at = time.Now()
	order_details.Order_Cart = make([]models.ProductUser, 0)
	order_details.Payment_method.COD = true

	if err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&prodDetails); err != nil {
		log.Println(err)
	}

	order_details.Price = prodDetails.Price

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: order_details}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	// adding the details of the order
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": prodDetails}}

	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}
	return nil
}
