package main

import (
	"log"
	"os"

	"github.com/amanraghuvanshi/ecombackend/controllers"
	database "github.com/amanraghuvanshi/ecombackend/databases"
	"github.com/amanraghuvanshi/ecombackend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	// router.Use(middleware.Authentication())
	// router.GET("/addtocart", app.AddToCart())
	// router.GET("/removeitem", app.RemoveItem())
	// router.GET("/cartcheckout", app.BuyFromCart())
	// router.GET("/instantbuy", app.InstantBuy())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/listcart", controllers.GetItemsFromCart())
	router.POST("/addaddress", controllers.AddAddress())
	router.PUT("/edithomeaddress", controllers.EditAddress())
	router.PUT("/editworkaddress", controllers.EditWorkAddress())
	router.GET("/deleteaddresses", controllers.DeleteAddress())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())
	log.Fatal(router.Run(":" + PORT))

}
