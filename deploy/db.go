package deploy

import (
	"AutoplayX/pgutils"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ScriptType int

const (
	CreateScriptType ScriptType = iota
	UpdateScriptType
	UpgradeScriptType
	DowngradeScriptType
)

const (
	UpdateTag    = "Update"
	UpgradeTag   = "Upgrade"
	DowngradeTag = "Downgrade"
)

type SQLToExecute struct {
	NameObject string
	Path       string

	scriptType ScriptType
}

var executes = []SQLToExecute{
	{NameObject: "Tables", Path: "Tables", scriptType: CreateScriptType},
	{NameObject: "Procedures", Path: "Procedures", scriptType: CreateScriptType},
	{NameObject: "Triggers", Path: "Triggers", scriptType: CreateScriptType},
	{NameObject: "Custom", Path: "Custom", scriptType: CreateScriptType},
	{NameObject: "Update", Path: UpdateTag, scriptType: UpdateScriptType},
	{NameObject: "Upgrade", Path: UpgradeTag, scriptType: UpgradeScriptType},
	{NameObject: "Downgrade", Path: DowngradeTag, scriptType: DowngradeScriptType},
}

func getExecutible(p string) (SQLToExecute, bool) {
	for _, e := range executes {
		if e.Path == p {
			return e, true
		}
	}
	return SQLToExecute{}, false
}

type Version struct {
	catalogName string
	version     int
}

func (v Version) isNull() bool {
	if v.version == 0 || len(v.catalogName) == 0 {
		return true
	}
	return false
}

func getVersions(relativePathPrefix string) []Version {
	versions := []Version{}
	if fs, err := os.ReadDir(relativePathPrefix); err == nil {
		for i := range fs {
			if fs[i].IsDir() {
				if ver, err := strconv.Atoi(fs[i].Name()); err != nil {
					log.Printf("Incorrect version name: %s! %v", fs[i].Name(), err)
				} else {
					versions = append(versions, Version{catalogName: fs[i].Name(), version: ver})
				}
			}
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].version < versions[j].version
	})

	return versions
}

func tagToVer(verTag string, versions []Version) Version {
	// Берем наибольшую версию?
	if verTag == pgutils.LATEST {
		if len(versions) == 0 {
			return Version{}
		}

		return versions[len(versions)-1]
	}

	for i := range versions {
		if verTag == fmt.Sprintf("%d", versions[i].version) || verTag == versions[i].catalogName {
			return versions[i]
		}
	}

	return Version{}
}

// -------------------------------------------------------------------------------------------------------
// Отобразить скрипты по пути
func ShowScriptsInfo(relativePathPrefix string) error {
	// Получим список "версий" по именам папок
	versions := getVersions(relativePathPrefix)
	// Отобразим список файлов и версий
	if len(versions) == 0 {
		return fmt.Errorf("no any versions at path: %s", relativePathPrefix)
	} else {
		for _, v := range versions {
			log.Print("------------------------------------------------------------------------------")
			log.Printf("Version: %v", v)
			if verForlders, err := os.ReadDir(fmt.Sprintf("%s/%s", relativePathPrefix, v.catalogName)); err == nil {
				for _, vf := range verForlders {
					if vf.IsDir() {
						if _, ok := getExecutible(vf.Name()); ok {
							log.Printf("\t%s\n", vf.Name())
							scriptsFiles := ""
							if scripts, err := os.ReadDir(fmt.Sprintf("%s/%s/%s", relativePathPrefix, v.catalogName, vf.Name())); err == nil {
								for _, sf := range scripts {
									scriptsFiles += sf.Name() + "; "
								}
							}
							log.Printf("\t\t%s", scriptsFiles)
						}
					}
				}
			}
		}
		return nil
	}
}

// -------------------------------------------------------------------------------------------------------
// Установить версию БД на
func updateDBVersion(tx pgx.Tx, ver int) error {
	_, err := tx.Exec(context.Background(), fmt.Sprintf(`UPDATE "DBVersion" SET "Version"=%d`, ver))
	return err
}

// -------------------------------------------------------------------------------------------------------
func CreateDBMethod(verTag string, dbConfig *pgutils.DatabaseConfig, relativePathPrefix string, opt pgutils.Options) error {
	if conn, err := pgutils.PrepareDBEnvironment(dbConfig, opt); err != nil {
		return err
	} else {
		defer conn.Close(context.Background())

		if tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{}); err == nil {
			// Получим список "версий" по именам папок
			versions := getVersions(relativePathPrefix)

			selVersion := tagToVer(verTag, versions)
			if selVersion.isNull() {
				return fmt.Errorf("can't select proper version for tag: %s", verTag)
			}

			for i := range versions {
				if versions[i] == selVersion {
					// Выполнение всех скриптов нужной версии
					for _, eo := range executes {
						if eo.scriptType == CreateScriptType {
							pathToScripts := fmt.Sprintf("%s/%s/%s", relativePathPrefix, selVersion.catalogName, eo.Path)
							pathToScripts = path.Clean(pathToScripts)
							log.Println("Execute " + eo.NameObject + "(" + pathToScripts + "):")
							if err := pgutils.ExecuteSqlScripts(tx, pathToScripts); err != nil {
								return err
							}
						}
					}
					//--------------------------------
					// Создаем версию БД

					if _, err := tx.Exec(context.Background(), `CREATE TABLE "DBVersion" ("Version" integer);`); err != nil {
						return err
					}
					if _, err := tx.Exec(context.Background(), fmt.Sprintf(`INSERT INTO "DBVersion" ("Version") VALUES (%d)`, selVersion.version)); err != nil {
						return err
					}

					err = tx.Commit(context.Background())

					if err == nil {
						log.Print("Create scripts successfuly done!")
					}

					return err
				}
			}
			return fmt.Errorf("not supported db version: %s", verTag)
		}
		return err
	}
}

// -------------------------------------------------------------------------------------------------------
func UpdateDBMethod(dbConfig *pgutils.DatabaseConfig, relativePathPrefix string, testing bool) error {
	if conn, err := pgutils.OpenDBWithSchema(dbConfig, dbConfig.Schema); err != nil {
		return err
	} else {
		defer conn.Close(context.Background())

		if ver, err := GetDBVersion(conn); err == nil {
			log.Printf("Current database version: %d", ver)
			if tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{}); err == nil {
				// Нужно произвести выполнение обновляемых скриптов текущей версии
				pathToScripts := path.Clean(fmt.Sprintf("%s/%d/%s", relativePathPrefix, ver, UpdateTag))
				if err = pgutils.ExecuteSqlScripts(tx, pathToScripts); err != nil {
					return err
				}
				// Применение изменений
				if testing {
					if err = tx.Rollback(context.Background()); err == nil {
						log.Print("Test PASSED. Scripts successful executed! Transaction declined.")
					}
				} else {
					err = tx.Commit(context.Background())
					if err == nil {
						log.Print("Successfull updated!")
					}
				}

				return err
			}
			return err
		} else {
			log.Print("Can't get current database version!")
		}

		return err
	}
}

// -------------------------------------------------------------------------------------------------------
// Обновление с текущей до запрашиваемой версии
func UpgradeDBMethod(toVerTag string, dbConfig *pgutils.DatabaseConfig, relativePathPrefix string, testing bool) error {
	if conn, err := pgutils.OpenDBWithSchema(dbConfig, dbConfig.Schema); err != nil {
		return err
	} else {
		defer conn.Close(context.Background())

		if curVer, err := GetDBVersion(conn); err == nil {
			log.Printf("Current database version: %d", curVer)
			// Произведем проверку
			// Получим список "версий" по именам папок
			versions := getVersions(relativePathPrefix)
			toVer := tagToVer(toVerTag, versions)
			if toVer.isNull() {
				return fmt.Errorf("can't select proper version for tag: %s", toVerTag)
			} else if curVer >= toVer.version {
				return fmt.Errorf("upgrade not needed! %v (current database version) >= %v (request upgdate version)", curVer, toVer)
			}

			if tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{}); err == nil {
				// Выполняем последоветльно с версии curVer до toVerTag
				executeThisScripts := false
				for i := range versions {
					if versions[i].version == curVer || curVer < versions[i].version && versions[i].version < toVer.version {
						executeThisScripts = true
					} else if versions[i].version < curVer {
						continue
					} else {
						break
					}
					// Выполнять ли текущую версию
					if executeThisScripts {
						log.Printf("Execute upgrade scripts for version: %v", versions[i])
						pathToScripts := path.Clean(fmt.Sprintf("%s/%s/%s", relativePathPrefix, versions[i].catalogName, UpgradeTag))
						if err = pgutils.ExecuteSqlScripts(tx, pathToScripts); err != nil {
							return err
						}
					}
				}
				// Применение изменений
				if testing {
					if err = tx.Rollback(context.Background()); err == nil {
						log.Print("Test PASSED. Scripts successful executed! Transaction declined.")
					}
				} else {
					if err = updateDBVersion(tx, toVer.version); err != nil {
						return err
					}
					err = tx.Commit(context.Background())
					if err == nil {
						log.Printf("Successfull upgraded to version: %v!", toVer)
					}
				}
				return err
			}
		} else {
			log.Print("Can't get current database version!")
		}

		return err
	}
}

// -------------------------------------------------------------------------------------------------------
// Понижение с текущей до запрашиваемой версии
func DowngradeDBMethod(toVerTag string, dbConfig *pgutils.DatabaseConfig, relativePathPrefix string, testing bool) error {
	if conn, err := pgutils.OpenDBWithSchema(dbConfig, dbConfig.Schema); err != nil {
		return err
	} else {
		defer conn.Close(context.Background())

		if curVer, err := GetDBVersion(conn); err == nil {
			log.Printf("Current database version: %d", curVer)
			// Произведем проверку
			// Получим список "версий" по именам папок
			versions := getVersions(relativePathPrefix)
			toVer := tagToVer(toVerTag, versions)
			if toVer.isNull() {
				return fmt.Errorf("can't select proper version for tag: %s", toVerTag)
			} else if curVer <= toVer.version {
				return fmt.Errorf("incorrect situation! %v (current database version) <= %v (request upgdate version)", curVer, toVer)
			}

			if tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{}); err == nil {
				// Выполняем последоветльно с версии curVer до toVerTag
				executeThisScripts := false
				for i := len(versions) - 1; i > 0; i-- {
					if versions[i].version == curVer || curVer > versions[i].version {
						executeThisScripts = true
					} else if versions[i].version <= toVer.version {
						break
					}
					// Выполнять ли текущую версию
					if executeThisScripts {
						log.Printf("Execute downgrade scripts for version: %v", versions[i])
						sp := path.Clean(fmt.Sprintf("%s/%s/%s", relativePathPrefix, versions[i].catalogName, DowngradeTag))

						if err = pgutils.ExecuteSqlScripts(tx, sp); err != nil {
							return err
						}
					}
				}
				// Применение изменений
				if testing {
					if err = tx.Rollback(context.Background()); err == nil {
						log.Print("Test PASSED. Scripts successful executed! Transaction declined.")
					}
				} else {
					if err = updateDBVersion(tx, toVer.version); err != nil {
						return err
					}

					err = tx.Commit(context.Background())
					if err == nil {
						log.Printf("Successfull downgraded to version: %v!", toVer)
					}
				}
				return err
			}
		} else {
			log.Print("Can't get current database version!")
		}

		return err
	}
}

// -------------------------------------------------------------------------------------------------------
func ExecuteScript(dbConfig pgutils.DatabaseConfig, script string, scan bool) error {
	if conn, err := pgutils.OpenDBWithSchema(&dbConfig, dbConfig.Schema); err != nil {
		return err
	} else {
		defer conn.Close(context.Background())

		if scan {
			log.Printf("Execute (query): %s", script)

			if rows, err := conn.Query(context.Background(), script); err != nil {
				return err
			} else {
				defer rows.Close()

				for rows.Next() {
					if vals, err := rows.Values(); err != nil {
						return err
					} else {
						log.Printf("Query row: %v", vals)
					}
				}
			}
		} else {
			log.Printf("Execute (exec): %s", script)

			_, err = conn.Exec(context.Background(), script)
		}

		return err
	}
}

// -------------------------------------------------------------------------------------------------------
func ExecuteScriptFile(dbConfig pgutils.DatabaseConfig, file string, scan bool) error {
	ba, err := os.ReadFile(file)
	if err != nil {
		return err
	} else {
		return ExecuteScript(dbConfig, string(ba), scan)
	}
}

// -------------------------------------------------------------------------------------------------------
func GetDBVersion(conn *pgx.Conn) (int, error) {
	ver := 0
	if err := conn.QueryRow(context.Background(), `SELECT "Version" from "DBVersion"`).Scan(&ver); err != nil {
		log.Fatal(err)
	}

	return ver, nil
}

// -------------------------------------------------------------------------------------------------------
func GetDBVersionPool(schema string, pool *pgxpool.Pool) (int, error) {
	ver := 0
	err := pool.QueryRow(context.Background(), fmt.Sprintf(`SELECT "Version" from %s."DBVersion"`, schema)).Scan(&ver)
	return ver, err
}
