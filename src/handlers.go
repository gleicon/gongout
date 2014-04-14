// Copyright 2013 v.chu.pe authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"regexp"
)

var roomNameRe = regexp.MustCompile("^([a-zA-Z0-9]+)$")

func RoomHandler(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Path[len("/r/"):]
	if !roomNameRe.MatchString(roomName) {
		http.Error(w, "Invalid room name", 400)
		return
	}

	rc, _ := Redis.HIncrBy("room:"+roomName, "access", 1)
	Room.ExecuteTemplate(w, "room.html", map[string]interface{}{"RoomName": roomName, "Counter": rc})
}

func RoomStatusHandler(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Path[len("/s/"):]
	if !roomNameRe.MatchString(roomName) {
		http.Error(w, "Invalid room name", 400)
		return
	}

	s, _ := Redis.HGetAll("room:" + roomName)
	Status.ExecuteTemplate(w, "status.html", map[string]interface{}{"RoomName": roomName, "Metadata": s})
	fmt.Fprintln(w, "Room ", roomName)
	fmt.Fprintln(w, "Status ", s)
}

func NewRoomHandler(w http.ResponseWriter, r *http.Request) {
	uuid := getUUID()
	eid := base62FromUUID(uuid)
	http.Redirect(w, r, "/r/"+eid, 302)
}
