package secret

type Secret struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Id    string      `json:"id"`
}

type Store interface {
	GetLatestByName(name string) (Secret, error)
	GetAllByName(name string) ([]Secret, error)
	GetById(id string) (Secret, error)
	Set(name string, value interface{}) (string, error)
	DeleteByName(name string) error
	Healthy() bool
}