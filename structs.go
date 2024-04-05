package main

// структура конфига экзекьютора

type mainConf struct {
	MocksRootFolder             string `yaml:"mocksRootFolder"`
	EnableMockHTTPS             bool   `yaml:"enableMockHttps"`
	MainPubKey                  string `yaml:"mainPubKey"`
	MainPrivKey                 string `yaml:"mainPrivKey"`
	DedicatedURLmockStatus      string `yaml:"dedicatedURLmockStatus"`
	DedicatedURLhealthcheck     string `yaml:"dedicatedURLhealthcheck"`
	EnableFiberPrometheus       bool   `yaml:"enableFiberPrometheus"`
	DedicatedURLfiberPrometheus string `yaml:"dedicatedURLfiberPrometheus"`
}

type serviceConf struct {
	MainPort               int      `yaml:"mainPort"`
	ServicePort            int      `yaml:"servicePort"`
	EnableServiceHTTPS     bool     `yaml:"enableServiceHttps"`
	ServicePubKey          string   `yaml:"servicePubKey"`
	ServicePrivKey         string   `yaml:"servicePrivKey"`
	AccessTokenRequirement bool     `yaml:"accessTokenRequirement"`
	AccessTokenList        []string `yaml:"accessTokenList"`
}

// b a s e d
// структура всего файла с моком
type Mock struct {
	Identificator   string         `yaml:"identificator"`
	Authored        string         `yaml:"authored"`
	About           string         `yaml:"about"`
	IncludeMockInfo bool           `yaml:"includeMockInfo"`
	Method          string         `yaml:"method"`
	Path            string         `yaml:"path"`
	PathVariables   []string       `yaml:"pathVariables"`
	QueryVariables  []string       `yaml:"queryVariables"`
	HeaderVariables []string       `yaml:"headerVariables"`
	BodyVariables   []BodyVariable `yaml:"bodyVariables"`
	Routes          []Route        `yaml:"routes"`
}

// структура, которая овтечает за извлекаемые из тела переменные, является вложением в основное
type BodyVariable struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Rule string `yaml:"rule"`
}

// структура, которая отвечает за конкретный ответ из списка - код, задержка, хэдеры, тело, подчасть основы
type Route struct {
	Priority            int                  `yaml:"priority"`
	FulfilledConditions []FulfilledCondition `yaml:"fulfilledConditions"`
	Code                int                  `yaml:"code"`
	SleepMin            int                  `yaml:"sleepMin"`
	SleepMax            int                  `yaml:"sleepMax"`
	Headers             []string             `yaml:"headers"`
	Body                string               `yaml:"body"`
}

// структура, которая отвечает за проверку условий переменных для выбора нужного варианта ответа из списка, подчасть Route
type FulfilledCondition struct {
	Variable string `yaml:"variable"`
	Value    string `yaml:"value"`
}

// структура с хэдерами, нужна для более удобного добавления в овтет
type Header struct {
	Name  string
	Value string
}
