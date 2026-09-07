package pgutils

// Версия БД
type DbVersion struct {
	Tags []string // Тэги
	Num  int      // Номер версии
}

func (dbv *DbVersion) ContainTag(tag string) bool {
	for _, v := range dbv.Tags {
		if v == tag {
			return true
		}
	}
	return false
}

const LATEST = "latest"
