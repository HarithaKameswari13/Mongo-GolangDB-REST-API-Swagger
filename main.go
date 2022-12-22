package main

import (
	"log"

	_ "myapp/docs"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"
)

// @title ticket
// @version 1.0
// @description This is a sample service for ticket Implementation
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email soberkoder@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9000
// @BasePath /
func main() {
	s := mux.NewRouter()
	// Routes
	s.HandleFunc("/Createticket", createTicket).Methods("POST")
	s.HandleFunc("/Getticket/{id}", getOneTicket).Methods("GET")
	s.HandleFunc("/GetAllTickets", getAllTickets).Methods("GET")
	s.HandleFunc("/Updateticket/{id}", updateTicket).Methods("PUT")
	s.HandleFunc("/GetMyTickets", getMyTickets).Methods("GET")

	s.HandleFunc("/CreateUser", createUser).Methods("POST")
	s.HandleFunc("/GetUser/{id}", getOneUser).Methods("GET")
	s.HandleFunc("/GetAllUsers", getAllUsers).Methods("GET")
	s.HandleFunc("/UpdateUser", updateUser).Methods("PUT")
	s.HandleFunc("/DeleteUser/{id}", deleteUser).Methods("DELETE")
	//s.HandleFunc("/Welcome", Welcome).Methods("GET")

	s.HandleFunc("/Signin", SignIn).Methods("POST")

	s.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":9000", s)) // Run Server
	}()
	Signin_main()

}
