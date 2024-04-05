package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

func prepairMainHandler(CONFIG mainConf, e string, t time.Time) *fiber.App {
	Mocks := []Mock{}
	newApp := fiber.New(fiber.Config{
		Immutable:             true,
		DisableStartupMessage: true,
	})

	fmt.Print("Searching files... ")
	MocksPathsArr, err := walkMatch(CONFIG.MocksRootFolder, "*.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %s file(s):\n", strconv.Itoa(len(MocksPathsArr)))

	for ii := range MocksPathsArr {
		fmt.Println(MocksPathsArr[ii])
	}

	for _, any := range MocksPathsArr {
		fmt.Print("Parsing " + any + "... ")
		new, status := ParseYamlToMock(any)
		if status < 0 {
			fmt.Print("Bad yaml file! Deleting...")
			err = os.Remove(any)
			if err != nil {
				fmt.Println("Error when deleting! Skiped.")
			} else {
				fmt.Println("OK")
			}
		} else {
			fmt.Println("Parsed OK")
			Mocks = append(Mocks, new)
		}
	}

	for i := range Mocks {
		for ii := range Mocks[i].Routes {
			if strings.HasPrefix(Mocks[i].Routes[ii].Body, "$[file(") && strings.HasSuffix(Mocks[i].Routes[ii].Body, ")]") {
				path := strings.TrimPrefix(Mocks[i].Routes[ii].Body, "$[file(")
				path = strings.TrimSuffix(path, ")]")
				respBody := tryReadBodyFile(CONFIG.MocksRootFolder + "/" + path)
				Mocks[i].Routes[ii].Body = respBody
			}
		}
	}

	for index := range Mocks {
		currMock := Mocks[index]
		newApp.Add(currMock.Method, currMock.Path, func(c *fiber.Ctx) error {
			start := time.Now()
			allVariables := extractAllVariables(currMock, *c)
			responseRoute := chooseResponseRoute(allVariables, currMock.Routes)
			convertedHeaders := ConvertHeaders(responseRoute.Headers)
			for _, convertedHeader := range convertedHeaders {
				c.Set(convertedHeader.Name, convertedHeader.Value)
			}
			if currMock.IncludeMockInfo {
				c.Set("BeeLoadMock-identificator", currMock.Identificator)
				c.Set("BeeLoadMock-authored", currMock.Authored)
				c.Set("BeeLoadMock-about", currMock.About)
				c.Set("BeeLoadMock-route", strconv.Itoa(responseRoute.Priority))
			}
			completedBody := CompleteBody(allVariables, responseRoute.Body)
			completedBody = USRreleaseFunctions(completedBody)
			c.Context().SetStatusCode(responseRoute.Code)
			if currMock.IncludeMockInfo {
				c.Set("BeeLoadMock-worktime", time.Since(start).String())
			}
			//elapsed := time.Now().Sub(start).Milliseconds()
			MakeSleep(RandIntInRange(responseRoute.SleepMin, responseRoute.SleepMax))
			return c.SendString(completedBody)
		})
	}

	newApp.Get(CONFIG.DedicatedURLmockStatus, func(c *fiber.Ctx) error {
		response := "BeeLoadMock, version " + VERSION + "\n"
		response += "Current uptime " + time.Since(t).String() + "\n"
		response += "Entity " + e + "\n"
		response += "\nMOCKS:\n"
		response += "identificator\t\t\t\tmethod\tpath" + "\n"
		for _, mock := range Mocks {
			response += mock.Identificator + "\t" + mock.Method + "\t" + mock.Path + "\n"
		}
		return c.SendString(response)
	})

	newApp.Get(CONFIG.DedicatedURLhealthcheck, func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	if CONFIG.EnableFiberPrometheus {
		prometheus := fiberprometheus.New("BeeLoadMock")
		prometheus.RegisterAt(appMain, CONFIG.DedicatedURLfiberPrometheus)
		appMain.Use(prometheus.Middleware)
	}

	return newApp
}
