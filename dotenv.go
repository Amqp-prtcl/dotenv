package dotenv

import (
	"bufio"
	"io"
	"os"
	"strings"
)

var (
	envPath = ".env"

	Envs map[string]string
)

// SetEnvPath sets the path of the .env file.
//
// Default: .env in working directory
func SetEnvPath(path string) {
	envPath = path
}

func GetEnvPath() string {
	return envPath
}

func Get(key string) string {
	return Envs[key]
}

func Set(key string, value string) error {
	Envs[key] = value
	return os.Setenv(key, value)
}

// SetSave acts like Set(key string, value string)error but also
// concurrently updates env file with new values.
//
// Please Note that this function will not return an
// error if it fails during saving as it will happen
// in another coroutine
func SetSave(key string, value string, blocking ...bool) error {
	Envs[key] = value
	var err = os.Setenv(key, value)
	if len(blocking) != 0 && blocking[0] {
		SaveEnv()
	} else {
		go SaveEnv()
	}
	return err
}

func LoadEnv() error {
	Envs = map[string]string{}

	f, err := os.Open(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	eof := false
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}
			eof = true
		}
		trimSpaces(&line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		vals := strings.SplitN(line, "=", 2)
		if len(vals) != 2 {
			continue
		}
		Envs[vals[0]] = vals[1]
		if eof {
			break
		}
	}
	for k, v := range Envs {
		os.Setenv(k, v)
	}

	return nil
}

// SaveEnv saves all environment variables that were previously set
// in the .env file as well as any new environment variable that have
// been set with Set(key string, value string)error
func SaveEnv() error {
	f, err := os.Create(envPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range Envs {
		_, err := f.WriteString(k + "=" + v)
		if err != nil {
			return err
		}
	}
	return nil
}

func trimSpaces(buf *string) {
	l := len(*buf)
	st := 0
	for st < l && emptyChar(*buf, st) {
		st++
	}
	end := l - 1
	for end >= 0 && emptyChar(*buf, st) {
		end--
	}
	(*buf) = (*buf)[st : end+1]
}

func emptyChar(b string, i int) bool {
	switch b[i] {
	case ' ', '\r', '\t', '\n':
		return true
	default:
		return false
	}
}
