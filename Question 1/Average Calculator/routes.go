package main

import (
	"github.com/labstack/echo/v4"
)

func RoutesInit(e echo.Echo) {
	e.GET("/numbers/:numberid", numbersHandler)
}
