package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rajendraventurit/radicaapi/lib/logger"
)

var otherScripts = []string{
	"org_types.sql",
	"roles.sql",
	"job_categories.sql",
	"jobs.sql",
	"permissions.sql",
}

// Migrate will run migration scripts up to version
func Migrate(db *sqlx.DB, spath string, target float64) error {
	_ = makeVersionTable(db) // ignore error
	if !strings.HasSuffix(spath, "/") {
		spath += "/"
	}
	current, err := getVersion(db)
	if err != nil {
		return fmt.Errorf("getting db version %v", err)
	}

	if current >= target {
		return nil
	}

	migs, err := walkPath(spath, target, current)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	for _, ms := range migs {
		logger.Messagef("Processing version %v\n", ms.ver)
		err := execFile(tx, ms.path)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("Migrate version %v %v", ms.ver, err)
		}
		if err := setVersion(tx, ms.ver); err != nil {
			err := tx.Rollback()
			return fmt.Errorf("Migrate setVersion %v", err)
		}
	}

	// run other scripts
	for _, s := range otherScripts {
		path := spath + s
		_, err := sqlx.LoadFile(tx, path)
		if err != nil {
			err := tx.Rollback()
			return fmt.Errorf("Migrate script %v %v", path, err)
		}
	}
	return tx.Commit()
}

func makeVersionTable(db *sqlx.DB) error {
	str := `
	CREATE TABLE _db_versions (
		ver_id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
		version DECIMAL(10, 5) NOT NULL,
		created_on timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB CHARSET=utf8
	`
	_, err := db.Exec(str)
	if err != nil {
		return err
	}
	return setVersion(db, 0.0)
}

func setVersion(db Execer, ver float64) error {
	str := "INSERT INTO _db_versions (version) VALUES (?)"
	_, err := db.Exec(str, ver)
	return err
}

func getVersion(db Queryer) (float64, error) {
	str := `
	SELECT version
	FROM _db_versions
	ORDER BY version DESC
	LIMIT 1
	`
	v := float64(0.0)
	err := db.Get(&v, str)
	return v, err
}

func execFile(db Execer, fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	file, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		request = strings.TrimSpace(request)
		if request == "" {
			continue
		}
		_, err := db.Exec(request)
		if err != nil {
			return fmt.Errorf("%v %v", request, err)
		}
	}
	return nil
}

type vScript struct {
	ver  float64
	path string
}

type sortVersion []vScript

func (sv sortVersion) Len() int           { return len(sv) }
func (sv sortVersion) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv sortVersion) Less(i, j int) bool { return sv[i].ver < sv[j].ver }

var regX = regexp.MustCompile(`version_(\d+)_(\d+).sql`)

func walkPath(spath string, target, current float64) ([]vScript, error) {
	migs := []vScript{}
	err := filepath.Walk(spath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		match := regX.FindAllStringSubmatch(path, -1)
		if len(match) == 0 || len(match[0]) != 3 {
			return nil
		}
		vs := fmt.Sprintf("%s.%s", match[0][1], match[0][2])
		fver, err := strconv.ParseFloat(vs, 64)
		if err != nil {
			return err
		}
		if fver > current && fver <= target {
			migs = append(migs, vScript{ver: fver, path: path})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", spath, err)
	}
	sort.Sort(sortVersion(migs))
	return migs, nil
}
