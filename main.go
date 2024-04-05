package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/yaml.v3"
)

var Mocks []Mock

var appService = fiber.New(fiber.Config{
	Immutable:             true,
	DisableStartupMessage: true,
})

var appMain = fiber.New(fiber.Config{
	Immutable:             true,
	DisableStartupMessage: true,
})

var mainConfig mainConf
var serviceConfig serviceConf

func main() {
	START_TIME := time.Now()
	fmt.Println("Initializing BeeLoadMock...")
	ENTITY_IDENTIFICATOR := USRmakeUUID()
	mainConfig = readMainConf()
	serviceConfig = readServiceConf()

	appService.Get("/stop", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		appMain.ShutdownWithTimeout(1000 * time.Millisecond)
		return c.SendString("Main mock handler is stopped")
	})

	appService.Get("/check", func(c *fiber.Ctx) error {
		return c.SendString("Service OK")
	})

	appService.Get("/shutdown", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		appMain.ShutdownWithTimeout(1000 * time.Millisecond)
		appService.ShutdownWithTimeout(1000 * time.Millisecond)
		os.Exit(0)
		return c.SendStatus(428)
	})
	Mocks = []Mock{}
	appMain = prepairMainHandler(mainConfig, ENTITY_IDENTIFICATOR, START_TIME)

	appService.Get("/start", func(c *fiber.Ctx) error {
		mainConfig = readMainConf()
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		Mocks = []Mock{}
		appMain = prepairMainHandler(mainConfig, ENTITY_IDENTIFICATOR, START_TIME)
		if mainConfig.EnableMockHTTPS {
			cer, err := tls.LoadX509KeyPair(mainConfig.MainPrivKey, mainConfig.MainPrivKey)
			if err != nil {
				panic(err)
			}
			serverTLSConf := &tls.Config{Certificates: []tls.Certificate{cer}}
			ln, err := tls.Listen("tcp", ":"+strconv.Itoa(serviceConfig.MainPort), serverTLSConf)
			if err != nil {
				panic(err)
			}
			go log.Fatal(appMain.Listener(ln))
		} else {
			go appMain.Listen(":" + strconv.Itoa(serviceConfig.MainPort))
		}
		return c.SendString("Main mock handler is started")
	})

	appService.Get("/reboot", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		appMain.ShutdownWithTimeout(1000 * time.Millisecond)
		Mocks = []Mock{}
		mainConfig = readMainConf()
		appMain = prepairMainHandler(mainConfig, ENTITY_IDENTIFICATOR, START_TIME)
		if mainConfig.EnableMockHTTPS {
			cer, err := tls.LoadX509KeyPair(mainConfig.MainPubKey, mainConfig.MainPrivKey)
			if err != nil {
				panic(err)
			}
			serverTLSConf := &tls.Config{Certificates: []tls.Certificate{cer}}
			ln, err := tls.Listen("tcp", ":"+strconv.Itoa(serviceConfig.MainPort), serverTLSConf)
			if err != nil {
				panic(err)
			}
			go log.Fatal(appMain.Listener(ln))
		} else {
			go appMain.Listen(":" + strconv.Itoa(serviceConfig.MainPort))
		}
		return c.SendString("Main mock handler rebooted")
	})

	appService.Post("/add-mock-file", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		filename := c.Query("name")
		fileroot := mainConfig.MocksRootFolder
		filepath := c.Query("path")
		if len(filename) < 1 {
			c.Status(400)
			return c.SendString("New file name cannot be empty")
		}
		if len(c.Query("name")) > 0 {
			payload := c.BodyRaw()
			full := ""
			if len(filepath) > 0 {
				full = fileroot + "/" + filepath + "/" + filename
			}
			if len(c.Query("path")) < 1 {
				full = fileroot + "/" + filename
			}
			fmt.Println(full)
			status, message := createFileAndWriteData(full, payload)
			c.Status(status)
			return c.SendString(message)
		}
		c.Status(500)
		return c.SendString("Unpredicted error while POST /add-mock-file")
	})

	appService.Post("/add-cert-file", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		filename := c.Query("name")
		fileroot := ""
		filepath := c.Query("path")
		if len(filename) < 1 {
			c.Status(400)
			return c.SendString("New file name cannot be empty")
		}
		if len(c.Query("name")) > 0 {
			payload := c.BodyRaw()
			full := ""
			if len(filepath) > 0 {
				full = fileroot + "/" + filepath + "/" + filename
			}
			if len(c.Query("path")) < 1 {
				full = fileroot + "/" + filename
			}
			fmt.Println(full)
			status, message := createFileAndWriteData(full, payload)
			c.Status(status)
			return c.SendString(message)
		}
		c.Status(500)
		return c.SendString("Unpredicted error while POST /add-cert-file")
	})

	appService.Delete("/remove-mock-file", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		deluuid := c.Query("mockuuid")
		fmt.Println("mockuuid to del = " + deluuid)
		MocksPathsArr, err := walkMatch(mainConfig.MocksRootFolder, "*.yaml")
		if err != nil {
			c.Status(500)
			return c.SendString("Error while DELETE /remove-mock-file")
		}
		for _, file := range MocksPathsArr {
			b, err := os.ReadFile(file)
			if err != nil {
				c.Status(500)
				return c.SendString("Error when DELETE /remove-mock-file")
			}
			if strings.Contains(string(b), deluuid) {
				fmt.Println(string(b))
				err = os.Remove(file)
				if err != nil {
					c.Status(500)
					fmt.Println(err)
					return c.SendString("Error deleting file when DELETE /remove-mock-file")
				} else {
					c.Status(200)
					return c.SendString("Mock file deleted. You need to restart main handler for the changes to take effect")
				}
			}
		}
		c.Status(500)
		return c.SendString("Unpredicted error while DELETE /remove-mock-file. May be there is no such file?")
	})

	appService.Get("/update-mainconfig-value", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		param := c.Query("parameter")
		newval := c.Query("newvalue")
		switch param {
		case "mocksRootFolder":
			mainConfig.MocksRootFolder = newval
		case "enableMockHttps":
			value, err := strconv.ParseBool(newval)
			if err != nil {
				c.Status(400)
				return c.SendString("Error converting new value to boolean")
			}
			mainConfig.EnableMockHTTPS = value
		case "mainPubKey":
			mainConfig.MainPubKey = newval
		case "mainPrivKey":
			mainConfig.MainPrivKey = newval
		case "dedicatedURLmockStatus":
			mainConfig.DedicatedURLmockStatus = newval
		case "dedicatedURLhealthcheck":
			mainConfig.DedicatedURLhealthcheck = newval
		case "enableFiberPrometheus":
			value, err := strconv.ParseBool(newval)
			if err != nil {
				c.Status(400)
				return c.SendString("Error converting new value to boolean")
			}
			mainConfig.EnableFiberPrometheus = value
		case "dedicatedURLfiberPrometheus":
			mainConfig.DedicatedURLfiberPrometheus = newval
		}
		bodybyte, err := yaml.Marshal(&mainConfig)
		if err != nil {
			c.Status(400)
			return c.SendString("Error marshalling yaml")
		}
		err = os.WriteFile("main-config.yaml", bodybyte, 0644)
		if err != nil {
			c.Status(400)
			fmt.Println(err)
			return c.SendString("Error writing yaml file")
		}
		c.Status(200)
		return c.SendString("Done. Reboot main handler.")
	})

	appService.Get("/get-mainmock-config", func(c *fiber.Ctx) error {
		if !isTokenValid(serviceConfig, c.Get("Blm-Agent-Token")) {
			c.Status(403)
			return c.SendString("Missing valid access token header")
		}
		byte, err := os.ReadFile("main-config.yaml")
		if err != nil {
			c.Status(400)
			return c.SendString("Error reading main-config.yaml")
		}
		c.Status(200)
		return c.SendString(string(byte))
	})

	fmt.Println(LOGO)

	if !mainConfig.EnableMockHTTPS {
		fmt.Println("Starting HTTP server... ")
		go appMain.Listen(":" + strconv.Itoa(serviceConfig.MainPort))
	} else {
		cer, err := tls.LoadX509KeyPair(mainConfig.MainPubKey, mainConfig.MainPrivKey)
		if err != nil {
			panic(err)
		}
		serverTLSConf := &tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err := tls.Listen("tcp", ":"+strconv.Itoa(serviceConfig.MainPort), serverTLSConf)
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting HTTPS server... ")
		go log.Fatal(appMain.Listener(ln))
	}

	if !serviceConfig.EnableServiceHTTPS {
		fmt.Println("\nVersion", VERSION)
		fmt.Println("Started time:", time.Since(START_TIME))

		appService.Listen(":" + strconv.Itoa(serviceConfig.ServicePort))
	} else {
		cer, err := tls.LoadX509KeyPair(serviceConfig.ServicePubKey, serviceConfig.ServicePrivKey)
		if err != nil {
			panic(err)
		}
		serverTLSConf := &tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err := tls.Listen("tcp", ":"+strconv.Itoa(serviceConfig.ServicePort), serverTLSConf)
		if err != nil {
			panic(err)
		}
		fmt.Println("\nVersion", VERSION)
		fmt.Println("Started time:", time.Since(START_TIME))

		log.Fatal(appService.Listener(ln))
	}
}
