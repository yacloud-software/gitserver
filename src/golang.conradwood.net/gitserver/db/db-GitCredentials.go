package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBGitCredentials
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence gitcredentials_seq;

Main Table:

 CREATE TABLE gitcredentials (id integer primary key default nextval('gitcredentials_seq'),userid text not null  ,host text not null  ,path text not null  ,username text not null  ,password text not null  ,expiry integer not null  );

Alter statements:
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS host text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS path text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS username text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS password text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS expiry integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE gitcredentials_archive (id integer unique not null,userid text not null,host text not null,path text not null,username text not null,password text not null,expiry integer not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBGitCredentials *DBGitCredentials
)

type DBGitCredentials struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBGitCredentials()
	})
}

func DefaultDBGitCredentials() *DBGitCredentials {
	if default_def_DBGitCredentials != nil {
		return default_def_DBGitCredentials
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGitCredentials(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGitCredentials = res
	return res
}
func NewDBGitCredentials(db *sql.DB) *DBGitCredentials {
	foo := DBGitCredentials{DB: db}
	foo.SQLTablename = "gitcredentials"
	foo.SQLArchivetablename = "gitcredentials_archive"
	return &foo
}

func (a *DBGitCredentials) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBGitCredentials) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBGitCredentials) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBGitCredentials) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGitCredentials", "insert into "+a.SQLArchivetablename+" (id,userid, host, path, username, password, expiry) values ($1,$2, $3, $4, $5, $6, $7) ", p.ID, p.UserID, p.Host, p.Path, p.Username, p.Password, p.Expiry)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBGitCredentials) buildSaveMap(ctx context.Context, p *savepb.GitCredentials) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["host"] = a.get_col_from_proto(p, "host")
	res["path"] = a.get_col_from_proto(p, "path")
	res["username"] = a.get_col_from_proto(p, "username")
	res["password"] = a.get_col_from_proto(p, "password")
	res["expiry"] = a.get_col_from_proto(p, "expiry")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBGitCredentials) Save(ctx context.Context, p *savepb.GitCredentials) (uint64, error) {
	qn := "save_DBGitCredentials"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBGitCredentials) SaveWithID(ctx context.Context, p *savepb.GitCredentials) error {
	qn := "insert_DBGitCredentials"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBGitCredentials) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.GitCredentials) (uint64, error) {
	// Save (and use database default ID generation)

	var rows *gosql.Rows
	var e error

	q_cols := ""
	q_valnames := ""
	q_vals := make([]interface{}, 0)
	deli := ""
	i := 0
	// build the 2 parts of the query (column names and value names) as well as the values themselves
	for colname, val := range smap {
		q_cols = q_cols + deli + colname
		i++
		q_valnames = q_valnames + deli + fmt.Sprintf("$%d", i)
		q_vals = append(q_vals, val)
		deli = ","
	}
	rows, e = a.DB.QueryContext(ctx, queryname, "insert into "+a.SQLTablename+" ("+q_cols+") values ("+q_valnames+") returning id", q_vals...)
	if e != nil {
		return 0, a.Error(ctx, queryname, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, queryname, errors.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, queryname, errors.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// if ID==0 save, otherwise update
func (a *DBGitCredentials) SaveOrUpdate(ctx context.Context, p *savepb.GitCredentials) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBGitCredentials) Update(ctx context.Context, p *savepb.GitCredentials) error {
	qn := "DBGitCredentials_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, host=$2, path=$3, username=$4, password=$5, expiry=$6 where id = $7", a.get_UserID(p), a.get_Host(p), a.get_Path(p), a.get_Username(p), a.get_Password(p), a.get_Expiry(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGitCredentials) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGitCredentials_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGitCredentials) ByID(ctx context.Context, p uint64) (*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No GitCredentials with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GitCredentials with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGitCredentials) TryByID(ctx context.Context, p uint64) (*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GitCredentials with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBGitCredentials) ByIDs(ctx context.Context, p []uint64) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBGitCredentials) All(ctx context.Context) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBGitCredentials" rows with matching UserID
func (a *DBGitCredentials) ByUserID(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching UserID
func (a *DBGitCredentials) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeUserID(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Host
func (a *DBGitCredentials) ByHost(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByHost"
	l, e := a.fromQuery(ctx, qn, "host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching Host
func (a *DBGitCredentials) ByMultiHost(ctx context.Context, p []string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByHost"
	l, e := a.fromQuery(ctx, qn, "host in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeHost(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeHost"
	l, e := a.fromQuery(ctx, qn, "host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Path
func (a *DBGitCredentials) ByPath(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPath"
	l, e := a.fromQuery(ctx, qn, "path = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching Path
func (a *DBGitCredentials) ByMultiPath(ctx context.Context, p []string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPath"
	l, e := a.fromQuery(ctx, qn, "path in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikePath(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikePath"
	l, e := a.fromQuery(ctx, qn, "path ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Username
func (a *DBGitCredentials) ByUsername(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUsername"
	l, e := a.fromQuery(ctx, qn, "username = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUsername: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching Username
func (a *DBGitCredentials) ByMultiUsername(ctx context.Context, p []string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUsername"
	l, e := a.fromQuery(ctx, qn, "username in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUsername: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeUsername(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeUsername"
	l, e := a.fromQuery(ctx, qn, "username ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUsername: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Password
func (a *DBGitCredentials) ByPassword(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPassword"
	l, e := a.fromQuery(ctx, qn, "password = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching Password
func (a *DBGitCredentials) ByMultiPassword(ctx context.Context, p []string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPassword"
	l, e := a.fromQuery(ctx, qn, "password in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikePassword(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikePassword"
	l, e := a.fromQuery(ctx, qn, "password ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Expiry
func (a *DBGitCredentials) ByExpiry(ctx context.Context, p uint32) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByExpiry"
	l, e := a.fromQuery(ctx, qn, "expiry = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with multiple matching Expiry
func (a *DBGitCredentials) ByMultiExpiry(ctx context.Context, p []uint32) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByExpiry"
	l, e := a.fromQuery(ctx, qn, "expiry in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeExpiry"
	l, e := a.fromQuery(ctx, qn, "expiry ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBGitCredentials) get_ID(p *savepb.GitCredentials) uint64 {
	return uint64(p.ID)
}

// getter for field "UserID" (UserID) [string]
func (a *DBGitCredentials) get_UserID(p *savepb.GitCredentials) string {
	return string(p.UserID)
}

// getter for field "Host" (Host) [string]
func (a *DBGitCredentials) get_Host(p *savepb.GitCredentials) string {
	return string(p.Host)
}

// getter for field "Path" (Path) [string]
func (a *DBGitCredentials) get_Path(p *savepb.GitCredentials) string {
	return string(p.Path)
}

// getter for field "Username" (Username) [string]
func (a *DBGitCredentials) get_Username(p *savepb.GitCredentials) string {
	return string(p.Username)
}

// getter for field "Password" (Password) [string]
func (a *DBGitCredentials) get_Password(p *savepb.GitCredentials) string {
	return string(p.Password)
}

// getter for field "Expiry" (Expiry) [uint32]
func (a *DBGitCredentials) get_Expiry(p *savepb.GitCredentials) uint32 {
	return uint32(p.Expiry)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGitCredentials) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.GitCredentials, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		/*
		   efname:=fmt.Sprintf("EXTRA_FIELD_%d",i)
		   query.Add(col_name+" = "+efname,QP{efname:value})
		*/
		query.AddEqual(col_name, value)
	}

	gw, paras := query.ToPostgres()
	queryname := "custom_dbquery"
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where "+gw, paras...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (a *DBGitCredentials) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GitCredentials, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBGitCredentials) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.GitCredentials, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	eq := ""
	if extra_fields != nil && len(extra_fields) > 0 {
		eq = " AND ("
		// build the extraquery "eq"
		i := len(args)
		deli := ""
		for col_name, value := range extra_fields {
			i++
			eq = eq + deli + col_name + fmt.Sprintf(" = $%d", i)
			deli = " AND "
			args = append(args, value)
		}
		eq = eq + ")"
	}
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where ( "+query_where+") "+eq, args...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBGitCredentials) get_col_from_proto(p *savepb.GitCredentials, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "host" {
		return a.get_Host(p)
	} else if colname == "path" {
		return a.get_Path(p)
	} else if colname == "username" {
		return a.get_Username(p)
	} else if colname == "password" {
		return a.get_Password(p)
	} else if colname == "expiry" {
		return a.get_Expiry(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBGitCredentials) Tablename() string {
	return a.SQLTablename
}

func (a *DBGitCredentials) SelectCols() string {
	return "id,userid, host, path, username, password, expiry"
}
func (a *DBGitCredentials) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".host, " + a.SQLTablename + ".path, " + a.SQLTablename + ".username, " + a.SQLTablename + ".password, " + a.SQLTablename + ".expiry"
}

func (a *DBGitCredentials) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GitCredentials, error) {
	var res []*savepb.GitCredentials
	for rows.Next() {
		// SCANNER:
		foo := &savepb.GitCredentials{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.UserID
		scanTarget_2 := &foo.Host
		scanTarget_3 := &foo.Path
		scanTarget_4 := &foo.Username
		scanTarget_5 := &foo.Password
		scanTarget_6 := &foo.Expiry
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6)
		// END SCANNER

		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBGitCredentials) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,host text not null ,path text not null ,username text not null ,password text not null ,expiry integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,host text not null ,path text not null ,username text not null ,password text not null ,expiry integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS path text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS username text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS password text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS expiry integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS host text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS path text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS username text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS password text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS expiry integer not null  default 0;`,
	}

	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}

	// these are optional, expected to fail
	csql = []string{
		// Indices:

		// Foreign keys:

	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBGitCredentials) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

