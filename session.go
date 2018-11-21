package main

import (
	"github.com/gorilla/securecookie"
	"net/http"
	"golang.org/x/crypto/bcrypt"

)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}


// logout handler
func checkLoginData(username string, password string) bool{
	bytePassword  := []byte{36, 50, 97, 36, 49, 48, 36, 109, 56, 102, 88, 54, 70, 66, 100, 86, 73, 116, 120, 102, 81, 109, 102, 80, 118, 103, 72, 50, 46, 46, 80, 57, 65, 81, 79, 90, 112, 122, 54, 111, 67, 73, 105, 81, 77, 82, 99,46, 88, 73, 77, 87, 71, 97, 89, 84, 51, 90, 65, 83}
	errCompare:=bcrypt.CompareHashAndPassword(bytePassword,[]byte(password))
	if username == "alena" && errCompare==nil{
		return true
	}
	return false
}

func checkLogin(w http.ResponseWriter, r *http.Request){
	if getUserName(r)==""{
		http.Redirect(w,r, "/login", 302)
	}
}

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/login", 302)
}
