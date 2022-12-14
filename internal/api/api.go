package api

import (
	"context"
	"encoding/json"
	"fmt"
	"l0/internal/models"
	"l0/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/nats-io/stan.go"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) Handler {
	return Handler{svc: svc}
}

//Метод сервиса для обработки POST запроса и возвращения ответа в виде json данных по заказу по id
func (h Handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqbody := struct {
		Id int `json:"id"`
	}{}

	// data, _ := io.ReadAll(r.Body)
	// json.Unmarshal(data, &reqbody)

	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.svc.GetOrderByID(r.Context(), reqbody.Id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respbody, err := json.Marshal(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(respbody)
}

//Метод для обработки полученной формы и последующим выводом на экран данных по запрошенному id
func (h Handler) IdOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var id int

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		log.Fatal(err)
	}

	order, err := h.svc.GetOrderByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respbody, err := json.Marshal(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", respbody)
}

//Метод обработки данных из канала Nats-streaming. Создаёт новый заказ в бд.
func (h Handler) CreateOrder(msg *stan.Msg) {
	var order models.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Println(err)
	}
	if _, err := h.svc.CreateOrder(context.Background(), order); err != nil {
		log.Println(err)
	}
}
