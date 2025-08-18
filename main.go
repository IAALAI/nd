package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var path *string
var port *string

// model
type Interface struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Alias     sql.NullString `json:"-"`
	Active    int            `json:"-"` // 反正只返回活跃的,这里不返回活跃状态了
	Created   string         `json:"created"`
	Updated   string         `json:"updated"`
	RxCounter int64          `json:"-"`
	TxCounter int64          `json:"-"`
	RxTotal   int64          `json:"rxtotal"`
	TxTotal   int64          `json:"txtotal"`
}

type dataCost struct {
	Date string `json:"date"`
	Rx   int64  `json:"rx"`
	Tx   int64  `json:"tx"`
}

func initFlag() {
	port = flag.String("port", "8080", "Port to listen on")
	path = flag.String("path", "./vnstat.db", "Path to serve files from")
	flag.Parse()
}

func ControlCacheMiddle() gin.HandlerFunc {
	h := func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=3600")
		c.Next()
	}
	return h
}

func main() {
	initFlag()

	db, err := sql.Open("sqlite3", "file:"+*path)
	if err != nil {
		fmt.Println("failed to open database:", err)
		os.Exit(1)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(ControlCacheMiddle())

	r.GET("/", func(c *gin.Context) {
		c.Header("cach", "")
		c.String(200, "Hello, World!\n")
	})

	r.GET("/t", func(c *gin.Context) {
		c.File("./static/t.html")
	})

	r.GET("/echarts.min.js", func(c *gin.Context) {
		c.File("./static/echarts.min.js")
	})

	r.GET("/tailwind.js", func(c *gin.Context) {
		c.File("./static/index.global.min.js")
	})

	r.GET("/info", func(c *gin.Context) {
		rows, err := db.Query("SELECT * from interface where active == 1;")
		if err != nil {
			c.String(500, "Error querying database: %v", err)
			return
		}
		defer rows.Close()
		var interfaces []Interface
		for rows.Next() {
			var iface Interface
			err := rows.Scan(&iface.ID, &iface.Name, &iface.Alias, &iface.Active, &iface.Created, &iface.Updated, &iface.RxCounter, &iface.TxCounter, &iface.RxTotal, &iface.TxTotal)
			if err != nil {
				c.String(500, "Error scanning row: %v", err)
				return
			}
			interfaces = append(interfaces, iface)
		}
		if err := rows.Err(); err != nil {
			c.String(500, "Error iterating rows: %v", err)
			return
		}
		c.JSON(200, interfaces)
	})

	r.GET("/h/:id", func(c *gin.Context) {
		id := c.Param("id")
		rows, err := db.Query("select date, tx, rx from hour where interface == ? order by id desc limit 24;", id)
		if err != nil {
			c.String(500, "Error querying database: %v", err)
			return
		}
		defer rows.Close()
		var data []dataCost
		for rows.Next() {
			var d dataCost
			err := rows.Scan(&d.Date, &d.Tx, &d.Rx)
			if err != nil {
				c.String(500, "Error scanning row: %v", err)
				return
			}
			data = append(data, d)
		}
		c.JSON(200, data)
	})

	r.GET("/d/:id", func(c *gin.Context) {
		id := c.Param("id")
		rows, err := db.Query("select date, tx, rx from day where interface == ? order by id desc limit 31;", id)
		if err != nil {
			c.String(500, "Error querying database: %v", err)
			return
		}
		defer rows.Close()
		var data []dataCost
		for rows.Next() {
			var d dataCost
			err := rows.Scan(&d.Date, &d.Tx, &d.Rx)
			if err != nil {
				c.String(500, "Error scanning row: %v", err)
				return
			}
			data = append(data, d)
		}
		c.JSON(200, data)
	})

	r.GET("/m/:id", func(c *gin.Context) {
		id := c.Param("id")
		rows, err := db.Query("select date, tx, rx from month where interface == ? order by id desc limit 12;", id)
		if err != nil {
			c.String(500, "Error querying database: %v", err)
			return
		}
		defer rows.Close()
		var data []dataCost
		for rows.Next() {
			var d dataCost
			err := rows.Scan(&d.Date, &d.Tx, &d.Rx)
			if err != nil {
				c.String(500, "Error scanning row: %v", err)
				return
			}
			data = append(data, d)
		}
		c.JSON(200, data)
	})

	r.GET("/y/:id", func(c *gin.Context) {
		id := c.Param("id")
		rows, err := db.Query("select date, tx, rx from year where interface == ? order by id desc limit 5;", id)
		if err != nil {
			c.String(500, "Error querying database: %v", err)
			return
		}
		defer rows.Close()
		var data []dataCost
		for rows.Next() {
			var d dataCost
			err := rows.Scan(&d.Date, &d.Tx, &d.Rx)
			if err != nil {
				c.String(500, "Error scanning row: %v", err)
				return
			}
			data = append(data, d)
		}
		c.JSON(200, data)
	})

	r.NoRoute(func(c *gin.Context) {
		c.String(200, "No Data")
	})

	err = r.Run(":" + *port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
