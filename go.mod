module PrintServer

go 1.19

require (
	AutoplayX v0.0.0-00010101000000-000000000000
	github.com/Masterminds/squirrel v1.5.4
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v4 v4.18.3
	github.com/lib/pq v1.10.2
	github.com/mmonterroca/docxgo/v2 v2.5.1
	github.com/stretchr/testify v1.11.1
	github.com/tdewolff/parse/v2 v2.8.13
	github.com/tidwall/gjson v1.18.0
	golang.org/x/net v0.21.0
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/segmentio/kafka-go v0.4.42
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/crypto v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/mmonterroca/docxgo/v2 => ./PrintServer/docxgo

replace AutoplayX => ./
