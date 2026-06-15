package configuration

import (
	"PrintServer/deploy"
	"PrintServer/jwtoken"
	"PrintServer/pgutils"
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/mitchellh/mapstructure"
)

type InfrastructureConfiguration struct {
	SubAccess []deploy.AccessConfig    // Межподсистемное взаимодействие
	MQAccess  []deploy.AccessConfig    // Очередь
	DBConfig  []pgutils.DatabaseConfig // База данных

	PublicKey                 rsa.PublicKey // Публичный ключ проверки токенов
	IntenalCommunicationToken string        // Ключ проверки межсервисного взаимодействия
}

func GetDBConfigFromPath(path string) (pgutils.DatabaseConfig, error) {
	pathParams := strings.Split(path, ":")
	log.Print("Params: ", pathParams)
	if len(pathParams) < 5 {
		return pgutils.DatabaseConfig{}, fmt.Errorf("invalid configuration format: %s! Use: <dbname>:<scheme>:<host>:<port>:<user>:<pass>", path)
	}

	if port, err := strconv.Atoi(pathParams[3]); err != nil {
		return pgutils.DatabaseConfig{}, err
	} else {
		return pgutils.DatabaseConfig{DBname: pathParams[0], Schema: pathParams[1], Host: pathParams[2], Port: port, User: pathParams[4], Password: pathParams[5]}, err
	}
}

func GetPSQLCommonConfiguration(dbcfg pgutils.DatabaseConfig) (ic InfrastructureConfiguration, err error) {
	if dbconn, err := pgutils.OpenDBWithSchema(&dbcfg, dbcfg.Schema); err != nil {
		return ic, err
	} else {
		defer dbconn.Close(context.Background())
		// Получим конфигурации точки доступа
		if sc, err := GetServiceConfiguration(dbconn, AccessPointTypeID); err != nil {
			return ic, err
		} else {
			for i := range sc {
				ac := deploy.AccessConfig{}
				if err = mapstructure.Decode(sc[i].Configuration, &ac); err != nil {
					return ic, err
				}
				ic.SubAccess = append(ic.SubAccess, ac)
			}
		}
		//------------------------------------------------
		// Получим конфигурацию Очереди
		if sc, err := GetServiceConfiguration(dbconn, QueueTypeID); err != nil {
			return ic, err
		} else {
			for i := range sc {
				ac := deploy.AccessConfig{}
				if err = mapstructure.Decode(sc[i].Configuration, &ac); err != nil {
					return ic, err
				}
				ic.MQAccess = append(ic.MQAccess, ac)
			}
		}
		//------------------------------------------------
		// Получим конфигурацию БД
		if sc, err := GetServiceConfiguration(dbconn, PostgreSQLSearchTypeID); err != nil {
			return ic, err
		} else {
			for i := range sc {
				dc := pgutils.DatabaseConfig{}
				if err = mapstructure.Decode(sc[i].Configuration, &dc); err != nil {
					return ic, err
				}
				ic.DBConfig = append(ic.DBConfig, dc)
			}
		}
		//------------------------------------------------
		// Получим публичный ключ
		if val, _, err := GetSQLConfigValue(dbconn, CFG_CERT_PUBLIC_PSQL); err != nil {
			return ic, err
		} else {
			if pm, _ := pem.Decode([]byte(val)); pm == nil {
				return ic, fmt.Errorf("can't decode public key")
			} else {
				if pkey, err := jwtoken.LoadPublicEM(pm); err != nil {
					log.Print(err)
					return ic, err
				} else {
					ic.PublicKey = *pkey
				}
			}
		}
		//------------------------------------------------
		// Получим ключ внутреннего взаимодействия
		if val, _, err := GetSQLConfigValue(dbconn, CFG_SUBSYSTEM_COMM_KEY_PSQL); err != nil {
			return ic, err
		} else {
			ic.IntenalCommunicationToken = val
		}

	}
	return ic, nil
}

// ---------------------------------------------------
func GetServiceConfiguration(conn *pgx.Conn, serviceType ServiceTypeID) ([]serviceConfiguration, error) {
	var val pgutils.NullString
	if err := conn.QueryRow(context.Background(), `SELECT * FROM config."GetServicesByType"($1)`, serviceType).Scan(&val); err != nil {
		log.Print(err)
		return []serviceConfiguration{}, err
	} else {
		if val.Valid {
			resp := struct {
				Services []serviceConfiguration `json:"services"`
				Ver      int                    `json:"ver"`
			}{}

			err = json.Unmarshal([]byte(val.Str), &resp)
			return resp.Services, err
		} else {
			err = fmt.Errorf("configurations for service type: %d not found", serviceType)
			log.Print(err)
			return []serviceConfiguration{}, err
		}
	}
}

// --------------------------------------------------------------------------------------
func GetSQLConfigValue(conn *pgx.Conn, configString string) (string, int, error) {
	res := struct {
		Version int    `json:"ver"`
		Value   string `json:"value"`
		Error   string `json:"error"`
	}{}
	if err := conn.QueryRow(context.Background(), `SELECT config."GetConfiguration"($1)`, configString).Scan(&res); err != nil {
		log.Print(err)
		return "", 0, err
	} else {
		// Если нет ощибки
		if len(res.Error) == 0 {
			return res.Value, res.Version, nil
		} else {
			err = fmt.Errorf("can't get value %s. error: %s", configString, res.Error)
			log.Print(err)
			return "", res.Version, err
		}
	}
}