package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var expireTime time.Time
var Usercollection = Db().Database("ticket_management").Collection("users")     // get collection users from db which will return a mongo client
var Ticketcollection = Db().Database("ticket_management").Collection("tickets") // get collection tickets from Db which will return a mongo client
var Femail, Fpwd int

//var count int64

type User struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	UserName  string `json:"username"`
	AuthLevel string `json:"authlevel"`
	Password  string `json:"password"`
}
type Ticket struct {
	ID                 string `bson:"_id,omitempty" json:"id"`
	Ticket_Description string `json:"description"`
	Ticket_Status      string `json:"status"`
	Client_response    string `json:"Client_response"`
	Created_by         string `json:"by"`
}
type Credentials struct {
	Password string `json:"password"`
	EmailId  string `json:"id"`
}
type Claims struct {
	EmailId string `json:"id"`
	jwt.StandardClaims
}

var jwtKey = []byte("my_secret_key")

func Signin(w http.ResponseWriter, r *http.Request) {
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
	password := CheckPassword(result1, creds.Password)

	ifTrue := CheckPasswordHash(creds.Password, password)
	if !ifTrue {
		w.WriteHeader(http.StatusUnauthorized)
		//fmt.Println("wrong password!!try again")
		json.NewEncoder(w).Encode("Incorrect Password")
		return
	}
	expirationTime := time.Now().Add(5 * time.Minute)
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
	entryPoint(result1, creds.EmailId)
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckPassword(data string, password string) string {
	re, _ := regexp.Compile("password")
	match := re.FindStringIndex(data)
	str := ""
	for i := match[1] + 1; true; i++ {
		if string(data[i]) == " " || string(data[i]) == "]" {
			break
		}
		str = str + string(data[i])
	}
	return str
}

func GetAuthLevel(data string) string {
	re, _ := regexp.Compile("authlevel")
	match := re.FindStringIndex(data)
	str := ""
	for i := match[1] + 1; true; i++ {
		if string(data[i]) == " " || string(data[i]) == "]" {
			break
		}
		str = str + string(data[i])
	}
	return str
}

func entryPoint(content, emailId string) {
	fmt.Println("Sign-in success!!")
	re, _ := regexp.Compile("authlevel")
	match := re.FindStringIndex(content)
	level := ""
	for i := match[1] + 1; true; i++ {
		if string(content[i]) == " " {
			break
		}
		level = level + string(content[i])
	}
	if strings.Compare(level, "admin") == 0 { //all 5 == administrtation
		adminFunctions(emailId)
	} else if strings.Compare(level, "client") == 0 { //update
		clientFunctions(emailId)
	} else if strings.Compare(level, "adminteam") == 0 { //except status
		adminteamFunctions(emailId)
	} else {
		fmt.Println("You do not have access")
	}
}
func Uniqueid() string {
	count, _ := Ticketcollection.CountDocuments(context.TODO(), bson.D{})
	val := strconv.Itoa(int(count + 1))
	var prefix string = "INC"
	for i := 0; i < (4 - len(val)); i++ {
		prefix = prefix + "0"
	}
	return prefix + val
}
func sendMail(message string, id []string) {
	from := "ash.dummy901@gmail.com"
	password := "Account@123"
	host := "smtp.gmail.com"
	port := "587"
	body := []byte(message)
	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, id, body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Mail sent")
}
func getOneticket() (int, string) {
	fmt.Print("Enter the ticket id: ")
	var ticket_id string
	fmt.Scanln(&ticket_id)
	var result primitive.M
	err := Ticketcollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: ticket_id}}).Decode(&result)
	errtype := fmt.Sprint(err) //converting err to string type
	opt := strings.Compare(errtype, "mongo: no documents in result")
	return opt, ticket_id
}
func createticket(emailId string) {
	var ticketbody Ticket
	fmt.Printf("Enter ticket description: ")
	reader := bufio.NewReader(os.Stdin)
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimRight(desc, "\r\n")
	ticketbody.Ticket_Description = desc
	ticketbody.Created_by = emailId
	ticketbody.ID = Uniqueid()
	ticketbody.Ticket_Status = "Open"
	ticketbody.Client_response = "Started"
	bson.Marshal(&ticketbody)
	insertResult, _ := Ticketcollection.InsertOne(context.TODO(), ticketbody)
	fmt.Println(insertResult)
}
func printTickets(status string) {
	cur, err := Ticketcollection.Find(context.TODO(), bson.D{{Key: "ticket_status", Value: status}})
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(elem)
	}
}

func adminteamFunctions(emailId string) {
	var option int = 0
forLoop:
	for option != 4 {
		fmt.Println("Select the operation:\n1. Create ticket\n2. Update ticket\n3. show my ticket\n4. Signout")
		fmt.Scanln(&option)
		if time.Now().Minute() < expireTime.Minute() || time.Now().Second() <= expireTime.Second() {
			switch option {
			case 1:
				createticket(emailId)
			case 2:
				found, ticket_id := getOneticket()
				if found == 0 {
					fmt.Println("Ticket Id not found in database")
				} else {
					fmt.Print("Update description: ")
					reader := bufio.NewReader(os.Stdin)
					desc, _ := reader.ReadString('\n')
					desc = strings.TrimRight(desc, "\r\n")
					filter := bson.D{{Key: "_id", Value: ticket_id}} // converting value to BSON type
					after := options.After                           // returns updated document
					returnOpt := options.FindOneAndUpdateOptions{
						ReturnDocument: &after,
					}
					update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_description", Value: desc}}}} //updated in database
					_ = Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
					fmt.Println("description updated successfully")
				}
			case 3:
				cur, err := Ticketcollection.Find(context.TODO(), bson.D{{Key: "created_by", Value: emailId}})
				if err != nil {
					fmt.Println(err)
				}
				for cur.Next(context.TODO()) { //Next() gets the next document
					var elem primitive.M
					err := cur.Decode(&elem)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println("Results", elem)
				}
			case 4:
				fmt.Println("Signed out successfully")
				break forLoop
			default:
				fmt.Println("Invalid try again!!!")
			}
		} else {
			fmt.Println("Session Expired")
			break forLoop
		}
	}
}
func clientFunctions(emailId string) {
	printTickets("Open")
	printTickets("Pending")
	var ticketId string
	var option int = 0
forLoop:
	for option != 2 {
		fmt.Println("Enter the options\n1.Respond to ticket\n2.Signout")
		fmt.Scanln(&option)
		if time.Now().Minute() < expireTime.Minute() || time.Now().Second() <= expireTime.Second() {
			switch option {
			case 1:
				fmt.Print("Which ticket do you want to respond: ")
				fmt.Scanln(&ticketId)
				var result primitive.M
				err := Ticketcollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: ticketId}}).Decode(&result)
				errtype := fmt.Sprint(err)
				//fmt.Println(errtype, result)
				if strings.Compare(errtype, "mongo: no documents in result") == 0 {
					fmt.Println("Ticket Id not found in database")
				} else {
					filter := bson.D{{Key: "_id", Value: ticketId}}
					after := options.After
					returnOpt := options.FindOneAndUpdateOptions{
						ReturnDocument: &after,
					}
					fmt.Print("Type your response: ")
					reader := bufio.NewReader(os.Stdin)
					client_response, _ := reader.ReadString('\n')
					client_response = strings.TrimRight(client_response, "\r\n")
					// converting value to BSON type
					var op string
					update := bson.D{{Key: "$set", Value: bson.D{{Key: "client_response", Value: client_response}}}} //updated in database
					_ = Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
					fmt.Println("Do you want to change the ticket status to Pending(Y/N)?")
					fmt.Scanln(&op)
					if op == "y" || op == "Y" {
						update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_status", Value: "Pending"}}}} //updated in database
						_ = Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
						sendMail(client_response, []string{result["created_by"].(string)})
					} else {
						update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_status", Value: "Open"}}}} //updated in database
						_ = Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
					}
					fmt.Println("client_response field updated successfully")
					sendMail(client_response, []string{result["created_by"].(string)})

				}

			case 2:
				fmt.Println("Signed out!!!")
				break forLoop
			default:
				fmt.Println("Invalid option")
			}
		} else {
			fmt.Println("Session Expired")
			break forLoop

		}

	}
}
func adminFunctions(emailid string) {
	var option int = 0
forLoop:
	for option != 7 {
		fmt.Println("1. Create ticket\n2.Get All Tickets\n3. Get One Ticket\n4. Get all tickets that are open and pending\n5. Close the Ticket\n6. Delete User\n7. Signout")
		fmt.Scanln(&option)
		if time.Now().Minute() < expireTime.Minute() || time.Now().Second() <= expireTime.Second() {
			switch option {
			case 1:
				createticket(emailid)
			case 2:
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
					fmt.Println(elem)
				}
			case 3:
				found, ticket_id := getOneticket()
				if found == 0 {
					fmt.Println("Ticket Id not found in database")
				} else {
					var result primitive.M
					_ = Ticketcollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: ticket_id}}).Decode(&result)
					fmt.Println(result)
				}
			case 4:
				printTickets("Open")
				printTickets("Pending")
			case 5:
				var ticketId string
				fmt.Print("Which ticket do you want to close: ")
				fmt.Scanln(&ticketId)
				var result primitive.M
				err := Ticketcollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: ticketId}}).Decode(&result)
				errtype := fmt.Sprint(err) //converting err to string type
				//fmt.Println(errtype, result)
				if strings.Compare(errtype, "mongo: no documents in result") == 0 {
					fmt.Println("Ticket Id not found in database")
				} else {
					filter := bson.D{{Key: "_id", Value: ticketId}}
					after := options.After
					returnOpt := options.FindOneAndUpdateOptions{
						ReturnDocument: &after,
					}
					// converting value to BSON type
					update := bson.D{{Key: "$set", Value: bson.D{{Key: "ticket_status", Value: "closed"}}}} //updated in database
					_ = Ticketcollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)
					sendMail(ticketId+" ticket closed", []string{result["created_by"].(string)})
					fmt.Println("client_response field updated successfully/n")
					fmt.Println("Ticket closed successfully")

				}
			case 6:
				var emailid string
				var result primitive.M
				fmt.Print("Which user do you want to delete: ")
				fmt.Scanln(&emailid)
				err := Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: emailid}}).Decode(&result)
				errtype := fmt.Sprint(err) //converting err to string type

				if strings.Compare(errtype, "mongo: no documents in result") == 0 {
					fmt.Println("User not found in database")
				} else {
					opts := options.Delete().SetCollation(&options.Collation{}) // to specify language-specific rules for string comparison, such as rules for lettercase
					res, err := Usercollection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: emailid}}, opts)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("deleted %v user\n", res.DeletedCount)
				}
			case 7:
				fmt.Println("Exit")
				break forLoop
			default:
				fmt.Println("invalid option")
			}
		} else {
			fmt.Println("Session Expired")
			break forLoop

		}
	}
}
