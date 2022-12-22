package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mailid string

// SignIn entity
// SignIn godoc
// @Summary signin
// @Description Signin
// @Tags Login
// @Accept  json
// @Produce  json
// @Param creds body Credentials true "Enter your email Id and password"
// @Success 200
// @failure 401
// @Router /Signin [post]
func SignIn(w http.ResponseWriter, r *http.Request) {

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var result primitive.M //---->bson
	err = Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: creds.EmailId}}).Decode(&result)
	errtype := fmt.Sprint(err) //converting err to string type
	Femail = strings.Compare(errtype, "mongo: no documents in result")
	if Femail == 0 {
		//fmt.Println("Email id not found in database")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Email id not found in database")
		return
	}
	if err != nil {
		fmt.Println(err)
	}
	result1 := fmt.Sprint(result)
	//creds.Password, _ = HashPassword(creds.Password)
	password := CheckPassword(result1, creds.Password)
	ifTrue := CheckPasswordHash(creds.Password, password)

	if !ifTrue {
		w.WriteHeader(http.StatusUnauthorized)
		//fmt.Println("wrong password!!try again")
		json.NewEncoder(w).Encode("Incorrect Password")
		return
	}
	mailid = creds.EmailId
	expirationTime := time.Now().Add(20 * time.Minute)
	expireTime = expirationTime
	claims := &Claims{
		EmailId: creds.EmailId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Sign-in success!!")

}
func IsAuthorized(w http.ResponseWriter, r *http.Request) (bool, string, string) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return false, "", ""
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return false, "", ""
	}

	// Get the JWT string from the cookie
	tknStr := c.Value
	// Initialize a new instance of `Claims`
	claims := &Claims{}
	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return false, "", ""
		}
		w.WriteHeader(http.StatusBadRequest)
		return false, "", ""
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return false, "", ""
	}
	// Finally, return the welcome message to the user, along with their
	// username given in the token

	var result primitive.M //  an unordered representation of a BSON document which is a Map
	err = Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: claims.EmailId}}).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	result1 := fmt.Sprint(result)
	fmt.Println(GetAuthLevel(result1), claims.EmailId)
	return true, GetAuthLevel(result1), claims.EmailId
}

// createUser entity
// createuser godoc
// @Summary create a new user
// @Description Signup
// @Tags User
// @Accept  json
// @Produce  json
// @Param incident body User true "create a new user"
// @Success 200
// @failure 401
// @Router /CreateUser [post]
func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // adding content type

	var incident User
	err := json.NewDecoder(r.Body).Decode(&incident) // storing incident variable of type User
	if err != nil {
		fmt.Println("Error", err)
	}
	incident.Password, _ = HashPassword(incident.Password)
	insertResult, err := Usercollection.InsertOne(context.TODO(), incident)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Inserted a single document:", insertResult)
	json.NewEncoder(w).Encode(insertResult.InsertedID) // return mongoDB ID of generated document
}

// getOneUser entity
// getOneUser godoc
// @Summary get one user
// @Description getoneuser
// @Tags User
// @Accept  json
// @Produce  json
// @Param id path string true "Enter email id"
// @Success 200
// @failure 401
// @Router /GetUser/{id} [get]
func getOneUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)["id"]
	var user User
	e := json.NewDecoder(r.Body).Decode(&user)
	if e != nil {
		fmt.Print(e)
	}
	var result primitive.M //  an unordered representation of a BSON document which is a Map
	err := Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: params}}).Decode(&result)

	if err != nil {

		fmt.Println(err)

	}

	json.NewEncoder(w).Encode(result) // returns a Map containing document

}

// getAllUsers entity
// getAllUsers godoc
// @Summary get all users
// @Description get all user
// @Tags User
// @Accept  json
// @Produce  json
// @Success 200
// @failure 401
// @Router /GetAllUsers [get]
func getAllUsers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var results []primitive.M                                   //slice for multiple documents
	cur, err := Usercollection.Find(context.TODO(), bson.D{{}}) //returns a *mongo.Cursor
	if err != nil {

		fmt.Println(err)

	}
	for cur.Next(context.TODO()) { //Next() gets the next document for corresponding cursor
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem) // appending document pointed by Next()
		//fmt.Println("Results", results)
	}
	//fmt.Println("no of documents: ", count)
	cur.Close(context.TODO()) // close the cursor once stream of documents has exhausted
	json.NewEncoder(w).Encode(results)
}

// deleteUser entity
// deleteUser godoc
// @Summary delete user
// @Description delete user
// @Tags User
// @Accept  json
// @Produce  json
// @Param id path string true "Enter emailid"
// @Success 200
// @failure 401
// @Router /DeleteUser/{id} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {

	ok, authlevel, _ := IsAuthorized(w, r)
	if !ok || authlevel != "admin" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)["id"]                                 //get Parameter value as string
	opts := options.Delete().SetCollation(&options.Collation{}) // to specify language-specific rules for string comparison, such as rules for lettercase
	res, err := Usercollection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: params}}, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %v documents\n", res.DeletedCount)
	json.NewEncoder(w).Encode(res.DeletedCount) // return number of documents deleted

}

func updateUser(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("Check 1 came inside update User")
	w.Header().Set("Content-Type", "application/json")

	type updateBody struct {
		ID        string `json:"id" bson:"_id,omitempty"`
		UserName  string `json:"username"`
		Authlevel string `json:"authlevel" bson:"authlevel"`
	}
	var body updateBody
	e := json.NewDecoder(r.Body).Decode(&body)
	if e != nil {
		fmt.Print(e)
	}
	filter := bson.D{{Key: "_id", Value: body.ID}} // converting value to BSON type
	after := options.After                         // for returning updated document
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "username", Value: body.UserName}, {Key: "authlevel", Value: body.Authlevel}}}}
	updateResult := Usercollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
	var result primitive.M
	_ = updateResult.Decode(&result)
	//fmt.Println("Result", result)
	json.NewEncoder(w).Encode(result)
}

type ticketdesc struct {
	Ticket_Description string `json:"description"`
}

// createTicket entity
// createTicket godoc
// @Summary create ticket
// @Description create ticket
// @Tags Ticket
// @Accept  json
// @Produce  json
// @Param TD body ticketdesc true "Enter the ticket description"
// @Success 200
// @failure 401
// @Router /Createticket [post]
func createTicket(w http.ResponseWriter, r *http.Request) {
	ok, authlevel, _ := IsAuthorized(w, r)
	c1 := authlevel == "admin"
	c2 := authlevel == "adminteam"
	if !(ok && (c1 || c2)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json") // adding content type

	var incident Ticket

	count, _ := Ticketcollection.CountDocuments(context.TODO(), bson.D{})
	val := strconv.Itoa(int(count + 1))
	var prefix string = "INC"

	for i := 0; i < (4 - len(val)); i++ {

		prefix = prefix + "0"

	}
	//fmt.Println(len(val), prefix, prefix+val)
	incident.ID = prefix + val
	incident.Ticket_Status = "Open"
	incident.Client_response = "Started"
	incident.Created_by = mailid
	err := json.NewDecoder(r.Body).Decode(&incident) // storing incident variable of type Ticket
	if err != nil {
		fmt.Println("Error", err)
	}
	insertResult, err := Ticketcollection.InsertOne(context.TODO(), incident)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Inserted a single document:", insertResult)
	json.NewEncoder(w).Encode(insertResult.InsertedID)

}

// getOneTicket entity
// getOneTicket godoc
// @Summary get one ticket
// @Description get one ticket
// @Tags Ticket
// @Accept  json
// @Produce  json
// @Param id path string true "Enter Ticketid"
// @Success 200
// @failure 401
// @Router /Getticket/{id} [get]
func getOneTicket(w http.ResponseWriter, r *http.Request) {
	ok, authlevel, _ := IsAuthorized(w, r)
	if !ok || authlevel != "admin" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var body Ticket
	params := mux.Vars(r)["id"]
	e := json.NewDecoder(r.Body).Decode(&body)
	if e != nil {
		fmt.Print(e)
	}
	var result primitive.M
	err := Ticketcollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: params}}).Decode(&result)
	if err != nil {

		fmt.Println(err)

	}
	json.NewEncoder(w).Encode(result) // returns a Map containing document

}

// getAllTicket entity
// getAllTicket godoc
// @Summary get All Ticket
// @Description get All Ticket
// @Tags Ticket
// @Accept  json
// @Produce json
// @Success 200
// @failure 401
// @Router /GetAllTickets [get]
func getAllTickets(w http.ResponseWriter, r *http.Request) {
	ok, authlevel, _ := IsAuthorized(w, r)
	c1 := authlevel == "admin"
	c2 := authlevel == "client"
	if !(ok && (c1 || c2)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var results []primitive.M
	cur, err := Ticketcollection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) { //Next() gets the next document
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem) // appending document pointed by Next()
		//fmt.Println("Results", results)
	}
	cur.Close(context.TODO())
	json.NewEncoder(w).Encode(results)
}

// getAllTicket entity
// getAllTicket godoc
// @Summary get All Ticket
// @Description get All Ticket
// @Tags Ticket
// @Accept  json
// @Produce  json
// @Success 200
// @failure 401
// @Router /GetMyTickets [get]
func getMyTickets(w http.ResponseWriter, r *http.Request) {
	ok, authlevel, _ := IsAuthorized(w, r)
	c1 := authlevel == "admin"
	c2 := authlevel == "adminteam"
	if !(ok && (c1 || c2)) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var results []primitive.M
	cur, err := Ticketcollection.Find(context.TODO(), bson.D{{Key: "created_by", Value: mailid}})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) { //Next() gets the next document
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem) // appending document pointed by Next()
		//fmt.Println("Results", results)
	}
	cur.Close(context.TODO())
	json.NewEncoder(w).Encode(results)
}

type updateBody struct {
	Ticket_Description string `json:"description"`
	Client_response    string `json:"client_response"`
	Ticket_Status      string `json:"ticket_status"`
}

// updateTicket entity
// updateTicket godoc
// @Summary update ticket
// @Description update ticket
// @Tags Ticket
// @Accept  json
// @Produce  json
// @Param id path string true "Enter Ticketid"
// @Param body body updateBody false "[Client - respond to ticket or change the ticket status]  [Admin - Close the ticket]  [Admin Team - Update the description]"
// @Success 200
// @failure 401
// @Router /Updateticket/{id} [put]
func updateTicket(w http.ResponseWriter, r *http.Request) {
	ok, authlevel, _ := IsAuthorized(w, r)
	if !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("you have no access to this function"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)["id"]
	filter := bson.D{{Key: "_id", Value: params}} // converting value to BSON type
	after := options.After                        // returns updated document
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	//fmt.Println("Check 1 came inside update Ticket")

	if authlevel == "adminteam" {
		var body updateBody
		e := json.NewDecoder(r.Body).Decode(&body)
		if e != nil {
			fmt.Print(e)
		}
		filter = bson.D{{Key: "_id", Value: params}, {Key: "created_by", Value: mailid}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_description", Value: body.Ticket_Description}}}}
		updateResult := Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
		var result primitive.M
		_ = updateResult.Decode(&result)
		if result == nil {
			json.NewEncoder(w).Encode("This is not your ticket-id")
		} else {
			json.NewEncoder(w).Encode(result)
		}
	} else if authlevel == "client" {
		var body updateBody
		e := json.NewDecoder(r.Body).Decode(&body)
		if e != nil {
			fmt.Print(e)
		}

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "client_response", Value: body.Client_response}, {Key: "ticket_status", Value: body.Ticket_Status}}}}
		updateResult := Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
		var result primitive.M
		_ = updateResult.Decode(&result)
		json.NewEncoder(w).Encode(result)
	} else {
		var body updateBody
		e := json.NewDecoder(r.Body).Decode(&body)
		if e != nil {
			fmt.Print(e)
		}

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_status", Value: body.Ticket_Status}}}}
		updateResult := Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
		var result primitive.M
		_ = updateResult.Decode(&result)
		json.NewEncoder(w).Encode(result)
	}

	//fmt.Println("Result", result)

}
