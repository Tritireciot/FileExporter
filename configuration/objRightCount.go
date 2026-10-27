package config

import (
	"log"
	"reflect"

	"github.com/jackc/pgx/v4/pgxpool"
)

// ---------------------------------------------------
func countObjectRightsForUserGroup(ugId int, objId int, objTypeId int, rgmap map[int]RightGroup) (rvals RightValues) {
	rvals = make(RightValues)
	for _, rg := range rgmap {
		for _, u := range rg.UserGroups {
			if u == ugId {
				for _, o := range rg.ObjectsRights {
					if o.ObjectId == objId && o.ObjectTypeId == objTypeId {
						rvals = SumRightGroup(rvals, o.Allowed, o.Restricted)
					}
				}
				break // группа может быть добвлена только раз
			}
		}
	}
	return
}

// ---------------------------------------------------
// Рекурсивный проход по группам для рассчета результата прав по указанному объекту
func CountUserObjectAccess(userId int, objectId int, objectTypeId int, rgMap map[int]RightGroup, ugMap map[int]UserGroup) RightValues {
	const maxDeepIn = 16
	var recUGFunc func(*RightValues, int /*в какой поль.гр.находимся*/, int /**/, int, map[int]UserGroup, map[int]RightGroup, *int)
	recUGFunc = func(v *RightValues, whereUgId int, objId int, objTypeId int, ugmap map[int]UserGroup, rgmap map[int]RightGroup, deepIn *int) {
		if *deepIn++; *deepIn > maxDeepIn {
			return
		}
		if ug, found := ugmap[whereUgId]; found {
			// Если у текущая группа не включает другие - крайняя. Применяем права
			if len(ug.GroupsIds) == 0 {
				(*v) = SumRights(*v, countObjectRightsForUserGroup(ug.Id, objId, objTypeId, rgmap))
			} else {
				// Проходим по группам текущей группы заходя в каждую
				for i := range ug.GroupsIds {
					recUGFunc(v, ug.GroupsIds[i], objId, objTypeId, ugmap, rgmap, deepIn)
				}
			}
			// При окончании обработки прав группы - суммируем с текущими правыми группы
			(*v) = SumRights(*v, countObjectRightsForUserGroup(ug.Id, objId, objTypeId, rgmap))
		}
		*deepIn--
	}

	// В каких пользовательских группах находится пользователь
	rv := RightValues{}
	// Расчет исходя из этих данных
	for ugId, ug := range ugMap {
		for _, uId := range ug.UsersIds {
			if userId == uId { // Мы нашли сою группу
				deepIn := 0
				recUGFunc(&rv, ugId, objectId, objectTypeId, ugMap, rgMap, &deepIn)
				break
			}
		}
	}
	return rv
}

// Проверка прав на объект (c локально переданными ugMap)
func CheckObjectAccessLocal(pool *pgxpool.Pool, objectId int, objectTypeId int, userId int, rightId int, ugMap map[int]UserGroup) (bool, error) {
	rgMap := map[int]RightGroup{}
	{ // Получим группы прав
		rgs, _, err := GetRightGroups(pool)
		if err != nil {
			return false, err
		}
		for _, rg := range rgs {
			rgMap[rg.Id] = rg
		}
	}
	rs := CountUserObjectAccess(userId, objectId, objectTypeId, rgMap, ugMap)
	//----------------------------------------------------------
	if rval, found := rs[rightId]; found {
		return rval == RightValTypeAllow, nil
	} else { // Право не нашли, возможно оно устрело либо не указано
		return false, nil
	}
}

// Проверка права на объект
func CheckObjectAccess(pool *pgxpool.Pool, objectId int, objectTypeId int, userId int, rightId int, subAccessPath string, subAccessToken string) (bool, error) {
	// Проверка на суперюзера
	if userId == superUserId {
		return true, nil
	}

	// Получим группы пользователей
	if ugMap, err := GetUserGroupsByRequest(subAccessPath, subAccessToken); err != nil {
		return false, err
	} else {
		return CheckObjectAccessLocal(pool, objectId, objectTypeId, userId, rightId, ugMap)
	}
}

// ----------------------------------------------------------------
func CheckObjectAvailabilityLocal(pool *pgxpool.Pool, objectId int, objectTypeId int, userId int, ugMap map[int]UserGroup) (bool, error) {
	rgMap := map[int]RightGroup{}
	{ // Получим группы прав
		rgs, _, err := GetRightGroups(pool)
		if err != nil {
			return false, err
		}
		for _, rg := range rgs {
			rgMap[rg.Id] = rg
		}
	}
	rs := CountUserObjectAccess(userId, objectId, objectTypeId, rgMap, ugMap)
	//----------------------------------------------------------
	for _, v := range rs {
		if v == RightValTypeAllow {
			return true, nil
		}
	}
	return false, nil
}

// Проверка права на объект
func CheckObjectAvailability(pool *pgxpool.Pool, objectId int, objectTypeId int, userId int, subAccessPath string, subAccessToken string) (bool, error) {
	// Проверка на суперюзера
	if userId == superUserId {
		return true, nil
	}
	// Получим группы пользователей
	if ugMap, err := GetUserGroupsByRequest(subAccessPath, subAccessToken); err != nil {
		return false, err
	} else {
		return CheckObjectAvailabilityLocal(pool, objectId, objectTypeId, userId, ugMap)
	}
}

// ----------------------------------------------------------------
// Интерфейс который определяем возможность объектов выдавать ИД.
// Этот ИД объектов используется при проверке прав на объекты
type Indentifiable interface {
	GetRightObjectId() int
}

func FilterObjectsByAvailabilityLocal(objs interface{} /*Indentifiable*/, objectTypeId int, userId int, rgMap map[int]RightGroup, ugMap map[int]UserGroup) interface{} {
	v := reflect.ValueOf(objs)
	if v.Kind() != reflect.Slice {
		log.Printf("incorrent obj type in FilterObjectsByAvailabilityLocal! It is not a slice! %v", objs)
		return reflect.MakeSlice(reflect.TypeOf(objs), 0, 0).Interface()
	} else { // Проверим на корректность типа
		if v.Len() > 0 {
			if _, ok := v.Index(0).Interface().(Indentifiable); !ok {
				log.Printf("incorrent obj type in FilterObjectsByAvailabilityLocal! %v", objs)
				return reflect.MakeSlice(reflect.TypeOf(objs), 0, 0).Interface()
			}
		}
	}

	availableObjects := reflect.MakeSlice(reflect.TypeOf(objs), 0, 0)
	for i := 0; i < v.Len(); i++ {
		if o, ok := v.Index(i).Interface().(Indentifiable); ok {
			// Получим набор прав по объекту
			rs := CountUserObjectAccess(userId, o.GetRightObjectId(), objectTypeId, rgMap, ugMap)
			for _, rv := range rs {
				if rv == RightValTypeAllow { // Если хотябы одно право по объекту разрешено - он видим
					availableObjects = reflect.Append(availableObjects, v.Index(i))
					break
				}
			}
		}
	}
	return availableObjects.Interface()
}

// Фильтр типовых объектов по доступным правам пользователя
func FilterObjectsByAvailability(pool *pgxpool.Pool, objs interface{} /*Indentifiable*/, objectTypeId int, userId int, subAccessPath string, subAccessToken string) (interface{}, error) {
	// Получим группы пользователей
	if ugMap, err := GetUserGroupsByRequest(subAccessPath, subAccessToken); err != nil {
		return nil, err
	} else {
		rgMap := map[int]RightGroup{}
		{ // Получим группы прав
			rgs, _, err := GetRightGroups(pool)
			if err != nil {
				return []Indentifiable{}, err
			}
			for _, rg := range rgs {
				rgMap[rg.Id] = rg
			}
		}

		return FilterObjectsByAvailabilityLocal(objs, objectTypeId, userId, rgMap, ugMap), nil
	}
}
