package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Global reference
var endpoint string = "https://go-shorty.herokuapp.com"
var db *mgo.Database
var collection string
var port string

func main() {
	// Connect to MongoDB
	session, err := mgo.Dial(os.Getenv("MONGODB_URI"))
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(os.Getenv("DB"))
	collection = os.Getenv("COLLECTION")
	port = os.Getenv("PORT")

	// Instatiate the Mux router
	router := mux.NewRouter()
	// Create routes
	router.HandleFunc("/{id}", redirectHandler).Methods("GET")
	router.HandleFunc("/shorten", shortenHandler).Methods("POST")
	router.HandleFunc("/expand", expandHandler).Methods("POST")
	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
	// Attatch the router and start the server
	log.Fatal(http.ListenAndServe(":"+port, router))
}

type myURL struct {
	ID       string `bson:"id" json:"id,omitempty"`
	LongURL  string `bson:"longURL" json:"longURL,omitempty"`
	ShortURL string `bson:"shortURL" json:"shortURL,omitempty"`
}

/** Handler to shorten a long URL **/
func shortenHandler(res http.ResponseWriter, req *http.Request) {
	var url myURL
	// Extract JSON from body
	err := json.NewDecoder(req.Body).Decode(&url)
	if err != nil {
		log.Fatal(err)
		// Respond with an error
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}

	log.Println("Received request to shorten URL:", url.LongURL)

	// Validate if shortURL already exist
	var result []myURL
	// Run query
	err = db.C(collection).Find(bson.M{"longURL": url.LongURL}).All(&result)
	if err != nil {
		log.Println("Error running query")
		log.Fatal(err)
		// Respond with an error
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	if len(result) == 0 {
		// No results
		log.Println("ShortURL doesn't exist for", url.LongURL, "creating")
		// Encode a hashID
		data := hashids.NewData()
		hash, _ := hashids.NewWithData(data)
		id, _ := hash.Encode([]int{int(time.Now().Unix())})
		// Create the short URL
		url.ID = id
		url.ShortURL = endpoint + "/" + id
		// Insert to DB
		err = db.C("shorty").Insert(url)
		if err != nil {
			log.Fatal(err)
			// Respond with an error
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		json.NewEncoder(res).Encode(url)
	} else {
		log.Println("ShortURL", result[0].ID, "already exists for", url.LongURL)
		// Respond with the URL
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(result[0])
	}
}

/** Handler to expand a short URL **/
func expandHandler(res http.ResponseWriter, req *http.Request) {
	var url myURL
	// Extract JSON from body
	err := json.NewDecoder(req.Body).Decode(&url)
	if err != nil {
		log.Fatal(err)
		// Respond with an error
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}
	log.Println("Received request to expand URL", url.ShortURL)
	// Find the corresponding long URL
	var result []myURL
	// Run query
	err = db.C(collection).Find(bson.M{"shortURL": url.ShortURL}).All(&result)
	if err != nil {
		log.Println("Error running query")
		log.Fatal(err)
		// Respond with an error
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	if len(result) == 0 {
		// Not a valid short URL
		log.Println("No longURL found for:", url.ShortURL)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(myURL{LongURL: "Invalid short URL, shorten first."})
	} else {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(result[0])
	}

}

/** Handler to redirect from a short URL **/
func redirectHandler(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id := params["id"]

	log.Println("Redirecting from:", id)
	var result []myURL
	err := db.C(collection).Find(bson.M{"id": id}).All(&result)
	if err != nil {
		log.Println("Error running query")
		log.Fatal(err)
		// Respond with an error
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	if len(result) == 0 {
		log.Println("ShortURL not found for ID:", id)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode("Impossible to redirect. URL not found.")
	} else {
		log.Println("Redirecting to:", result[0].LongURL)
		http.Redirect(res, req, result[0].LongURL, 301)
	}

}
