package main

import (
	"log"
	"os/exec"
	"sync"

	"github.com/gofiber/fiber/v2"
	// "io"
	"os"
	"time"
)

var LOG_FILE = "log.txt"
var LOG_DIR = "tmp"

var ECHO_FILEPATH string = LOG_DIR + "/echo_" + LOG_FILE

var logChan chan string
var (
	aoFile *os.File
	aoMux  sync.Mutex
)

type PingRequest struct {
	Message string `json:"message"`
}

func pingPongStr(c *fiber.Ctx) error {
	return c.SendString("pong")
}

func pingPongJson(c *fiber.Ctx) error {
	req := new(PingRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	if req.Message == "ping" {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	}
	return c.JSON(fiber.Map{
		"message": req.Message,
	})
}

func logStr(c *fiber.Ctx) error {
	c.Accepts("text/plain")
	body := c.Body()

	file, err := os.OpenFile(LOG_DIR+"/"+LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(
		time.Now().String() +
			" | " + c.IP() +
			" | " + c.Method() +
			" | " + string(body) +
			"\n"); err != nil {
		return err
	}

	// Flush any buffered data to disk
	// if err := file.Sync(); err != nil {
	// 	log.Println(err)
	// }

	return c.SendString("ACK Was sent: " + string(body))
}

func logEchoStr(c *fiber.Ctx) error {
	c.Accepts("text/plain")
	body := c.Body()

	cmd := exec.Command("sh", "-c", "echo "+string(body)+" > "+ECHO_FILEPATH)
	if err := cmd.Run(); err != nil {
		return err
	}

	return c.SendString("ACK Was sent: " + string(body))
}

func logAoStr(c *fiber.Ctx) error {
	c.Accepts("text/plain")
	body := c.Body()

	aoMux.Lock()
	defer aoMux.Unlock()
	if aoFile == nil {
		var err error
		aoFile, err = os.OpenFile(LOG_DIR+"/"+LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	if _, err := aoFile.WriteString(
		time.Now().String() +
			" | " + c.IP() +
			" | " + c.Method() +
			" | " + string(body) +
			"\n"); err != nil {
		return err
	}

	return c.SendString("ACK")
}

func logAoChanStr(c *fiber.Ctx) error {
	c.Accepts("text/plain")
	body := c.Body()

	// Send the request to the main thread via a channel
	logChan <- time.Now().String() +
		" | " + c.IP() +
		" | " + c.Method() +
		" | " + string(body)
	return c.SendString("ACK")
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "fiberpt",
		AppName:       "Fiber Perf Test v1.0.0",
	})

	// Generic noop endpoint

	app.Get("/", pingPongStr)

	// Ping-pong endpoint
	app.Get("/ping", pingPongStr)

	// Ping-pong endpoint with JSON input
	app.Post("/ping", pingPongJson)

	app.Post("/log", logStr)
	app.Post("/logecho", logEchoStr)
	app.Post("/logao", logAoStr)
	app.Post("/logaochan", logAoChanStr)

	// Open file to append
	aoChanFile, err := os.OpenFile(LOG_DIR+"/ao_chan_"+LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer aoChanFile.Close()

	// Create a channel for logging requests
	logChan = make(chan string)

	// Start a goroutine to handle logging requests
	go func() {
		for {
			// Read requests from the channel
			msg := <-logChan

			// Write the request to the file
			if _, err := aoChanFile.WriteString(msg + "\n"); err != nil {
				log.Println(err)
			}

			// Flush any buffered data to disk
			// if err := aoChanFile.Sync(); err != nil {
			// 	log.Println(err)
			// }
		}
	}()

	// Start server
	err = app.Listen("localhost:3000")
	if err != nil {
		panic(err)
	}
}
