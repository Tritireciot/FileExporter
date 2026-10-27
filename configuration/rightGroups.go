package config

import (
	"AutoplayX/pgutils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Права на объект
type ObjectRights struct {
	ObjectId     int   `json:"objectId"`
	ObjectTypeId int   `json:"objectTypeId"`
	Allowed      []int `json:"allowedRights"`
	Restricted   []int `json:"restrictedRights"`
}

// --------------------------------------------------------------------------------------
// Группа прав
type RightGroup struct {
	Id               int            `json:"id,omitempty"`
	Name             string         `json:"name"`
	AllowedRights    []int          `json:"allowedRights"`
	RestrictedRights []int          `json:"restrictedRights"`
	UserGroups       []int          `json:"userGroups"`
	ObjectsRights    []ObjectRights `json:"objectsRights"`
}

// ---------------------------------------------------
func GetRightGroups(pool *pgxpool.Pool) ([]RightGroup, int /*ver*/, error) {
	ans := struct {
		Ver         int          `json:"version"`
		RightGroups []RightGroup `json:"rightGroups"`
	}{}
	err := pool.QueryRow(context.Background(), `select config."GetRightGroups"()`).Scan(&ans)
	return ans.RightGroups, ans.Ver, err
}

// ---------------------------------------------------
func GetRightGroupsMap(pool *pgxpool.Pool) (map[int]RightGroup, int /*ver*/, error) {
	ans := struct {
		Ver         int          `json:"version"`
		RightGroups []RightGroup `json:"rightGroups"`
	}{}
	err := pool.QueryRow(context.Background(), `select config."GetRightGroups"()`).Scan(&ans)

	rgMap := map[int]RightGroup{}
	for _, rg := range ans.RightGroups {
		rgMap[rg.Id] = rg
	}

	return rgMap, ans.Ver, err
}

// ---------------------------------------------------
type UserGroup struct {
	Id        int    `json:"id,omitempty"`
	Name      string `json:"name"`
	UsersIds  []int  `json:"users"`
	GroupsIds []int  `json:"groups"`
}

type RightValType int

const (
	RightValTypeNo         RightValType = 0
	RightValTypeAllow      RightValType = 1
	RightValTypeRestricted RightValType = 2
)

// Установка для права
type RightValue struct {
	Id  int          `json:"id"` // Идентификатор права
	Val RightValType `json:"type"`
}

type RightValues map[int]RightValType

// -------------------------------------------------------------------------------
// Отображение. Здесь ключ - ISO значение языка. значение - текст на данном языке
const (
	RusMapLangTag = "ru"
	EngMapLangTag = "en"
)

type LangMapString map[string]string

// Один из вариантов параметра схемы
type SchemaParam struct {
	Name LangMapString `json:"name"`
	Id   int           `json:"id"`

	Default bool `json:"default,omitempty"`

	Description LangMapString `json:"description,omitempty"`

	Childs []SchemaParam `json:"childs,omitempty"`
}

type ObjectSchemaParams struct {
	ObjectType int `json:"objectType"`

	Rights []SchemaParam `json:"rights"`
}

type RightSchema struct {
	Rights        []SchemaParam        `json:"rights"`
	ObjectsRights []ObjectSchemaParams `json:"objectsRights"`
}

const rightSchemaValue = "rights.schema"

// ----------------------------------------------------------
func GetRightSchema(pool *pgxpool.Pool) (RightSchema, int /*version*/, error) {
	answer := struct {
		SchemaStr string `json:"value"`
	}{}

	schema := RightSchema{}

	nstr := pgutils.NullString{}
	if err := pool.QueryRow(context.Background(), fmt.Sprintf(`select config."GetConfiguration"('%s')`, rightSchemaValue)).Scan(&nstr); err != nil {
		return schema, 0, err
	}

	err := json.Unmarshal([]byte(nstr.Str), &answer)
	err = json.Unmarshal([]byte(answer.SchemaStr), &schema)

	return schema, 0 /*not implemented yet*/, err
}

// ----------------------------------------------------------
// func applySchemaDefaults(pool *pgxpool.Pool, rightValues RightValues) (RightValues, error) {
// 	// Получим схему прав
// 	schema, _, err := GetRightSchema(pool)
// 	if err != nil {
// 		return rightValues, err
// 	}
// 	// Применим значения по-умолчанию схемы
// 	var applyParam func(SchemaParam, *RightValues)
// 	applyParam = func(sparam SchemaParam, rv *RightValues) {
// 		for _, chpar := range sparam.Childs {
// 			applyParam(chpar, rv)
// 		}
// 	}
// 	// Применим все параметры схемы для учета значений по умолчанию
// 	for i := range schema.Rights {
// 		applyParam(schema.Rights[i], &rightValues)
// 	}

// 	return rightValues, nil
// }
