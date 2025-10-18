package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBUserRepositoryAccess
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence userrepositoryaccess_seq;

Main Table:

 CREATE TABLE userrepositoryaccess (id integer primary key default nextval('userrepositoryaccess_seq'),repoid bigint not null  ,userid text not null  ,read boolean not null  ,write boolean not null  );

Alter statements:
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS read boolean not null default false;
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS write boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE userrepositoryaccess_archive (id integer unique not null,repoid bigint not null,userid text not null,read boolean not null,write boolean not null);
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
	default_def_DBUserRepositoryAccess *DBUserRepositoryAccess
)

type DBUserRepositoryAccess struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBUserRepositoryAccess()
	})
}

func DefaultDBUserRepositoryAccess() *DBUserRepositoryAccess {
	if default_def_DBUserRepositoryAccess != nil {
		return default_def_DBUserRepositoryAccess
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBUserRepositoryAccess(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBUserRepositoryAccess = res
	return res
}
func NewDBUserRepositoryAccess(db *sql.DB) *DBUserRepositoryAccess {
	foo := DBUserRepositoryAccess{DB: db}
	foo.SQLTablename = "userrepositoryaccess"
	foo.SQLArchivetablename = "userrepositoryaccess_archive"
	return &foo
}

func (a *DBUserRepositoryAccess) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBUserRepositoryAccess) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBUserRepositoryAccess) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBUserRepositoryAccess) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBUserRepositoryAccess", "insert into "+a.SQLArchivetablename+" (id,repoid, userid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.UserID, p.Read, p.Write)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBUserRepositoryAccess) buildSaveMap(ctx context.Context, p *savepb.UserRepositoryAccess) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repoid"] = a.get_col_from_proto(p, "repoid")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["read"] = a.get_col_from_proto(p, "read")
	res["write"] = a.get_col_from_proto(p, "write")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBUserRepositoryAccess) Save(ctx context.Context, p *savepb.UserRepositoryAccess) (uint64, error) {
	qn := "save_DBUserRepositoryAccess"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBUserRepositoryAccess) SaveWithID(ctx context.Context, p *savepb.UserRepositoryAccess) error {
	qn := "insert_DBUserRepositoryAccess"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBUserRepositoryAccess) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.UserRepositoryAccess) (uint64, error) {
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
func (a *DBUserRepositoryAccess) SaveOrUpdate(ctx context.Context, p *savepb.UserRepositoryAccess) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBUserRepositoryAccess) Update(ctx context.Context, p *savepb.UserRepositoryAccess) error {
	qn := "DBUserRepositoryAccess_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repoid=$1, userid=$2, read=$3, write=$4 where id = $5", a.get_RepoID(p), a.get_UserID(p), a.get_Read(p), a.get_Write(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBUserRepositoryAccess) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBUserRepositoryAccess_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBUserRepositoryAccess) ByID(ctx context.Context, p uint64) (*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No UserRepositoryAccess with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) UserRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBUserRepositoryAccess) TryByID(ctx context.Context, p uint64) (*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) UserRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBUserRepositoryAccess) ByIDs(ctx context.Context, p []uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBUserRepositoryAccess) All(ctx context.Context) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBUserRepositoryAccess" rows with matching RepoID
func (a *DBUserRepositoryAccess) ByRepoID(ctx context.Context, p uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with multiple matching RepoID
func (a *DBUserRepositoryAccess) ByMultiRepoID(ctx context.Context, p []uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeRepoID(ctx context.Context, p uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with matching UserID
func (a *DBUserRepositoryAccess) ByUserID(ctx context.Context, p string) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with multiple matching UserID
func (a *DBUserRepositoryAccess) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeUserID(ctx context.Context, p string) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with matching Read
func (a *DBUserRepositoryAccess) ByRead(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRead"
	l, e := a.fromQuery(ctx, qn, "read = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with multiple matching Read
func (a *DBUserRepositoryAccess) ByMultiRead(ctx context.Context, p []bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRead"
	l, e := a.fromQuery(ctx, qn, "read in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeRead(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeRead"
	l, e := a.fromQuery(ctx, qn, "read ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with matching Write
func (a *DBUserRepositoryAccess) ByWrite(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with multiple matching Write
func (a *DBUserRepositoryAccess) ByMultiWrite(ctx context.Context, p []bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeWrite"
	l, e := a.fromQuery(ctx, qn, "write ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBUserRepositoryAccess) get_ID(p *savepb.UserRepositoryAccess) uint64 {
	return uint64(p.ID)
}

// getter for field "RepoID" (RepoID) [uint64]
func (a *DBUserRepositoryAccess) get_RepoID(p *savepb.UserRepositoryAccess) uint64 {
	return uint64(p.RepoID)
}

// getter for field "UserID" (UserID) [string]
func (a *DBUserRepositoryAccess) get_UserID(p *savepb.UserRepositoryAccess) string {
	return string(p.UserID)
}

// getter for field "Read" (Read) [bool]
func (a *DBUserRepositoryAccess) get_Read(p *savepb.UserRepositoryAccess) bool {
	return bool(p.Read)
}

// getter for field "Write" (Write) [bool]
func (a *DBUserRepositoryAccess) get_Write(p *savepb.UserRepositoryAccess) bool {
	return bool(p.Write)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBUserRepositoryAccess) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.UserRepositoryAccess, error) {
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

func (a *DBUserRepositoryAccess) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.UserRepositoryAccess, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBUserRepositoryAccess) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.UserRepositoryAccess, error) {
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
func (a *DBUserRepositoryAccess) get_col_from_proto(p *savepb.UserRepositoryAccess, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repoid" {
		return a.get_RepoID(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "read" {
		return a.get_Read(p)
	} else if colname == "write" {
		return a.get_Write(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBUserRepositoryAccess) Tablename() string {
	return a.SQLTablename
}

func (a *DBUserRepositoryAccess) SelectCols() string {
	return "id,repoid, userid, read, write"
}
func (a *DBUserRepositoryAccess) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repoid, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".read, " + a.SQLTablename + ".write"
}

func (a *DBUserRepositoryAccess) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.UserRepositoryAccess, error) {
	var res []*savepb.UserRepositoryAccess
	for rows.Next() {
		// SCANNER:
		foo := &savepb.UserRepositoryAccess{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepoID
		scanTarget_2 := &foo.UserID
		scanTarget_3 := &foo.Read
		scanTarget_4 := &foo.Write
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4)
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
func (a *DBUserRepositoryAccess) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,userid text not null ,read boolean not null ,write boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,userid text not null ,read boolean not null ,write boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS read boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS write boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repoid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS read boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS write boolean not null  default false;`,
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
func (a *DBUserRepositoryAccess) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

