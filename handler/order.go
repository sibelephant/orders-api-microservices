package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an Order")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all Orders")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get an order by id")
}
func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("update an order by id")
}
func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete an order by id")
}
