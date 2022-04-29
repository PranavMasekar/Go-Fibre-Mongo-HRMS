package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017/" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}
	mg = MongoInstance{Client: client, Db: db}
	return nil
}

func GetEmployee(c *fiber.Ctx) error {
	// Query for Mongo db => We want all data so empty brackets
	query := bson.D{{}}
	// Getting the data from the mongo db
	cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	var employees []Employee = make([]Employee, 0)
	// Converting the data received into strycts of tyep Employee
	if err := cursor.All(c.Context(), &employees); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	// Returning the response
	return c.JSON(employees)
}

func CreateEmployee(c *fiber.Ctx) error {
	collection := mg.Db.Collection("employees")

	employee := new(Employee)
	// Get the body of the request
	if err := c.BodyParser(employee); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	// Set id to empty string so mongo db can assign id
	employee.ID = ""
	// Insert the created obj to database
	insertionResult, err := collection.InsertOne(c.Context(), employee)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	// Create query to find the created employee
	filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
	// Get the created employee
	createdRecord := collection.FindOne(c.Context(), filter)

	createdEmployee := &Employee{}
	// Convert that employee to struct
	createdRecord.Decode(createdEmployee)
	// Return the response
	return c.Status(201).JSON(createdEmployee)
}

func UpdateEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	// Convert hex
	employeeId, err := primitive.ObjectIDFromHex(idParam)

	if err != nil {
		return c.SendStatus(400)
	}

	employee := new(Employee)
	// Get body of request
	if err := c.BodyParser(employee); err != nil {
		c.Status(400).SendString(err.Error())
	}
	// Search query
	query := bson.D{{Key: "_id", Value: employeeId}}
	// Update query
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{

				{Key: "name", Value: employee.Name},
				{Key: "age", Value: employee.Age},
				{Key: "salary", Value: employee.Salary},
			},
		},
	}
	// Find the employee and update
	err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.SendStatus(400)
		}
		return c.SendStatus(500)
	}

	employee.ID = idParam
	// return JSON
	return c.Status(200).JSON(employee)

}

func DeleteEmployee(c *fiber.Ctx) error {
	employeeId, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		return c.SendStatus(400)
	}
	// Search query
	query := bson.D{{Key: "_id", Value: employeeId}}
	// Delete from DB
	result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), &query)

	if err != nil {
		c.SendStatus(500)
	}

	if result.DeletedCount < 1 {
		c.SendStatus(404)
	}
	return c.Status(200).JSON("Record Deleted")
}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()

	app.Get("/employee", GetEmployee)
	app.Post("/employee", CreateEmployee)
	app.Put("/employee/:id", UpdateEmployee)
	app.Delete("/employee/:id", DeleteEmployee)

	log.Fatal(app.Listen(":3000"))
}
