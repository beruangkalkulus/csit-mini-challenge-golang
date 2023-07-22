package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	router := gin.Default()
	router.GET("/", indexHandler)
	router.GET("/flight", getFlight)
	router.GET("/hotel", getHotel)

	router.Run(":8080")
}

func indexHandler(c *gin.Context) {
	fmt.Println("indexHandler")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World!",
	})
}

func getFlight(c *gin.Context) {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("No MONGODB_URI environment variable found")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("minichallenge").Collection("flights")

	// Get query parameters
	departureDateString := c.Query("departureDate")
	returnDateString := c.Query("returnDate")

	departureDate, err := time.Parse("2006-01-02", departureDateString)
	returnDate, err := time.Parse("2006-01-02", returnDateString)

	origin := "Singapore"
	destination := c.Query("destination")

	// FINDING DEPARTURE FLIGHTS

	departureQuery := bson.D{
		{Key: "srccity", Value: origin},
		{Key: "destcity", Value: destination},
		{Key: "date", Value: departureDate},
	}

	// Find all documents in which the "i" field equals 71
	opts := options.Find().SetSort(bson.D{{Key: "price", Value: 1}})
	cursor, err := coll.Find(context.TODO(), departureQuery, opts)

	if err != nil {
		panic(err)
	}

	// Get a list of all returned documents and print them out
	var depResults []bson.M
	if err = cursor.All(context.TODO(), &depResults); err != nil {
		panic(err)
	}

	if len(depResults) == 0 {
		fmt.Println("No departure flights found")
		emptyResult := make([]gin.H, 0)
		c.IndentedJSON(http.StatusOK, emptyResult)
		return
	}

	mainDepFlight := depResults[0]

	var depFlights []bson.M

	for _, depFlight := range depResults {
		if depFlight["price"] == mainDepFlight["price"] {
			depFlights = append(depFlights, depFlight)	
		} else {
			break
		}
	}

	// FINDING RETURN FLIGHTS

	returnQuery := bson.D{
		{Key: "srccity", Value: destination},
		{Key: "destcity", Value: origin},
		{Key: "date", Value: returnDate},
	}

	// Find all documents in which the "i" field equals 71
	cursor, err = coll.Find(context.TODO(), returnQuery, opts)

	if err != nil {
		panic(err)
	}

	// Get a list of all returned documents and print them out
	var retResults []bson.M
	if err = cursor.All(context.TODO(), &retResults); err != nil {
		panic(err)
	}

	if len(retResults) == 0 {
		fmt.Println("No return flights found")
		emptyResult := make([]gin.H, 0)
		c.IndentedJSON(http.StatusOK, emptyResult)
		return
	}
	
	mainRetFlight := retResults[0]

	var retFlights []bson.M

	for _, retFlight := range retResults {
		if retFlight["price"] == mainRetFlight["price"] {
			retFlights = append(retFlights, retFlight)	
		} else {
			break
		}
	}

	var results []gin.H

	for _, depFlight := range depFlights {
		for _, retFlight := range retFlights {
			result := gin.H{
				"City":              destination,
				"Departure Date":    departureDate,
				"Departure Airline": depFlight["airlinename"],
				"Departure Price":   depFlight["price"],
				"Return Date":       returnDate,
				"Return Airline":    retFlight["airlinename"],
				"Return Price":      retFlight["price"],
			}
			results = append(results, result)
		}
	}

	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, results)
}

func getHotel(c *gin.Context) {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("No MONGODB_URI environment variable found")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("minichallenge").Collection("hotels")

	// Get query parameters
	checkInDateString := c.Query("checkInDate")
	checkOutDateString := c.Query("checkOutDate")
	destination := c.Query("destination")

	checkInDate, err := time.Parse("2006-01-02", checkInDateString)
	checkOutDate, err := time.Parse("2006-01-02", checkOutDateString)

	// Aggregate Pipeline
	pipeline := bson.A{
		// Match stage
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "city", Value: destination},
			{Key: "date", Value: bson.D{{Key: "$gte", Value: checkInDate}, {Key: "$lte", Value: checkOutDate}}},
		}}},
		// Group stage
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$hotelName"}, 
			{Key: "price", Value: bson.D{{Key: "$sum", Value: "$price"}}},
		}}},
		// Sort stage
		bson.D{{Key: "$sort", Value: bson.D{{Key: "price", Value: 1}}}},
	}

	cursor, err := coll.Aggregate(context.TODO(), pipeline)

	if err != nil {
		panic(err)
	}

	var result []bson.M
	if err = cursor.All(context.TODO(), &result); err != nil {
		panic(err)
	}

	if len(result) == 0 {
		emptyResult := make([]gin.H, 0)
		c.IndentedJSON(http.StatusOK, emptyResult)
		return
	}

	bestHotel := result[0]

	var results []gin.H

	for _, hotel := range result {
		if hotel["price"] == bestHotel["price"] {
			response := gin.H{
				"City":           destination,
				"Check In Date":  checkInDateString,
				"Check Out Date": checkOutDateString,
				"Hotel":          hotel["_id"],
				"Price":          hotel["price"],
			}
			results = append(results, response)
		} else {
			break
		}
	}

	c.IndentedJSON(http.StatusOK, results)
}
