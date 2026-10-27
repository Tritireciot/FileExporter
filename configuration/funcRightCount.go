package config

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// Описание суммирования одикичного права
func SumRight(v1 RightValType, v2 RightValType) RightValType {
	if v1 == RightValTypeRestricted || v2 == RightValTypeRestricted { // if запрещено - запрещено в любом случае
		return RightValTypeRestricted
	} else { // Тут запрещенно не будет
		if v1 == RightValTypeAllow || v2 == RightValTypeAllow {
			return RightValTypeAllow
		}
	}
	return RightValTypeNo
}

// ---------------------------------------------------
// Суммирование значений прав
func SumRights(v1 RightValues, v2 RightValues) RightValues {
	for id, v := range v2 {
		v1Val, foundInV1 := v1[id]
		// Если в первом массиве есть уже это значение
		if foundInV1 {
			v1[id] = SumRight(v, v1Val)
		} else { // Если в первом массиве нет значений - присваиваем в любом случае
			v1[id] = v
		}

	}
	return v1
}

// ---------------------------------------------------
func SumRightGroup(rvals RightValues, AllowedRights []int, RestrictedRights []int) RightValues {
	for _, r := range AllowedRights {
		if v, found := rvals[r]; found {
			rvals[r] = SumRight(v, RightValTypeAllow)
		} else {
			rvals[r] = RightValTypeAllow
		}
	}
	for _, r := range RestrictedRights {
		if v, found := rvals[r]; found {
			rvals[r] = SumRight(v, RightValTypeRestricted)
		} else {
			rvals[r] = RightValTypeRestricted
		}
	}
	return rvals
}

// ---------------------------------------------------
func countRightsForUserGroup(ugId int, rgmap map[int]RightGroup) (rvals RightValues) {
	rvals = make(RightValues)
	for _, rg := range rgmap {
		for _, u := range rg.UserGroups {
			if u == ugId {
				rvals = SumRightGroup(rvals, rg.AllowedRights, rg.RestrictedRights)

				break // группа может быть добвлена только раз
			}
		}
	}
	return
}

// ----------------------------------------------------------
// Непосредственный расчет функциональных прав пользователя исходя из пользовательских и правовых групп
func CountUserAccess(userId int, rgMap map[int]RightGroup, ugMap map[int]UserGroup) RightValues {
	const maxDeepIn = 16
	var recUGFunc func(*RightValues, int /*в какой поль.гр.находимся*/, map[int]UserGroup, map[int]RightGroup, *int)
	recUGFunc = func(v *RightValues, whereUgId int, ugmap map[int]UserGroup, rgmap map[int]RightGroup, deepIn *int) {
		if *deepIn++; *deepIn > maxDeepIn {
			return
		}
		if ug, found := ugmap[whereUgId]; found {
			// Если у текущая группа не включает другие - крайняя. Применяем права
			if len(ug.GroupsIds) == 0 {
				SumRights(*v, countRightsForUserGroup(ug.Id, rgmap))
			} else {
				// Проходим по группам текущей группы заходя в каждую
				for i := range ug.GroupsIds {
					recUGFunc(v, ug.GroupsIds[i], ugmap, rgmap, deepIn)
				}
			}
			// При окончании обработки прав группы - суммируем с текущими правыми группы
			SumRights(*v, countRightsForUserGroup(ug.Id, rgmap))
		}
		*deepIn--
	}

	// В каких пользовательских группах находится пользователь
	rv := RightValues{}
	// Расчет исходя из этих данных
	for ugId, ug := range ugMap {
		for _, uId := range ug.UsersIds {
			if userId == uId {
				deepIn := 0
				recUGFunc(&rv, ugId, ugMap, rgMap, &deepIn)
				break
			}
		}
	}
	return rv
}

// ---------------------------------------------------
// Проверка права для определенного пользователя локально(если есть локальный доступ к БД)
func CheckUserAccessLocal(pool *pgxpool.Pool, rightId int, userId int, ugMap map[int]UserGroup) (bool, error) {
	if userId == superUserId {
		return true, nil
	}
	// ----------------------------------------------------------
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
	//----------------------------------------------------------
	// Рассчитаем права пользователя исходя из групп
	rs := CountUserAccess(userId, rgMap, ugMap)

	// Применим значения прав по-умолчанию (временно отключаем)
	/*var err error
	if rs, err = applySchemaDefaults(pool, rs); err != nil {
		return false, err
	}*/
	//----------------------------------------------------------
	if rval, found := rs[rightId]; found {
		return rval == RightValTypeAllow, nil
	} else { // Право не нашли, возможно оно устрело либо не указано
		return false, nil
	}
}

const superUserId = 1

// Проверка права для определенного пользователя
func CheckUserAccess(pool *pgxpool.Pool, rightId int, userId int, subAccessPath string, subAccessToken string) (bool, error) {
	// Проверка на суперюзера
	if userId == superUserId {
		return true, nil
	}
	// Получим группы пользователей
	if ugMap, err := GetUserGroupsByRequest(subAccessPath, subAccessToken); err != nil {
		return false, err
	} else {
		return CheckUserAccessLocal(pool, rightId, userId, ugMap)
	}
}
