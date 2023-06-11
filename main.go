package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type City struct {
	ID          int    `json:"ID,omitempty" db:"ID"`
	Name        string `json:"name,omitempty" db:"Name"`
	CountryCode string `json:"cuntryCode,omitempty" db:"CountryCode"`
	District    string `json:"district,omitempty" db:"District"`
	Population  int    `json:"population" db:"Population"`
}

func dbConnect() *sqlx.DB {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}

	conf := mysql.Config{
		User:      os.Getenv("DB_USERNAME"),
		Passwd:    os.Getenv("DB_PASSWORD"),
		Net:       "tcp",
		Addr:      os.Getenv("DB_HOSTNAME") + ":" + os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_DATABASE"),
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}

	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected")
	return db
}

func main() {
	e := echo.New()

	e.GET("/cities/:cityName", getCityInfoHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

func getCityInfoHandler(c echo.Context) error {
	cityName := c.Param("cityName")
	fmt.Println(cityName)
	var response City
	db := dbConnect()
	err := db.Get(&response, "SELECT * FROM city WHERE Name = ?", cityName)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("No such city Name = %s", cityName))
	} else if err != nil {
		log.Fatalf("DB error: %s", err)
	}

	return c.JSON(http.StatusOK, response)
}
