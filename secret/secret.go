package secret

type Secret struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Id    string      `json:"id"`
}

type Store interface {
	Exists(name string) bool
	GetLatestByName(name string) (Secret, error)
	GetByName(name string) ([]Secret, error)
	GetById(id string) (Secret, error)
	Set(name string, value interface{}) (string, error)
	DeleteByName(name string) error
	Healthy() bool
}
