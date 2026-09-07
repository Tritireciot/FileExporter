package config

import (
	"AutoplayX/paths"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ---------------------------------------------------
// Запрос на получение пользовательских групп
func GetUserGroupsByRequest(subAccessPath string, subAccessToken string) (map[int]UserGroup, error) {
	client := &http.Client{}
	ugmap := make(map[int]UserGroup)
	req, err := http.NewRequest("GET", fmt.Sprintf(paths.CoreGetUserGroups, subAccessPath), nil)
	if err != nil {
		return ugmap, err
	}
	req.Header.Set(paths.SubAccessTokenTag, subAccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return ugmap, err
	} else if resp.StatusCode != http.StatusOK {
		return ugmap, fmt.Errorf("error when making sub get user groups request: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ugmap, err
	}
	ugVer := struct {
		Ver     int         `json:"version"`
		Ugroups []UserGroup `json:"userGroups"`
	}{}
	if err = json.Unmarshal(body, &ugVer); err != nil {
		return ugmap, err
	}
	for _, ug := range ugVer.Ugroups {
		ugmap[ug.Id] = ug
	}
	return ugmap, err
}

type FunctionRights struct {
	AllowedRights    []int `json:"allowedRights"`
	RestrictedRights []int `json:"restrictedRights"`
}

type UserRights struct {
	FunctionRights FunctionRights `json:"funcRights"`
	ObjectsRights  []ObjectRights `json:"objectsRights"`
}

// Получение списка ВСЕХ доступных прав пользователя.
func GetUserRights(userId int, subAccessPath string, subAccessToken string, subsystemName string) (UserRights, error) {
	client := &http.Client{}
	rdata := UserRights{}
	url := fmt.Sprintf(paths.ConfigGetUserRights, subAccessPath, subsystemName, userId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return rdata, err
	}
	req.Header.Set(paths.SubAccessTokenTag, subAccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return rdata, err
	} else if resp.StatusCode != http.StatusOK {
		return rdata, fmt.Errorf("error when making sub request %s: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rdata, err
	}
	err = json.Unmarshal(body, &rdata)
	return rdata, err
}

// Проверка функционального права пользователя
func CheckUserRight(userId int, rightId int, subAccessPath string, subAccessToken string, subsystemName string) (bool, error) {
	client := &http.Client{}
	url := fmt.Sprintf(paths.ConfigCheckUserRight, subAccessPath, subsystemName, userId, rightId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set(paths.SubAccessTokenTag, subAccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		return false, nil
	default:
		return false, fmt.Errorf("error when making sub request %s: %d", url, resp.StatusCode)
	}
}

// Проверка функционального права пользователя
func CheckUserObjectRight(userId int, objectId int, objectTypeId int, rightId int, subAccessPath string, subAccessToken string, subsystemName string) (bool, error) {
	client := &http.Client{}
	url := fmt.Sprintf(paths.ConfigCheckUserObjectRight, subAccessPath, subsystemName, userId, objectId, rightId, objectTypeId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set(paths.SubAccessTokenTag, subAccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized:
		return false, nil
	default:
		return false, fmt.Errorf("error when making sub request %s: %d", url, resp.StatusCode)
	}
}
