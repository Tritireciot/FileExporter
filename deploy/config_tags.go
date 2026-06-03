package deploy

// Переменные среды:
// CFG_METHOD - Метод получения конфигурации
// Возможные:
// * zk
// * ini
// CFG_PATH - путь до источника конфигурации. Может быть ip адрес, а может путь до ini
const CFG_METHOD_TAG = "CFG_METHOD"
const CFG_PATH_TAG = "CFG_PATH"
const CFG_ZK_METHOD = "zk"
const CFG_PSQL_METHOD = "psql"
const CFG_INI_METHOD = "ini"
