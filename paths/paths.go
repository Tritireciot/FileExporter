package paths

// Subsystem paths
const (
	SubAccessTokenTag = "SubToken" // Значение заголовка токена для межпод	системного взаимодействия
	//------------
	// Запросы к Core
	//------------
	CoreGetUserGroups       = "http://%s/api/core/svc/user/group" // GET
	CoreGetUsers            = "http://%s/api/core/svc/user"       // GET
	CoreGetUser             = "http://%s/api/core/svc/user/%d"    // GET
	CoreGetUserMetaData     = "http://%s/api/core/svc/user/%d/meta/category/%d"
	CoreGetUserNotification = "http://%s/api/core/svc/user/%d/meta/notification" // GET - notification data
	// -----------
	// Straight to CONFIGURATION
	// -----------
	// Получить права пользователя
	// 1. путь до точки доступа
	// 2. имя подсистемы (news)
	// 3. ИД пользователя
	ConfigGetUserRights = "http://%s/api/%s/config/svc/user/%d/rights" // GET
	// Проверить право у пользователя
	// 1. путь до точки доступа
	// 2. имя подсистемы (news)
	// 3. ИД пользователя
	// 4. ИД права
	ConfigCheckUserRight = "http://%s/api/%s/config/svc/check/user/%d/right/%d" // GET
	// Проверить право на объект у пользователя
	// 1. путь до точки доступа
	// 2. имя подсистемы (news)
	// 3. ИД пользователя
	// 4. ИД объекта
	// 5. ИД права
	// 6. ИД типа объекта
	ConfigCheckUserObjectRight = "http://%s/api/%s/config/svc/check/user/%d/object/%d/right/%d?tid=%d" // GET
	// -----------
	// Straight to Storage
	// -----------
	// params: 1 - address; 2 - subsystem; 3 - storageId
	StorageCreateFilePath = "http://%s/api/%ssvc/storage/%d/file"         // POST
	StorageUpdateFilePath = "http://%s/api/%ssvc/storage/%d/file"         // PUT
	StorageDeleteFilePath = "http://%s/api/%ssvc/storage/%d/file"         // DELETE
	StorageGetFilePath    = "http://%s/api/%ssvc/storage/%d/file/content" // POST - Получение содержимого файла

	StorageCreateObjectPath = "http://%s/api/%ssvc/storage/object?name=%s&storageId=%d&c=%d" // POST
	StorageUpdateObjectPath = "http://%s/api/%ssvc/storage/object?id=%s&storageId=%d&c=%d"   // PUT
	StorageDeleteObjectPath = "http://%s/api/%ssvc/storage/object?id=%s&storageId=%d&c=%d"   // DELETE
	StorageGetObjectPath    = "http://%s/api/%ssvc/storage/object/content"                   // POST

	StorageFileInfoPath = "http://%s/api/%sstorage/%d/info" // POST

	StorageObjectInfoPath = "http://%s/api/%ssvc/storage/%d/info" // POST

	StorageBrowse = "http://%s/api/%ssvc/storage/%d/browse" // GET

	StorageWatchToSource    = "http://%s/api/%ssvc/storage/watch/source"     // POST - Установить(заменить) слежение за объектами по источнику
	StorageWatchAddToSource = "http://%s/api/%ssvc/storage/watch/source/add" // POST - Добавить элементы слежения за объектами по источнику
	StorageWatchTo          = "http://%s/api/%ssvc/storage/watch"            // POST - Следить за элементом плейлиста
	StorageStopWatchTo      = StorageWatchTo                                 // DELETE - Следить за элементом плейлиста

	StorageAddCopy    = "http://%s/api/%ssvc/storage/copy"                 // POST - Команда на фрмирование задания копирования
	StorageTaskStatus = "http://%s/api/%ssvc/storage/files/task/%s/status" // GET - Команда на получения статуса файлового задания
	StorageTaskStop   = "http://%s/api/%ssvc/storage/files/task/%s/stop"   // POST - остановка файлового задания
	StorageTaskChange = "http://%s/api/%ssvc/storage/files/task/%s"        // PATCH - изменение файлового задания

	// -----------
	// Straight to AIR
	// -----------
	AirTractSetPlaylist    = "http://%s/api/air/svc/tract/%d/playlist" // POST - Установить плейлист на тракте
	AirTractUpdatePlaylist = "http://%s/api/air/svc/tract/%d/playlist" // PUT - Обновить плейлист на тракте
	AirTractGetPlaylist    = "http://%s/api/air/svc/tract/%d/playlist" // GET - Получить плейлист на тракте
	AirGetTracts           = "http://%s/api/air/svc/tract"             // GET - получить список трактов
	AirGetUsedTracts       = "http://%s/api/air/svc/tract/used"        // GET - получить список используемых трактов
	// -------------
	// Configuration
	// -------------
	ConfigUpdateService = "http://%s/api/%sconfig/svc/service/%d" // PUT - Обновить службу

	// ------------------
	// FileProcessSystem
	// ------------------
	FileProcessingSystemCopyTask       = "http://%s/api/files/fms/task/copy"      // POST - Создание задания на копирование
	FileProcessingSystemCopyTaskChange = "http://%s/api/files/fms/task/change"    // POST - Изменения задания копирования
	FileProcessingSystemDeleteTask     = "http://%s/api/files/fms/task/delete"    // POST - Создание задания на удаление
	FileProcessingSystemBulk           = "http://%s/api/files/fms/task/bulk"      // POST - Массовые операции
	FileProcessingSystemStatus         = "http://%s/api/files/fms/tasks/status"   // GET - Статус заданий
	FileProcessingSystemTaskStatus     = "http://%s/api/files/fms/task/%s/status" // GET - Статус задания
	//---------------------
	// Studio VSM Paths
	StudioVSMPlaylistCurrentItems = "http://%s/api/air/playlist/items/current"

	StudioVSMAirStatus      = "http://%s/api/air/status"
	StudioVSMItemsNewsState = "http://%s/api/air/playlist/items/news/state"
	StudioMOSConnect        = "http://%s/api/studio/mos/svc/connect"
	StudioMOSDisconnect     = "http://%s/api/studio/mos/svc/disconnect"

	// --------------------
	// Air VSM Paths
	AirVSMConfigurationUpdate = "http://%s/api/configuration/update"
	AirVSMConfigurationGet    = "http://%s/api/configuration/get"

	AirVSMChannelUse  = "http://%s/api/channel/use"
	AirVSMChannelFree = "http://%s/api/channel/free"
	AirVSMChannelSync = "http://%s/api/channel/sync"

	AirVSMTractUse  = "http://%s/api/tract/use"
	AirVSMTractFree = "http://%s/api/tract/free"
	AirVSMTractSync = "http://%s/api/tract/sync"

	AirVSMAirPlaylistUpdate            = "http://%s/api/air/playlist/update"
	AirVSMAirPlaylistItemsSecondaryAdd = "http://%s/api/air/playlist/items/secondary/add"
	AirVSMAirPlaylistItemsAdd          = "http://%s/api/air/playlist/items/add"
	AirVSMAirPlaylistItemsUpdate       = "http://%s/api/air/playlist/items/update"
	AirVSMAirPlaylistItemsDelete       = "http://%s/api/air/playlist/items/delete"
	AirVSMAirPlaylistItemsMove         = "http://%s/api/air/playlist/items/move"

	AirVSMPreviewStart    = "http://%s/api/air/tract/preview/start"
	AirVSMPreviewStop     = "http://%s/api/air/tract/preview/stop"
	AirVSMPreviewPause    = "http://%s/api/air/tract/preview/pause"
	AirVSMPreviewSweep    = "http://%s/api/air/tract/preview/sweep"
	AirVSMPreviewSeek     = "http://%s/api/air/tract/preview/seek"
	AirVSMPreviewState    = "http://%s/api/air/tract/preview/state"
	AirVSMPreviewPlaylist = "http://%s/api/air/tract/preview/playlist"

	AirVSMAirPlaylistClean = "http://%s/api/air/playlist/clean"
	AirVSMAirPlaylistGet   = "http://%s/api/air/playlist/get"
	AirVSMAirPlaylistSet   = "http://%s/api/air/playlist/set"

	AirVSMAirStart        = "http://%s/api/air/start"
	AirVSMAirStartFromNow = "http://%s/api/air/startfromnow"
	AirVSMAirPause        = "http://%s/api/air/pause"
	AirVSMAirStop         = "http://%s/api/air/stop"
	AirVSMAirPlayNext     = "http://%s/api/air/playnext"
	AirVSMAirPlayPrev     = "http://%s/api/air/playprev"
	AirVSMAirModeSet      = "http://%s/api/air/mode/set"
	AirVSMAirModeGet      = "http://%s/api/air/mode/get"
	AirVSMAirStatus       = "http://%s/api/air/status/air"
	AirVSMAirSetState     = "http://%s/api/air/state/set"
	AirVSMWorkStatus      = "http://%s/api/air/status/work"

	AirVSMAirItemStateUpdate = "http://%s/api/air/item/state/update"

	AirVSMAirJog     = "http://%s/api/air/jog"
	AirVSMAirShuttle = "http://%s/api/air/shuttle"
	AirVSMAirSeek    = "http://%s/api/air/seek"

	AirVSMGroupUse          = "http://%s/api/group/use"
	AirVSMGroupFree         = "http://%s/api/group/free"
	AirVSMGroupStateMembers = "http://%s/api/group/state/members"
	AirVSMGroupStateAir     = "http://%s/api/group/state/air"
	AirVSMGroupStateWork    = "http://%s/api/group/state/work"
	AirVSMGroupSyncMembers  = "http://%s/api/group/sync/members"
	AirVSMGroupSyncSettings = "http://%s/api/group/sync/settings"

	AirVSMGroupMemberAdd    = "http://%s/api/group/member/add"
	AirVSMGroupMemberRemove = "http://%s/api/group/member/remove"
	AirVSMGroupDisband      = "http://%s/api/group/member/disband"
	//------------
	// Запросы к Catalogs
	//------------
	CatalogRegisterItem = "http://%s/api/catalogs/svc/catalog/%d/item" // // POST - Регистрация объекта в каталоге
)
