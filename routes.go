package main

import (
	"net/http"

	"github.com/hoernschen/goat/handler"
	"github.com/hoernschen/goat/handler/signaling"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{"WebSocket Connection", "GET", "/r/{room}", signaling.WSConnection},

	Route{"GetUsers", "GET", "/users", handler.GetUsers},
	Route{"GetUser", "GET", "/users/{id}", handler.GetUser},
	Route{"CreateUser", "POST", "/users", handler.CreateUser},
	Route{"DeleteUser", "DELETE", "/users/{id}", handler.DeleteUser},

	Route{"GetRooms", "GET", "/rooms", handler.GetRooms},
	Route{"GetRoom", "GET", "/rooms/{id}", handler.GetRoom},
	Route{"CreateRoom", "POST", "/rooms", handler.CreateRoom},
	Route{"JoinRoom", "PUT", "/rooms/{id}/join", handler.JoinRoom},
	Route{"LeaveRoom", "PUT", "/rooms/{id}/leave", handler.LeaveRoom},
	Route{"DeleteRoom", "DELETE", "/rooms/{id}", handler.DeleteRoom},
}
