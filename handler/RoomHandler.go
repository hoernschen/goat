package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoernschen/goat/models"
	"github.com/hoernschen/goat/repository"
)

func GetRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(repository.GetRooms()); err != nil {
		panic(err)
	}
}

func GetRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(repository.GetRoom(params["id"])); err != nil {
		panic(err)
	}
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var room models.Room

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.Unmarshal(b, &room)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	repository.CreateRoom(room)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(repository.GetRooms()); err != nil {
		panic(err)
	}
}

func JoinRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var user models.User

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.Unmarshal(b, &user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	repository.JoinRoom(params["id"], user)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func LeaveRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var user models.User

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.Unmarshal(b, &user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	repository.LeaveRoom(params["id"], user.ID)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func DeleteRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	repository.DeleteRoom(params["id"])
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
