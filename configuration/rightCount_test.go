package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Проверка функции суммирования прав
func TestSumRights(t *testing.T) {
	assert.Equal(t, RightValues{
		1: RightValTypeAllow,
		2: RightValTypeRestricted,
		3: RightValTypeAllow,
	}, SumRights(RightValues{
		1: RightValTypeAllow,
		2: RightValTypeAllow,
	}, RightValues{
		2: RightValTypeRestricted,
		3: RightValTypeAllow,
	}))
}

// Проверка расчета прав в случае зацикливания пользовательских групп. Учитываем до определенной глубины
func TestCountRightsInfinite(t *testing.T) {
	ugMap := map[int]UserGroup{
		1: {Id: 1, GroupsIds: []int{2, 3}, UsersIds: []int{1}},
		2: {Id: 2, GroupsIds: []int{4, 5}},
		3: {Id: 3},
		4: {Id: 4},
		5: {Id: 5, GroupsIds: []int{1}},
	}
	// 1->2->5->1
	rgMap := map[int]RightGroup{
		1: {
			UserGroups: []int{1},

			AllowedRights:    []int{},
			RestrictedRights: []int{6},
		},
		2: {
			UserGroups: []int{2},

			RestrictedRights: []int{5},
		},
		3: {
			UserGroups: []int{3},

			AllowedRights: []int{1, 2},
		},
		4: {
			UserGroups: []int{4},

			AllowedRights: []int{1, 4, 6},
		},
		5: {
			UserGroups: []int{5},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
		},
	}
	rv := CountUserAccess(1, rgMap, ugMap)
	assert.Equal(t, rv, RightValues{
		1: RightValTypeRestricted,
		2: RightValTypeAllow,
		3: RightValTypeAllow,
		4: RightValTypeAllow,
		5: RightValTypeRestricted,
		6: RightValTypeRestricted,
	})
}

// Типичный расчет прав для пользователя
func TestCountRights(t *testing.T) {
	ugMap := map[int]UserGroup{
		1: {Id: 1, GroupsIds: []int{2, 3}, UsersIds: []int{1}},
		2: {Id: 2, GroupsIds: []int{4, 5}},
		3: {Id: 3},
		4: {Id: 4},
		5: {Id: 5, UsersIds: []int{1}},
	}
	rgMap := map[int]RightGroup{
		1: {
			UserGroups: []int{1},

			AllowedRights:    []int{},
			RestrictedRights: []int{6},
		},
		2: {
			UserGroups: []int{2},

			RestrictedRights: []int{5},
		},
		3: {
			UserGroups: []int{3},

			AllowedRights: []int{1, 2},
		},
		4: {
			UserGroups: []int{4},

			AllowedRights: []int{1, 4, 6},
		},
		5: {
			UserGroups: []int{5},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
		},
	}
	rv := CountUserAccess(1, rgMap, ugMap)
	assert.Equal(t, rv, RightValues{
		1: RightValTypeRestricted,
		2: RightValTypeAllow,
		3: RightValTypeAllow,
		4: RightValTypeAllow,
		5: RightValTypeRestricted,
		6: RightValTypeRestricted,
	})
}

// Проверка расчета  прав на объекты типового случая
func TestCountObjectsRights(t *testing.T) {
	ugMap := map[int]UserGroup{
		1: {Id: 1, GroupsIds: []int{2, 3}, UsersIds: []int{1}},
		2: {Id: 2, GroupsIds: []int{4, 5}},
		3: {Id: 3},
		4: {Id: 4},
		5: {Id: 5, UsersIds: []int{1}},
		6: {Id: 6},
	}
	rgMap := map[int]RightGroup{
		1: {
			UserGroups: []int{1},

			AllowedRights:    []int{},
			RestrictedRights: []int{6},
			ObjectsRights: []ObjectRights{
				{ObjectId: 1, ObjectTypeId: 31, Allowed: []int{4001, 4002}},
				{ObjectId: 2, ObjectTypeId: 31, Allowed: []int{4001, 4002}},
				{ObjectId: 10, ObjectTypeId: 40, Allowed: []int{5001, 5002}}},
		},
		2: {
			UserGroups: []int{2},

			RestrictedRights: []int{5},
		},
		3: {
			UserGroups: []int{3},

			AllowedRights: []int{1, 2},
			ObjectsRights: []ObjectRights{
				{ObjectId: 1, ObjectTypeId: 31, Allowed: []int{}, Restricted: []int{4002, 4003}},
				{ObjectId: 10, ObjectTypeId: 40, Allowed: []int{5004}}},
		},
		4: {
			UserGroups: []int{4},

			AllowedRights: []int{1, 4, 6},
		},
		5: {
			UserGroups: []int{5},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
			ObjectsRights: []ObjectRights{
				{ObjectId: 5, ObjectTypeId: 33, Allowed: []int{}, Restricted: []int{4002, 4003}}},
		},
		6: {
			UserGroups: []int{6},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
			ObjectsRights: []ObjectRights{
				{ObjectId: 15, ObjectTypeId: 33, Allowed: []int{}, Restricted: []int{4002, 4003}}},
		},
	}
	// Проверим права на объект 1
	rv := CountUserObjectAccess(1 /*userId*/, 1, 31, rgMap, ugMap)
	assert.Equal(t, RightValues{
		4001: RightValTypeAllow,
		4002: RightValTypeRestricted,
		4003: RightValTypeRestricted,
	}, rv)

	// Проверим права на объект 2
	rv = CountUserObjectAccess(1 /*userId*/, 2, 31, rgMap, ugMap)
	assert.Equal(t, RightValues{
		4001: RightValTypeAllow,
		4002: RightValTypeAllow,
	}, rv)

	// Проверим права на объект 10
	rv = CountUserObjectAccess(1 /*userId*/, 10, 40, rgMap, ugMap)
	assert.Equal(t, RightValues{
		5001: RightValTypeAllow,
		5002: RightValTypeAllow,
		5004: RightValTypeAllow,
	}, rv)

	// Проверим права на объект 10 но укажем некорретный тип
	rv = CountUserObjectAccess(1 /*userId*/, 10, 41, rgMap, ugMap)
	assert.Equal(t, RightValues{}, rv)

	// Проверим права на объект 15, пользователь не в группе где указан этот объект
	rv = CountUserObjectAccess(1 /*userId*/, 15, 33, rgMap, ugMap)
	assert.Equal(t, RightValues{}, rv)

	// Проверим права на объект 5
	rv = CountUserObjectAccess(1 /*userId*/, 5, 33, rgMap, ugMap)
	assert.Equal(t, RightValues{
		4002: RightValTypeRestricted,
		4003: RightValTypeRestricted,
	}, rv)
}

type testObject struct {
	id   int
	name string
}

func (t testObject) GetRightObjectId() int {
	return t.id
}

// Проверка расчета  прав на объекты типового случая
func TestFilterObjectsRights(t *testing.T) {
	ugMap := map[int]UserGroup{
		1: {Id: 1, GroupsIds: []int{2, 3}, UsersIds: []int{1}},
		2: {Id: 2, GroupsIds: []int{4, 5}},
		3: {Id: 3},
		4: {Id: 4},
		5: {Id: 5, UsersIds: []int{1}},
		6: {Id: 6},
	}
	rgMap := map[int]RightGroup{
		1: {
			UserGroups: []int{1},

			AllowedRights:    []int{},
			RestrictedRights: []int{6},
			ObjectsRights: []ObjectRights{
				{ObjectId: 1, ObjectTypeId: 31, Allowed: []int{4001, 4002}},
				{ObjectId: 2, ObjectTypeId: 31, Allowed: []int{4001, 4002}},
				{ObjectId: 10, ObjectTypeId: 40, Allowed: []int{5001, 5002}}},
		},
		2: {
			UserGroups: []int{2},

			RestrictedRights: []int{5},
		},
		3: {
			UserGroups: []int{3},

			AllowedRights: []int{1, 2},
			ObjectsRights: []ObjectRights{
				{ObjectId: 1, ObjectTypeId: 31, Allowed: []int{}, Restricted: []int{4002, 4003}},
				{ObjectId: 10, ObjectTypeId: 40, Allowed: []int{5004}}},
		},
		4: {
			UserGroups: []int{4},

			AllowedRights: []int{1, 4, 6},
		},
		5: {
			UserGroups: []int{5},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
			ObjectsRights: []ObjectRights{
				{ObjectId: 5, ObjectTypeId: 33, Allowed: []int{}, Restricted: []int{4002, 4003}}},
		},
		6: {
			UserGroups: []int{6},

			AllowedRights:    []int{2, 3, 5},
			RestrictedRights: []int{1},
			ObjectsRights: []ObjectRights{
				{ObjectId: 15, ObjectTypeId: 33, Allowed: []int{}, Restricted: []int{4002, 4003}}},
		},
	}
	// Обычная выборка объектов
	assert.Equal(t, []testObject{testObject{1, "111"}, testObject{2, "222"}},
		FilterObjectsByAvailabilityLocal([]testObject{
			{1, "111"},
			{2, "222"},
			{3, "333"},
			{4, "444"},
		}, 31, 1, rgMap, ugMap))
	// Таких объектов нет в правах - не тот тип
	assert.Equal(t, []testObject{},
		FilterObjectsByAvailabilityLocal([]testObject{
			{1, "111"},
			{2, "222"},
			{3, "333"},
			{4, "444"},
		}, 32, 1, rgMap, ugMap))
	// Объект запрещен
	assert.Equal(t, []testObject{},
		FilterObjectsByAvailabilityLocal([]testObject{
			{5, "555"},
			{6, "666"},
		}, 33, 1, rgMap, ugMap))
}
