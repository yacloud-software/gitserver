package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBWatchers
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence watchers_seq;

Main Table:

 CREATE TABLE watchers (id integer primary key default nextval('watchers_seq'),userid text not null  ,repositoryid bigint not null  ,notifytype integer not null  );

Alter statements:
ALTER TABLE watchers ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE watchers ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE watchers ADD COLUMN IF NOT EXISTS notifytype integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE watchers_archive (id integer unique not null,userid text not null,repositoryid bigint not null,notifytype integer not null);
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
	default_def_DBWatchers *DBWatchers
)

type DBWatchers struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBWatchers()
	})
}

func DefaultDBWatchers() *DBWatchers {
	if default_def_DBWatchers != nil {
		return default_def_DBWatchers
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBWatchers(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBWatchers = res
	return res
}
func NewDBWatchers(db *sql.DB) *DBWatchers {
	foo := DBWatchers{DB: db}
	foo.SQLTablename = "watchers"
	foo.SQLArchivetablename = "watchers_archive"
	return &foo
}

func (a *DBWatchers) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBWatchers) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBWatchers) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBWatchers) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBWatchers", "insert into "+a.SQLArchivetablename+" (id,userid, repositoryid, notifytype) values ($1,$2, $3, $4) ", p.ID, p.UserID, p.RepositoryID, p.Notifytype)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBWatchers) buildSaveMap(ctx context.Context, p *savepb.Watchers) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["notifytype"] = a.get_col_from_proto(p, "notifytype")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBWatchers) Save(ctx context.Context, p *savepb.Watchers) (uint64, error) {
	qn := "save_DBWatchers"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBWatchers) SaveWithID(ctx context.Context, p *savepb.Watchers) error {
	qn := "insert_DBWatchers"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBWatchers) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.Watchers) (uint64, error) {
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
func (a *DBWatchers) SaveOrUpdate(ctx context.Context, p *savepb.Watchers) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBWatchers) Update(ctx context.Context, p *savepb.Watchers) error {
	qn := "DBWatchers_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, repositoryid=$2, notifytype=$3 where id = $4", a.get_UserID(p), a.get_RepositoryID(p), a.get_Notifytype(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBWatchers) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBWatchers_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBWatchers) ByID(ctx context.Context, p uint64) (*savepb.Watchers, error) {
	qn := "DBWatchers_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No Watchers with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Watchers with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBWatchers) TryByID(ctx context.Context, p uint64) (*savepb.Watchers, error) {
	qn := "DBWatchers_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Watchers with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBWatchers) ByIDs(ctx context.Context, p []uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBWatchers) All(ctx context.Context) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBWatchers" rows with matching UserID
func (a *DBWatchers) ByUserID(ctx context.Context, p string) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBWatchers" rows with multiple matching UserID
func (a *DBWatchers) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBWatchers) ByLikeUserID(ctx context.Context, p string) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBWatchers" rows with matching RepositoryID
func (a *DBWatchers) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBWatchers" rows with multiple matching RepositoryID
func (a *DBWatchers) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBWatchers) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBWatchers" rows with matching Notifytype
func (a *DBWatchers) ByNotifytype(ctx context.Context, p uint32) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByNotifytype"
	l, e := a.fromQuery(ctx, qn, "notifytype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotifytype: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBWatchers" rows with multiple matching Notifytype
func (a *DBWatchers) ByMultiNotifytype(ctx context.Context, p []uint32) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByNotifytype"
	l, e := a.fromQuery(ctx, qn, "notifytype in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotifytype: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBWatchers) ByLikeNotifytype(ctx context.Context, p uint32) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeNotifytype"
	l, e := a.fromQuery(ctx, qn, "notifytype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByNotifytype: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBWatchers) get_ID(p *savepb.Watchers) uint64 {
	return uint64(p.ID)
}

// getter for field "UserID" (UserID) [string]
func (a *DBWatchers) get_UserID(p *savepb.Watchers) string {
	return string(p.UserID)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBWatchers) get_RepositoryID(p *savepb.Watchers) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "Notifytype" (Notifytype) [uint32]
func (a *DBWatchers) get_Notifytype(p *savepb.Watchers) uint32 {
	return uint32(p.Notifytype)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBWatchers) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.Watchers, error) {
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

func (a *DBWatchers) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Watchers, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBWatchers) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.Watchers, error) {
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
func (a *DBWatchers) get_col_from_proto(p *savepb.Watchers, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "notifytype" {
		return a.get_Notifytype(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBWatchers) Tablename() string {
	return a.SQLTablename
}

func (a *DBWatchers) SelectCols() string {
	return "id,userid, repositoryid, notifytype"
}
func (a *DBWatchers) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".notifytype"
}

func (a *DBWatchers) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Watchers, error) {
	var res []*savepb.Watchers
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Watchers{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.UserID
		scanTarget_2 := &foo.RepositoryID
		scanTarget_3 := &foo.Notifytype
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3)
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
func (a *DBWatchers) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,repositoryid bigint not null ,notifytype integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,repositoryid bigint not null ,notifytype integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS notifytype integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS notifytype integer not null  default 0;`,
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
func (a *DBWatchers) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

