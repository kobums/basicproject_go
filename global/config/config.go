package config

// 설정은 환경변수가 우선이다 (.env 파일 → 환경변수 → .env.yml 순서로 채운다).
//   - .env      : KEY=VALUE 형식. 이미 설정된 환경변수는 덮어쓰지 않는다 (cp .env.example .env).
//   - 환경변수   : DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET, PORT, CORS_ORIGINS ...
//   - .env.yml  : (선택) 예전 방식의 모드별(develop/production) 설정. 환경변수가 없을 때만 쓴다.

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type DatabaseType int

// 생성 코드가 config.Database.Type 을 이 상수와 비교한다. 지원 DB 는 MySQL / MariaDB 뿐이다.
const (
	_ DatabaseType = iota
	Mysql
	Postgresql
	Sqlserver
)

type _Tls struct {
	Use  bool   `yaml:"use"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type _Mode struct {
	Port         string    `yaml:"port"`
	Tls          _Tls      `yaml:"tls"`
	UploadPath   string    `yaml:"path"`
	DocumentRoot string    `yaml:"documentRoot"`
	Mail         _Mail     `yaml:"mail"`
	Sms          _Sms      `yaml:"sms"`
	Cors         []string  `yaml:"cors"`
	Database     _Database `yaml:"database"`
	Log          _Log      `yaml:"log"`
	JwtSecret    string    `yaml:"jwtSecret"`
}

type _Database struct {
	Host             string       `yaml:"host"`
	Port             string       `yaml:"port"`
	Name             string       `yaml:"name"`
	User             string       `yaml:"user"`
	Password         string       `yaml:"password"`
	Type             DatabaseType `yaml:"-"`
	TypeString       string       `yaml:"type"`
	ConnectionString string       `yaml:"connectionString"`

	// 커넥션 풀 (models/db.go 의 공유 *sql.DB)
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"-"`
}

type _Mail struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Sender   string `yaml:"sender"`
}

type _Sms struct {
	User   string `yaml:"user"`
	Key    string `yaml:"key"`
	Sender string `yaml:"sender"`
}

type _Log struct {
	Level    string `yaml:"level"`
	Console  bool   `yaml:"console"`
	Web      bool   `yaml:"web"`
	Database bool   `yaml:"database"` // SQL 과 파라미터를 debug 로그로 남긴다 (bcrypt 해시 등 민감값 포함, 기본 off)
	File     string `yaml:"file"`
	Limit    struct {
		Size  int `yaml:"size"`
		Count int `yaml:"count"`
		Days  int `yaml:"days"`
	} `yaml:"limit"`
}

type Config struct {
	Version    string `yaml:"version"`
	Develop    _Mode  `yaml:"develop"`
	Production _Mode  `yaml:"production"`
}

// 개발 기본 CORS 허용 출처 (Vite / React / basicproject_web).
const DefaultCorsOrigins = "http://localhost:5173,http://localhost:3000,http://localhost:9007"

// JWT 비밀키 최소 길이 (HS256, 바이트).
const MinJwtSecretLength = 32

// 기본 토큰 만료 시간: 24시간 (Spring app.jwt.expiration-ms 와 같다).
const DefaultJwtExpiration = 24 * time.Hour

var Mail _Mail
var Database _Database
var Sms _Sms
var Tls _Tls
var Log _Log
var UploadPath string
var DocumentRoot string
var Version string
var Mode string
var Port string
var JwtSecret string
var JwtExpiration time.Duration
var Cors []string
var _value map[string]interface{}

func Init() {
	loadDotEnv(".env")

	config := &Config{}
	obj := make(map[string]interface{})

	if buf, err := os.ReadFile(".env.yml"); err == nil {
		if err := yaml.Unmarshal(buf, config); err != nil {
			log.Println(".env.yml:", err.Error())
		} else if err := yaml.Unmarshal(buf, obj); err != nil {
			log.Println(".env.yml:", err.Error())
		}
	}

	Mode = os.Getenv("APP_MODE")
	if len(os.Args) == 3 && os.Args[1] == "--mode" {
		Mode = os.Args[2]
	}
	if Mode != "production" {
		Mode = "develop"
	}

	mode := config.Develop
	if Mode == "production" {
		mode = config.Production
	}
	if v, ok := obj[Mode].(map[string]interface{}); ok {
		_value = v
	}

	Mail = mode.Mail
	Sms = mode.Sms
	UploadPath = mode.UploadPath
	DocumentRoot = mode.DocumentRoot
	Port = mode.Port
	Database = mode.Database
	Cors = mode.Cors
	Log = mode.Log
	Tls = mode.Tls
	JwtSecret = mode.JwtSecret
	Version = config.Version

	// 로그: .env.yml 에 log 항목이 없으면 기본값 (콘솔 + 파일, 개발은 debug / 운영은 info)
	if mode.Log == (_Log{}) {
		Log.Console = true
		Log.Web = true
		Log.File = "webdata/log/system.log"
		if Mode == "production" {
			Log.Level = "info"
		} else {
			Log.Level = "debug"
		}
	}
	setString(&Log.Level, "LOG_LEVEL")
	setBool(&Log.Console, "LOG_CONSOLE")
	setBool(&Log.Web, "LOG_WEB")
	setBool(&Log.Database, "LOG_DATABASE")
	if v, ok := os.LookupEnv("LOG_FILE"); ok {
		Log.File = strings.TrimSpace(v) // 빈 값이면 파일 로그를 끈다
	}
	if v := os.Getenv("LOG_DAYS"); v != "" {
		Log.Limit.Days, _ = strconv.Atoi(v)
	}
	if testing.Testing() {
		Log.File = "" // go test 가 패키지 디렉터리마다 webdata/log 를 만들지 않게 한다
	}

	// HTTP
	setString(&Port, "PORT")
	if Port == "" {
		Port = "8007"
	}
	setString(&UploadPath, "UPLOAD_PATH")
	if UploadPath == "" {
		UploadPath = "webdata"
	}
	setString(&DocumentRoot, "DOCUMENT_ROOT")
	if DocumentRoot == "" {
		DocumentRoot = "dist"
	}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		Cors = splitList(v)
	}
	if len(Cors) == 0 && Mode != "production" {
		Cors = splitList(DefaultCorsOrigins)
	}

	// DB (MySQL / MariaDB)
	setString(&Database.TypeString, "DB_TYPE")
	setString(&Database.Host, "DB_HOST")
	setString(&Database.Port, "DB_PORT")
	setString(&Database.User, "DB_USER")
	setString(&Database.Password, "DB_PASSWORD")
	setString(&Database.Name, "MARIADB_DATABASE") // docker-compose 와 같은 .env 를 쓸 때
	setString(&Database.Name, "DB_NAME")
	setString(&Database.ConnectionString, "DB_DSN")
	if Database.TypeString == "" {
		Database.TypeString = "mysql"
	}
	Database.Type = Mysql
	if Database.Host == "" {
		Database.Host = "127.0.0.1"
	}
	if Database.Port == "" {
		Database.Port = "3306"
	}
	if Database.ConnectionString == "" {
		// parseTime 을 켜지 않는다: 생성 모델은 날짜를 문자열(yyyy-MM-dd HH:mm:ss)로 읽는다.
		Database.ConnectionString = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_unicode_ci",
			Database.User, Database.Password, Database.Host, Database.Port, Database.Name)
	}
	Database.MaxOpenConns = envInt("DB_MAX_OPEN_CONNS", Database.MaxOpenConns, 25)
	Database.MaxIdleConns = envInt("DB_MAX_IDLE_CONNS", Database.MaxIdleConns, 10)
	Database.ConnMaxLifetime = time.Duration(envInt("DB_CONN_MAX_LIFETIME_SEC", 0, 300)) * time.Second

	// JWT (통일 JWT: HS256, JWT_SECRET)
	setString(&JwtSecret, "JWT_SECRET")
	JwtExpiration = DefaultJwtExpiration
	if v := os.Getenv("JWT_EXPIRATION_MS"); v != "" {
		if ms, err := strconv.ParseInt(v, 10, 64); err == nil && ms > 0 {
			JwtExpiration = time.Duration(ms) * time.Millisecond
		} else {
			log.Printf("JWT_EXPIRATION_MS=%q is not a positive integer, using %v", v, DefaultJwtExpiration)
		}
	}

	// Mail / SMS (선택)
	setString(&Mail.Server, "MAIL_SERVER")
	setString(&Mail.Port, "MAIL_PORT")
	setString(&Mail.User, "MAIL_USER")
	setString(&Mail.Password, "MAIL_PASSWORD")
	setString(&Mail.Sender, "MAIL_SENDER")
	setString(&Sms.User, "SMS_USER")
	setString(&Sms.Key, "SMS_KEY")
	setString(&Sms.Sender, "SMS_SENDER")

	// TLS (선택). 경로는 대소문자를 그대로 쓴다.
	if v := strings.TrimSpace(os.Getenv("TLS_CERT")); v != "" {
		Tls.Cert = v
	}
	if v := strings.TrimSpace(os.Getenv("TLS_KEY")); v != "" {
		Tls.Key = v
	}
	setBool(&Tls.Use, "TLS_USE")
	if Tls.Use {
		if Tls.Cert == "" {
			Tls.Cert = path.Join(UploadPath, "certs/ssl.crt")
		}
		if Tls.Key == "" {
			Tls.Key = path.Join(UploadPath, "certs/ssl.key")
		}
	}
}

// Validate 는 서버를 띄우기 전에 필수 설정을 확인한다 (main 에서 부른다).
func Validate() error {
	if len(JwtSecret) < MinJwtSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d bytes (got %d); set it in .env or the environment", MinJwtSecretLength, len(JwtSecret))
	}
	if Mode == "production" && strings.HasPrefix(JwtSecret, "change-me") {
		return errors.New("JWT_SECRET is still the .env.example placeholder; generate one (openssl rand -base64 48)")
	}
	if Database.Name == "" && os.Getenv("DB_DSN") == "" {
		return errors.New("DB_NAME is not set; set DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME in .env or the environment")
	}
	return nil
}

func Get(name string) interface{} {
	return _value[name]
}

func GetString(name string) string {
	s, _ := _value[name].(string)
	return s
}

func GetInt(name string) int {
	i, _ := _value[name].(int)
	return i
}

func setString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func setBool(dst *bool, key string) {
	switch strings.ToUpper(strings.TrimSpace(os.Getenv(key))) {
	case "Y", "YES", "T", "TRUE", "1", "ON":
		*dst = true
	case "N", "NO", "F", "FALSE", "0", "OFF":
		*dst = false
	}
}

func envInt(key string, current int, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		log.Printf("%s=%q is not an integer, ignored", key, v)
	}
	if current != 0 {
		return current
	}
	return def
}

func splitList(s string) []string {
	ret := make([]string, 0)
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			ret = append(ret, part)
		}
	}
	return ret
}

// loadDotEnv 는 KEY=VALUE 형식의 파일을 읽어 아직 없는 환경변수만 설정한다.
// 빈 줄, # 주석, export 접두사, 따옴표로 감싼 값을 지원한다.
func loadDotEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists && key != "" {
			os.Setenv(key, value)
		}
	}
}
