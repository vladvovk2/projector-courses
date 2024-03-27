package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	influxdb "webapp/influx"
	"webapp/mongo"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	mongoURI = "mongodb://mongodb:27017" // MongoDB connection URI
	dbName   = "main"                    // Your MongoDB database name
	collName = "users"                   // Your MongoDB collection name
)

type User struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Username string `json:"username" bson:"username"`
}

func main() {
	// Initialize MongoDB connection
	influxdb.Init()
	defer influxdb.Close()

	if err := mongo.Init(mongoURI); err != nil {
		log.Fatal("Error initializing MongoDB:", err)
	}

	defer func() {
		// Disconnect from MongoDB when the application exits
		if client := mongo.GetClient(); client != nil {
			if err := client.Disconnect(context.Background()); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// Set up HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/users", metricsMiddleware(getUsersHandler)).Methods("GET")
	router.HandleFunc("/users", metricsMiddleware(postUsersHandler)).Methods("POST")
	router.HandleFunc("/users", metricsMiddleware(deleteAllUsersHandler)).Methods("DELETE")
	router.HandleFunc("/health", metricsMiddleware(healthCheckHandler))

	// Start HTTP server
	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Access MongoDB client from the mongo package
	client := mongo.GetClient()

	// Access a MongoDB database
	db := client.Database(dbName)

	// Access a MongoDB collection
	collection := db.Collection(collName)

	// Retrieve all documents from MongoDB collection
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Println("Error finding documents:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		log.Println("Error decoding documents:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Encode the selected documents as JSON and send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func postUsersHandler(w http.ResponseWriter, r *http.Request) {
	client := mongo.GetClient()
	db := client.Database(dbName)
	collection := db.Collection(collName)

	// Define 10 users to insert
	users := []interface{}{
		bson.M{"name": "User 1", "email": "user1@example.com", "username": "user1"},
		bson.M{"name": "User 2", "email": "user2@example.com", "username": "user2"},
		bson.M{"name": "User 3", "email": "user3@example.com", "username": "user3"},
		bson.M{"name": "User 4", "email": "user4@example.com", "username": "user4"},
		bson.M{"name": "User 5", "email": "user5@example.com", "username": "user5"},
		bson.M{"name": "User 6", "email": "user6@example.com", "username": "user6"},
		bson.M{"name": "User 7", "email": "user7@example.com", "username": "user7"},
		bson.M{"name": "User 8", "email": "user8@example.com", "username": "user8"},
		bson.M{"name": "User 9", "email": "user9@example.com", "username": "user9"},
		bson.M{"name": "User 10", "email": "user10@example.com", "username": "user10"},
	}

	// Insert users into the collection
	result, err := collection.InsertMany(context.Background(), users)
	if err != nil {
		log.Println("Error inserting documents:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	influxdb.WriteToInfluxDB("user_operations", map[string]string{"operation": "insert"}, map[string]interface{}{"count": len(users)})

	// Create a response struct

	response := struct {
		Message string   `json:"message"`
		UserIDs []string `json:"userIds"`
	}{
		Message: "10 users inserted successfully",
		UserIDs: make([]string, len(result.InsertedIDs)),
	}

	// Populate the UserIDs with the inserted IDs
	for i, id := range result.InsertedIDs {
		response.UserIDs[i] = id.(string)
	}

	// Set content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func deleteAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	client := mongo.GetClient()
	db := client.Database(dbName)
	collection := db.Collection(collName)

	// Delete all documents in the collection
	result, err := collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Println("Error deleting documents:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	influxdb.WriteToInfluxDB("user_operations", map[string]string{"operation": "delete"}, map[string]interface{}{"count": result.DeletedCount})

	// Create a response struct
	response := struct {
		Message      string `json:"message"`
		DeletedCount int64  `json:"deletedCount"`
	}{
		Message:      "All users deleted successfully",
		DeletedCount: result.DeletedCount,
	}

	// Set content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Perform any necessary health checks, e.g., database connectivity, external dependencies, etc.
	// Respond with a 200 OK if everything is healthy.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "up"})
}

func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Call the handler
		next.ServeHTTP(w, r)

		// Calculate request duration
		duration := time.Now().Sub(startTime)

		// Record the request count and duration in InfluxDB
		influxdb.WriteToInfluxDB("webapp", map[string]string{
			"method": r.Method,
			"path":   r.URL.Path,
		}, map[string]interface{}{
			"count":    1,
			"duration": duration.Seconds(),
		})
	}
}
