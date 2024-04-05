package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

func USRmakeUUID() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "err creating user UUID"
	}
	return uuid.String()
}

// рандомная строка заданной длинны, состоящая из указанных в строке символов
func USRmakeRandString(length int, chars string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// время строкой в указанном формате
// https://yourbasic.org/golang/format-parse-string-time-date-example/
func USRmakeCurrentTimeFormatted(format string) string {
	return time.Now().Format(format)
}

// прогнать все функции в теле; $[uuid]
func USRreleaseFunctions(body string) string {
	// подставить все uuid
	for strings.Contains(body, "$[uuid]") {
		body = strings.Replace(body, "$[uuid]", USRmakeUUID(), 1)
	}
	// подставить все times; $[timeNowFormatted(2009-11-10 23:00:00)]
	for strings.Contains(body, "$[timeNowFormatted(") {
		formatter := BoundaryExtract(body, "$[timeNowFormatted(", ")]")
		body = strings.Replace(body, "$[timeNowFormatted("+formatter+")]", USRmakeCurrentTimeFormatted(formatter), 1)
	}
	// подставить все string; $[randomString(abcd1234;;10)]
	for strings.Contains(body, "$[randomString(") {
		stringRule := BoundaryExtract(body, "$[randomString(", ")]")
		ruleVars := stringRule[:strings.Index(stringRule, ";;")]
		ruleLengthSTR := stringRule[strings.Index(stringRule, ";;")+2:]
		ruleLength, err := strconv.Atoi(ruleLengthSTR)
		if err != nil {
			ruleLength = 0
		}
		body = strings.Replace(body, "$[randomString("+stringRule+")]", USRmakeRandString(ruleLength, ruleVars), 1)
	}
	return body
}

// проверяем каждую из переменных есть ли она в теле, да - меняем
func CompleteBody(variablesList map[string]string, givenBody string) string {
	for key, val := range variablesList {
		for strings.Contains(givenBody, "${"+key+"}") {
			givenBody = ReplaceWithTag(givenBody, "${"+key+"}", val)
		}
	}
	return givenBody
}

func readMainConf() mainConf {
	fmt.Print("Trying to read main-config.yaml...")
	data, err := os.ReadFile("main-config.yaml")
	if err != nil {
		panic(err)
	}
	mainConfig := mainConf{}
	err = yaml.Unmarshal(data, &mainConfig)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(" Probably good")
	}
	fmt.Println(mainConfig)
	return mainConfig
}

func readServiceConf() serviceConf {
	fmt.Print("Trying to read service-config.yaml...")
	data, err := os.ReadFile("service-config.yaml")
	if err != nil {
		panic(err)
	}
	serviceConfig := serviceConf{}
	err = yaml.Unmarshal(data, &serviceConfig)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(" Probably good")
	}
	fmt.Println(serviceConfig)
	return serviceConfig
}

// более удобное переваривание хэдеров
func ConvertHeaders(headers []string) []Header {
	var convertedHeaders = []Header{}
	for _, header := range headers {
		a, b := CutString(header, ":")
		convertedHeaders = append(convertedHeaders, Header{a, b})
	}
	return convertedHeaders
}

// выбор нужного роута для ответа, пробегается по приоритету с нуля, возвращается номер приоритета
func chooseResponseRoute(variablesList map[string]string, routesList []Route) Route {
	routesCount := len(routesList)
	if routesCount < 1 {
		return Route{-1, []FulfilledCondition{}, 404, 0, 0, []string{}, "ERR: BeeLoadMock has no routes avaliable for this URL"}
	}
	for i, currentRoute := range routesList {
		idx := slices.IndexFunc(routesList, func(c Route) bool { return c.Priority == i })
		successCounter := 0
		for _, condition := range routesList[idx].FulfilledConditions {
			if condition.Value == variablesList[condition.Variable] {
				successCounter++
			}
		}
		if successCounter == len(routesList[idx].FulfilledConditions) {
			return currentRoute
		}
	}
	return Route{-2, []FulfilledCondition{}, 404, 0, 0, []string{}, "ERR: BeeLoadMock didn't match any route for this URL"}
}

// функция для поиска файлов, возврщает массив строк с адресами файлов
func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, err
}

// получает путь до yaml-файла и возвращает объект типа Mock
func ParseYamlToMock(MockPath string) (Mock, int) {
	file, err := os.Open(MockPath)
	if err != nil {
		return Mock{}, -1
	}
	defer file.Close()
	mockByte, err := io.ReadAll(file)
	if err != nil {
		if err != nil {
			return Mock{}, -1
		}
	}
	var mock Mock
	err = yaml.Unmarshal(mockByte, &mock)
	if err != nil {
		return Mock{}, -1
	}
	return mock, 1
}

// вытаскиваем все переменные в единый а-ля словарь
func extractAllVariables(currentMock Mock, request fiber.Ctx) map[string]string {
	// база словаря
	dict := make(map[string]string)
	// прогоняем переменные из урла
	for _, curr := range currentMock.PathVariables {
		dict[curr] = request.Params(curr)
	}
	// прогоняем переменные из квери
	for _, curr := range currentMock.QueryVariables {
		dict[curr] = request.Query(curr)
	}
	// прогоняем переменные из хэдеров
	for _, curr := range currentMock.HeaderVariables {
		dict[curr] = string(request.Request().Header.Peek(curr))
	}
	// прогоянем переменные из бади
	for _, bodyVar := range currentMock.BodyVariables {
		body := string(request.Body())
		// если баундари извлечение
		if bodyVar.Type == "boundary-extract" {
			left, right := CutString(bodyVar.Rule, "<<l$r>>")
			dict[bodyVar.Name] = BoundaryExtract(body, left, right)
		}
		// если регуляркой
		if bodyVar.Type == "regexp-extract" {
			dict[bodyVar.Name] = RegExpExtract(body, bodyVar.Rule)
		}
		// если проверка true/false регуляркой
		if bodyVar.Type == "regexp-exist" {
			if RegExpExists(body, bodyVar.Rule) {
				dict[bodyVar.Name] = "true"
			} else {
				dict[bodyVar.Name] = "FALSE"
			}
		}
	}
	return dict
}

// вытаскивает подстроку по левой и правой границе
func BoundaryExtract(where string, begin string, end string) string {
	if !strings.Contains(where, begin) || !strings.Contains(where, end) {
		return `NOT_FOUND_BOUNDARY`
	}
	startIndex := strings.Index(where, begin)
	length := len(begin)
	where = where[startIndex+length:]
	endIndex := strings.Index(where, end)
	if endIndex < 0 {
		return `NOT_FOUND_BOUNDARY`
	}
	resp := where[:endIndex]
	return resp
}

// вытаскивает из строки первое вхождение регуляркой
func RegExpExtract(where string, with string) string {
	var re = regexp.MustCompile(with)
	matches := re.FindStringSubmatch(where)
	if len(matches) == 0 {
		return `NOT_FOUND_REGEXP`
	}
	return matches[1]
}

// проверяет выполнится ли регулярка, возвращает true/false
func RegExpExists(where string, with string) bool {
	var re = regexp.MustCompile(with)
	res := re.Match([]byte(where))
	if res {
		return true
	} else {
		return false
	}
}

// разрезает строку с помощью подстроки, возвращает левую и правую части
func CutString(target string, with string) (string, string) {
	a := target[:strings.Index(target, with)]
	b := target[strings.Index(target, with)+len(with):]
	return a, b
}

func ReplaceWithTag(where string, tag string, with string) string {
	startIndex := strings.Index(where, tag)
	endIndex := startIndex + len(tag)
	return where[:startIndex] + with + where[endIndex:]
}

// прождать целое число милисекунд
func MakeSleep(lasts int) {
	time.Sleep(time.Millisecond * time.Duration(lasts))
}

// выбрать случайное целое в диапазоне ОТ и ДО
func RandIntInRange(min int, max int) int {
	return rand.Intn(max-min+1) + min
}

func tryReadBodyFile(filePath string) string {
	byteFile, err := os.ReadFile(filePath)
	if err != nil {
		return "ERR: BeeLoadMock cannot find specific file to replace the body!"
	}
	return string(byteFile)
}

func createFileAndWriteData(name string, data []byte) (int, string) {
	os.MkdirAll(filepath.Dir(name), 0777)
	file, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return 400, "Error while creating file"
	}
	_, err = file.Write(data)
	if err != nil {
		return 400, "Error while writing data to file"
	}
	defer file.Close()
	return 201, "Created. You need to restart main handler for the changes to take effect"
}

func isTokenValid(config serviceConf, headerValue string) bool {
	if !config.AccessTokenRequirement {
		return true
	} else if slices.Contains(config.AccessTokenList, headerValue) {
		return true
	}
	return false
}
