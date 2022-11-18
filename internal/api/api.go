package api

import (
	"context"
	"encoding/json"
	"l0/internal/models"
	"l0/internal/service"
	"log"
	"net/http"

	"github.com/nats-io/stan.go"
)

type Handler struct {
	svc service.Service
}

func New(svc service.Service) Handler {
	return Handler{svc: svc}
}

func (h Handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqbody := struct {
		ID int `json:"id"`
	}{}

	// data, _ := io.ReadAll(r.Body)
	// json.Unmarshal(data, &reqbody)

	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.svc.GetOrderByID(r.Context(), reqbody.ID)
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

func (h Handler) CreateOrder(msg *stan.Msg) {
	var order models.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Println(err)
	}
	if _, err := h.svc.CreateOrder(context.Background(), order); err != nil {
		log.Println(err)
	}
}
