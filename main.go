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

type Country struct {
	Code           string  `json:"code" db:"Code"`
	Name           string  `json:"name" db:"Name"`
	Continent      string  `json:"continent" db:"Continent"`
	Region         string  `json:"region" db:"Region"`
	SurfaceArea    float32 `json:"surfaceArea" db:"SurfaceArea"`
	IndepYear      int     `json:"indepYear" db:"IndepYear"`
	Population     int     `json:"population" db:"Population"`
	LifeExpectancy float32 `json:"lifeExpectancy" db:"LifeExpectancy"`
	GNP            float32 `json:"gnp" db:"GNP"`
	GNPOld         float32 `json:"gnpOld" db:"GNPOld"`
	LocalName      string  `json:"localName" db:"LocalName"`
	GovernmentForm string  `json:"governmentForm" db:"GovernmentForm"`
	HeadOfState    string  `json:"headOfState" db:"HeadOfState"`
	Capital        int     `json:"capital" db:"Capital"`
	Code2          string  `json:"code2" db:"Code2"`
	CapitalCity    City    `json:"capitalCity"`
}

type City struct {
	ID          int    `json:"ID,omitempty" db:"ID"`
	Name        string `json:"name,omitempty" db:"Name"`
	CountryCode string `json:"countryCode,omitempty" db:"CountryCode"`
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

	e.GET("/cities", getAllCityHandler)
	e.GET("/cities/:cityName", getCityInfoHandler)
	e.GET("/countries/:countryName", getCountryInfoHandler)
	e.POST("/addcity", addCityHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

func getAllCityHandler(c echo.Context) error {
	db := dbConnect()
	var cities []City
	if err := db.Select(&cities, "SELECT * FROM city"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%+v", err))
	}
	return c.JSON(http.StatusOK, &cities)
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

	return c.JSON(http.StatusOK, &response)
}

func getCountryInfoHandler(c echo.Context) error {
	countryName := c.Param("countryName")
	var country Country

	db := dbConnect()
	err := db.Get(&country, "SELECT * FROM country WHERE Name = ?", countryName)
	if errors.Is(err, sql.ErrNoRows) {
		return c.String(http.StatusBadRequest, fmt.Sprintf("No such country Name = %s", countryName))
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%+v", err))
	}

	if country.Capital != 0 {
		db.Get(&country.CapitalCity, "SELECT * FROM city WHERE ID = ?", country.Capital)
	}
	return c.JSON(http.StatusOK, &country)
}

func addCityHandler(c echo.Context) error {
	var newCity City
	if err := c.Bind(&newCity); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	fmt.Printf("%+v", newCity)
	db := dbConnect()
	_, err := db.NamedExec(`INSERT INTO city (Name, CountryCode, District, Population) VALUES (:Name, :CountryCode, :District, :Population)`, newCity)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%+v", err))
	}
	return c.JSON(http.StatusOK, &newCity)
}
