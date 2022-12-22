package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Signin_main() {
	var options int = 0
	var flag int = 1
	var email, password, authlevel, username string
	for {
		fmt.Println("1.Signup 2.Signin")
		fmt.Scanln(&options)
		if options == 1 {
			flag = 1
			for flag == 1 {
				var userbody User
				fmt.Print("Enter emailid: ")
				fmt.Scanln(&email)
				var result primitive.M
				err := Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: email}}).Decode(&result)
				errtype := fmt.Sprint(err) //converting err to string type
				if strings.Compare(errtype, "mongo: no documents in result") != 0 {
					fmt.Println("emailid already exist")
					flag = 1
				} else {
					fmt.Print("enter password: ")
					fmt.Scanln(&password)
					fmt.Print("enter username: ")
					fmt.Scanln(&username)
					fmt.Print("enter auth level :")
					fmt.Scanln(&authlevel)
					userbody.ID = email
					userbody.AuthLevel = authlevel
					userbody.Password = password
					userbody.UserName = username
					bson.Marshal(&userbody)
					Usercollection.InsertOne(context.TODO(), userbody) //stored in db
					fmt.Println(email, "is created successfully!")
					flag = 0
				}

			}

		} else {
			flag = 1
			for flag == 1 {
				fmt.Printf("Enter your email: ")
				fmt.Scanln(&email)
				var result primitive.M //---->bson
				//emailid---to check if email id is there in db--using filter
				err := Usercollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: email}}).Decode(&result)
				errtype := fmt.Sprint(err) //converting err to string type
				Femail := strings.Compare(errtype, "mongo: no documents in result")
				if Femail == 0 {
					fmt.Println("Email id not found in database")
					flag = 1
				} else {
					if err != nil {

						fmt.Println(err)

					}
					fmt.Printf("Enter your password: ")
					fmt.Scanln(&password)
					result1 := fmt.Sprint(result)
					Fpwd := CheckPassword(result1, password)
					ifTrue := CheckPasswordHash(password, Fpwd)
					if !ifTrue {
						fmt.Println("wrong password!!try again")
						flag = 1
					} else {
						flag = 0
					}

				}

				values := map[string]string{"id": email, "password": password}
				json_data, err := json.Marshal(values)

				if err != nil {
					log.Fatal(err)
				}
				_, err = http.Post("http://localhost:9000/api/Signin", "application/json", bytes.NewBuffer(json_data))
				if err != nil {

					log.Fatal(err)

				}
			}
		}
	}
}
