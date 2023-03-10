package main

import (
	"log"
	"os"

	"github.com/amanraghuvanshi/ecombackend/controllers"
	"github.com/amanraghuvanshi/ecombackend/middleware"
	"github.com/amanraghuvanshi/ecombackend/routes"
	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.LoggerWithConfig())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + PORT))

}
