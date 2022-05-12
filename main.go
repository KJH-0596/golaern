package main

import (
	"os"
	"strings"

	"github.com/KJH-0596/golaern/scrapper"
	"github.com/labstack/echo"
)

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

const fileName string = "jobs.csv"

func handlScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)
}


func main(){
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handlScrape)
	e.Logger.Fatal(e.Start(":1324"))
}