package cloudroot

type DatabaseConfig struct {
	DatabaseName string
	Driver       string
	Host         string
	Port         string
	User         string
	Password     string
	Verify       bool
}

type OutgoingService struct {
	Name                  string `json:"Name"`
	Host                  string `json:"Host"`
	Timeout               int64  `json:"Timeout"`
	DefaultRequestHeaders []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"DefaultRequestHeaders"`
}
