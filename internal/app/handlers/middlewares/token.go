package middlewares

import (
	"fmt"
	"net/http"
)

// 1 Проверяем что у пользователя нет куки
// 2. Если нет, генерим для его укникальный ключ и токен по нему
// 3 сохраняем пару ключ-токен и выставляем куку

// CookieTokenName define cookie name
const CookieTokenName = "user_token"

// TokenHandle check and set user token
func TokenHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(CookieTokenName); err == nil {
			fmt.Println("from cookie")
			fmt.Println(cookie)
			//if data, err2 := Decode(s, key, cookie.Value); err2 == nil {
			//	return data
			//}
		} else {
			// Set user cookie
			cookie := http.Cookie{
				Name:  CookieTokenName,
				Value: "abcd",
			}
			println("Set cookie")
			http.SetCookie(w, &cookie)
		}

		next.ServeHTTP(w, r)
	})
}
