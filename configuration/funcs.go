package config

import (
	"AutoplayX/deploy"
	"AutoplayX/jwtoken"
	"AutoplayX/pgutils"
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mitchellh/mapstructure"
)

// Структура изменения какого-то объекта которая возвращает psql
type ChangeObjectStruct struct {
	Id    int    `json:"id,omitempty"`
	Ver   int    `json:"version,omitempty"`
	Error string `json:"error,omitempty"`
}

func (o *ChangeObjectStruct) IsError() bool {
	return len(o.Error) > 0
}

func (o *ChangeObjectStruct) GetError() error {
	if o.IsError() {
		return fmt.Errorf(o.Error)
	} else {
		return nil
	}
}

// -------------------------------------------------------------------------------------
// Сервисы
func AddService(p *pgxpool.Pool, service Service) (int /*id*/, int /*ver*/, error) {
	ans := ""
	if err := p.QueryRow(context.Background(), `SELECT * FROM config."AddService"($1,$2,$3,$4)`, service.Props.Name, service.Props.TypeID, service.Props.ServerID, service.Configuration).Scan(&ans); err != nil {
		return 0, 0, err
	} else {
		a := ChangeObjectStruct{}
		json.Unmarshal([]byte(ans), &a)
		if a.Id == 0 {
			return 0, 0, a.GetError()
		}
		return a.Id, a.Ver, err
	}
}

// ---------------------------------------------------
func AddServiceTx(pool *pgxpool.Pool, service Service) (int /*id*/, int /*ver*/, error) {
	cfg := "[ ]"
	if service.Configuration != nil {
		ba, _ := json.Marshal(service.Configuration)
		cfg = string(ba)
	}

	ans := ""
	if err := pool.QueryRow(context.Background(), `SELECT * FROM config."AddService"($1,$2,$3,$4)`, service.Props.Name, service.Props.TypeID, service.Props.ServerID, cfg).Scan(&ans); err != nil {
		return 0, 0, err
	} else {
		a := ChangeObjectStruct{}
		json.Unmarshal([]byte(ans), &a)
		if a.Id == 0 {
			return 0, 0, a.GetError()
		}
		return a.Id, a.Ver, err
	}
}

// ---------------------------------------------------
func UpdateServiceTx(tx pgx.Tx, service Service) (int /*ver*/, error) {
	co := ChangeObjectStruct{}
	err := tx.QueryRow(context.Background(), `SELECT * FROM config."UpdateService"($1,$2,$3,$4,$5)`, service.ID, service.Props.ServerID, service.Props.TypeID, service.Props.Name, service.Configuration).Scan(&co)
	return co.Ver, err
}

// ---------------------------------------------------
func GetService(p *pgxpool.Pool, serviceId int) (s Service, err error) {
	jcfg := pgutils.NullString{}
	if err = p.QueryRow(context.Background(), `SELECT * FROM config."GetService"($1)`, serviceId).Scan(&jcfg); err == nil && jcfg.Valid {
		err = json.Unmarshal([]byte(jcfg.Str), &s)
	}

	return s, err
}

// ---------------------------------------------------
func GetServices(p *pgxpool.Pool) ([]Service, int, error) {
	jcfg := pgutils.NullString{}
	if err := p.QueryRow(context.Background(), `SELECT * FROM config."GetServices"()`).Scan(&jcfg); err == nil && jcfg.Valid {
		resp := struct {
			Services []Service `json:"services"`
			Ver      int       `json:"ver"`
		}{}

		err = json.Unmarshal([]byte(jcfg.Str), &resp)
		return resp.Services, resp.Ver, err
	} else {
		return []Service{}, 0, err
	}
}

// ---------------------------------------------------
func GetServicesMap(p *pgxpool.Pool) (map[int]Service, int, error) {
	jcfg := pgutils.NullString{}
	if err := p.QueryRow(context.Background(), `SELECT * FROM config."GetServices"()`).Scan(&jcfg); err == nil && jcfg.Valid {
		resp := struct {
			Services []Service `json:"services"`
			Ver      int       `json:"ver"`
		}{}

		err = json.Unmarshal([]byte(jcfg.Str), &resp)

		smap := map[int]Service{}
		for i := range resp.Services {
			smap[resp.Services[i].ID] = resp.Services[i]
		}
		return smap, resp.Ver, err
	} else {
		return map[int]Service{}, 0, err
	}
}

// ---------------------------------------------------
// Получить все службы с определнным типом
func GetServicesByType(p *pgxpool.Pool, st ServiceTypeID) ([]Service, error) {
	jcfg := pgutils.NullString{}
	if err := p.QueryRow(context.Background(), `SELECT * FROM config."GetServicesByType"($1)`, st).Scan(&jcfg); err == nil && jcfg.Valid {
		resp := struct {
			Services []Service `json:"services"`
			Ver      int       `json:"ver"`
		}{}

		err = json.Unmarshal([]byte(jcfg.Str), &resp)

		return resp.Services, err

	} else {
		return []Service{}, err
	}
}

func DeleteService(pool *pgxpool.Pool, id int) (ver int, err error) {
	ans := ""
	if err := pool.QueryRow(context.Background(), `SELECT * FROM config."DeleteService"($1)`, id).Scan(&ans); err != nil {
		return 0, err
	} else {
		r := ChangeObjectStruct{}
		if err := json.Unmarshal([]byte(ans), &r); err != nil {
			return 0, err
		} else if r.IsError() {
			return 0, r.GetError()
		} else {
			return r.Ver, nil
		}
	}
}

// -------------------------------------------------------------------------------------
// Сервера
func AddServer(p *pgxpool.Pool, srv ServerProps) (int, int, error) {
	jcfg := ""
	if err := p.QueryRow(context.Background(), `SELECT * FROM config."AddServer"($1,$2)`, srv.Name, srv.Path).Scan(&jcfg); err != nil {
		return 0, 0, err
	} else {
		r := ChangeObjectStruct{}
		if err := json.Unmarshal([]byte(jcfg), &r); err != nil {
			return 0, 0, err
		} else {
			if r.Id == 0 {
				return 0, 0, r.GetError()
			}
			return r.Id, r.Ver, err
		}
	}
}

func DeleteServer(pool *pgxpool.Pool, id int) (ver int, err error) {
	ans := ""
	if err := pool.QueryRow(context.Background(), `SELECT * FROM config."DeleteServer"($1)`, id).Scan(&ans); err != nil {
		return 0, err
	} else {
		r := ChangeObjectStruct{}
		if err := json.Unmarshal([]byte(ans), &r); err != nil {
			return 0, err
		} else if r.IsError() {
			return 0, r.GetError()
		} else {
			return r.Ver, nil
		}
	}
}

// ---------------------------------------------------
func GetServer(pool *pgxpool.Pool, sid int) (cfg ServerContainer, err error) {
	jcfg := ""
	resp := struct {
		Server ServerContainer `json:"object"`
		Ver    int             `json:"ver"`
	}{}
	if err = pool.QueryRow(context.Background(), `SELECT * FROM config."GetServer"($1)`, sid).Scan(&jcfg); err == nil {
		err = json.Unmarshal([]byte(jcfg), &resp)
	}
	return resp.Server, err
}

// ---------------------------------------------------
func GetServers(pool *pgxpool.Pool) (cfg []ServerContainer, err error) {
	jcfg := ""
	resp := struct {
		Servers []ServerContainer `json:"servers"`
		Ver     int               `json:"ver"`
	}{}
	if err = pool.QueryRow(context.Background(), `SELECT * FROM config."GetServers"()`).Scan(&jcfg); err == nil {
		err = json.Unmarshal([]byte(jcfg), &resp)
	}
	return resp.Servers, err
}

// ---------------------------------------------------
func GetServersMap(pool *pgxpool.Pool) (m map[int]ServerContainer, err error) {
	servers, err := GetServers(pool)

	m = make(map[int]ServerContainer)
	for i := range servers {
		m[servers[i].Id] = servers[i]
	}
	return m, err
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

type InfrastructureConfiguration struct {
	SubAccess []deploy.AccessConfig    // Межподсистемное взаимодействие
	MQAccess  []deploy.AccessConfig    // Очередь
	DBConfig  []pgutils.DatabaseConfig // База данных

	PublicKey                 rsa.PublicKey // Публичный ключ проверки токенов
	IntenalCommunicationToken string        // Ключ проверки межсервисного взаимодействия
}

func GetDBConfigFromPath(path string) (pgutils.DatabaseConfig, error) {
	pathParams := strings.Split(path, ":")
	if len(pathParams) < 5 {
		return pgutils.DatabaseConfig{}, fmt.Errorf("invalid configuration format: %s! Use: <dbname>:<scheme>:<host>:<port>:<user>:<pass>", path)
	}

	if port, err := strconv.Atoi(pathParams[3]); err != nil {
		return pgutils.DatabaseConfig{}, err
	} else {
		return pgutils.DatabaseConfig{DBname: pathParams[0], Schema: pathParams[1], Host: pathParams[2], Port: port, User: pathParams[4], Password: pathParams[5]}, err
	}
}

func GetPrivateCert(dbcfg pgutils.DatabaseConfig) (pk rsa.PrivateKey, err error) {
	if dbconn, err := pgutils.OpenDBWithSchema(&dbcfg, dbcfg.Schema); err != nil {
		return pk, err
	} else {
		defer dbconn.Close(context.Background())

		//------------------------------------------------
		// Получим публичный ключ
		if val, _, err := GetSQLConfigValue(dbconn, CFG_CERT_PRIVATE_PSQL); err != nil {
			log.Print(err)
			return pk, err
		} else {
			pm, _ := pem.Decode([]byte(val))
			if pkey, err := jwtoken.LoadPrivateEM(pm); err != nil {
				log.Print(err)
				return pk, err
			} else {
				pk = *pkey
			}
		}
		return pk, err
	}
}

// --------------------------------------------------------------------------------------
// Получение конфигурации pathL: "<dbname>:<host>:<port>:<user>:<pass>:"
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

// Получить версию определенного объекта по его ИД
func GetObjectVersion(pool *pgxpool.Pool, objectVersionID int) (int, error) {
	version := 0
	err := pool.QueryRow(context.Background(), `SELECT "Version" FROM config."ObjectVersions" WHERE "ID"=$1`, objectVersionID).Scan(&version)
	return version, err
}

func IncObjectVersionTx(tx pgx.Tx, objectVersionID int) (int, error) {
	version := 0
	err := tx.QueryRow(context.Background(), `SELECT config."IncObjectVersion"($1)`, objectVersionID).Scan(&version)
	return version, err
}
