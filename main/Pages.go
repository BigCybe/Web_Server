package main

import (
	"github.com/labstack/echo"
	"net/http"
)

func MainPage(c echo.Context) error {
	return c.Render(http.StatusOK, "MainPage.html", echo.Map{"title": "Page file title!!"})
}

func MainPageTest(c echo.Context) error {
	return c.Render(http.StatusOK, "MainPageTest.html", echo.Map{"title": "Page file title!!"})
}

func TokenPage(c echo.Context) error {
	return c.Render(http.StatusOK, "TokenPage.html", echo.Map{"title": "Page file title!!"})
}

func GroupPage(c echo.Context) error {
	return c.Render(http.StatusOK, "GroupPage.html", echo.Map{"title": "Page file title!!"})
}

func TaskPage(c echo.Context) error {
	return c.Render(http.StatusOK, "TaskPage.html", echo.Map{"title": "Page file title!!"})
}
