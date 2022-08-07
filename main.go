package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Vehicle struct {
	Id int
	Make string
	Model string
	Price int
}

func (v Vehicle) flatten() string {
	value := fmt.Sprintf("%d,%s,%s,%d\n", v.Id, v.Make, v.Model, v.Price)
	return value
}

var vehicles = []Vehicle{}

func initializeVehicles(vehicles *[]Vehicle, fileName string) {
	*vehicles = nil
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// advances to the next token, which is newlines
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")
		var car Vehicle
		car.Id, err = strconv.Atoi(s[0])
		car.Make = s[1]
		car.Model = s[2]
		car.Price, err = strconv.Atoi(s[3])

		*vehicles = append(*vehicles, car)
	}
}

func returnAllCars(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func returnCarsByBrand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carM := vars["make"]
	cars := &[]Vehicle{}
	for _, car := range vehicles {
		if car.Make == carM {
			*cars = append(*cars, car)
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cars)
}

func returnCarById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carId, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Unable to convert to string")
	}
	for _, car := range vehicles {
		if car.Id == carId {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(car)
		}
	}
}

func updateCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carId, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Unable to convert to string")
	}
	var updatedCar Vehicle
	json.NewDecoder(r.Body).Decode(&updatedCar)

	for i, car := range vehicles {
		if car.Id == carId {
			vehicles = append(vehicles[:i], vehicles[i+1:]...) // remove the car
			vehicles = append(vehicles[:i], updatedCar) // adding the updated car
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func createCar(w http.ResponseWriter, r *http.Request) {
	// O_WRONLY		set readonly
	// os.O_CREATE	create file if it doenst exist
	// os.O_APPEND	append to the file
	file, err := os.OpenFile("vehicles.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var newCar Vehicle
	json.NewDecoder(r.Body).Decode(&newCar)
	// add car to file
	if _, err := file.WriteString(newCar.flatten()); err != nil {
		log.Println(err)
	}

	// update state in this by getting the updated dat from the database (vehicles.txt)
	initializeVehicles(&vehicles, "vehicles.txt")

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func removeCarByIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	carId, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println("Unable to convert to string")
	}
	for i, car := range vehicles {
		if car.Id == carId {
			vehicles = append(vehicles[:i], vehicles[i+1:]...)
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func main() {
	initializeVehicles(&vehicles, "vehicles.txt")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/cars", returnAllCars).Methods("GET")
	router.HandleFunc("/cars/make/{make}", returnCarsByBrand).Methods("GET")
	router.HandleFunc("/cars/{id}", returnCarById).Methods("GET")
	router.HandleFunc("/cars/{id}", updateCar).Methods("PUT")
	router.HandleFunc("/cars", createCar).Methods("POST")
	router.HandleFunc("/cars/{id}", removeCarByIndex).Methods("DELETE")

	fmt.Println("Listening on 8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
