package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBSourceRepositoryURL
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence sourcerepositoryurl_seq;

Main Table:

 CREATE TABLE sourcerepositoryurl (id integer primary key default nextval('sourcerepositoryurl_seq'),v2repositoryid bigint not null  ,host text not null  ,path text not null  );

Alter statements:
ALTER TABLE sourcerepositoryurl ADD COLUMN IF NOT EXISTS v2repositoryid bigint not null default 0;
ALTER TABLE sourcerepositoryurl ADD COLUMN IF NOT EXISTS host text not null default '';
ALTER TABLE sourcerepositoryurl ADD COLUMN IF NOT EXISTS path text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE sourcerepositoryurl_archive (id integer unique not null,v2repositoryid bigint not null,host text not null,path text not null);
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
	default_def_DBSourceRepositoryURL *DBSourceRepositoryURL
)

type DBSourceRepositoryURL struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBSourceRepositoryURL()
	})
}

func DefaultDBSourceRepositoryURL() *DBSourceRepositoryURL {
	if default_def_DBSourceRepositoryURL != nil {
		return default_def_DBSourceRepositoryURL
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBSourceRepositoryURL(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBSourceRepositoryURL = res
	return res
}
func NewDBSourceRepositoryURL(db *sql.DB) *DBSourceRepositoryURL {
	foo := DBSourceRepositoryURL{DB: db}
	foo.SQLTablename = "sourcerepositoryurl"
	foo.SQLArchivetablename = "sourcerepositoryurl_archive"
	return &foo
}

func (a *DBSourceRepositoryURL) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBSourceRepositoryURL) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBSourceRepositoryURL) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBSourceRepositoryURL) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSourceRepositoryURL", "insert into "+a.SQLArchivetablename+" (id,v2repositoryid, host, path) values ($1,$2, $3, $4) ", p.ID, p.V2RepositoryID, p.Host, p.Path)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBSourceRepositoryURL) buildSaveMap(ctx context.Context, p *savepb.SourceRepositoryURL) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["v2repositoryid"] = a.get_col_from_proto(p, "v2repositoryid")
	res["host"] = a.get_col_from_proto(p, "host")
	res["path"] = a.get_col_from_proto(p, "path")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBSourceRepositoryURL) Save(ctx context.Context, p *savepb.SourceRepositoryURL) (uint64, error) {
	qn := "save_DBSourceRepositoryURL"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBSourceRepositoryURL) SaveWithID(ctx context.Context, p *savepb.SourceRepositoryURL) error {
	qn := "insert_DBSourceRepositoryURL"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBSourceRepositoryURL) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.SourceRepositoryURL) (uint64, error) {
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
func (a *DBSourceRepositoryURL) SaveOrUpdate(ctx context.Context, p *savepb.SourceRepositoryURL) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBSourceRepositoryURL) Update(ctx context.Context, p *savepb.SourceRepositoryURL) error {
	qn := "DBSourceRepositoryURL_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set v2repositoryid=$1, host=$2, path=$3 where id = $4", a.get_V2RepositoryID(p), a.get_Host(p), a.get_Path(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBSourceRepositoryURL) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBSourceRepositoryURL_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBSourceRepositoryURL) ByID(ctx context.Context, p uint64) (*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No SourceRepositoryURL with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SourceRepositoryURL with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBSourceRepositoryURL) TryByID(ctx context.Context, p uint64) (*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SourceRepositoryURL with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBSourceRepositoryURL) ByIDs(ctx context.Context, p []uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBSourceRepositoryURL) All(ctx context.Context) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBSourceRepositoryURL" rows with matching V2RepositoryID
func (a *DBSourceRepositoryURL) ByV2RepositoryID(ctx context.Context, p uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByV2RepositoryID"
	l, e := a.fromQuery(ctx, qn, "v2repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByV2RepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with multiple matching V2RepositoryID
func (a *DBSourceRepositoryURL) ByMultiV2RepositoryID(ctx context.Context, p []uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByV2RepositoryID"
	l, e := a.fromQuery(ctx, qn, "v2repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByV2RepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepositoryURL) ByLikeV2RepositoryID(ctx context.Context, p uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikeV2RepositoryID"
	l, e := a.fromQuery(ctx, qn, "v2repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByV2RepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with matching Host
func (a *DBSourceRepositoryURL) ByHost(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByHost"
	l, e := a.fromQuery(ctx, qn, "host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with multiple matching Host
func (a *DBSourceRepositoryURL) ByMultiHost(ctx context.Context, p []string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByHost"
	l, e := a.fromQuery(ctx, qn, "host in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepositoryURL) ByLikeHost(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikeHost"
	l, e := a.fromQuery(ctx, qn, "host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with matching Path
func (a *DBSourceRepositoryURL) ByPath(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByPath"
	l, e := a.fromQuery(ctx, qn, "path = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with multiple matching Path
func (a *DBSourceRepositoryURL) ByMultiPath(ctx context.Context, p []string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByPath"
	l, e := a.fromQuery(ctx, qn, "path in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepositoryURL) ByLikePath(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikePath"
	l, e := a.fromQuery(ctx, qn, "path ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBSourceRepositoryURL) get_ID(p *savepb.SourceRepositoryURL) uint64 {
	return uint64(p.ID)
}

// getter for field "V2RepositoryID" (V2RepositoryID) [uint64]
func (a *DBSourceRepositoryURL) get_V2RepositoryID(p *savepb.SourceRepositoryURL) uint64 {
	return uint64(p.V2RepositoryID)
}

// getter for field "Host" (Host) [string]
func (a *DBSourceRepositoryURL) get_Host(p *savepb.SourceRepositoryURL) string {
	return string(p.Host)
}

// getter for field "Path" (Path) [string]
func (a *DBSourceRepositoryURL) get_Path(p *savepb.SourceRepositoryURL) string {
	return string(p.Path)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSourceRepositoryURL) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.SourceRepositoryURL, error) {
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

func (a *DBSourceRepositoryURL) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SourceRepositoryURL, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBSourceRepositoryURL) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.SourceRepositoryURL, error) {
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
func (a *DBSourceRepositoryURL) get_col_from_proto(p *savepb.SourceRepositoryURL, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "v2repositoryid" {
		return a.get_V2RepositoryID(p)
	} else if colname == "host" {
		return a.get_Host(p)
	} else if colname == "path" {
		return a.get_Path(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBSourceRepositoryURL) Tablename() string {
	return a.SQLTablename
}

func (a *DBSourceRepositoryURL) SelectCols() string {
	return "id,v2repositoryid, host, path"
}
func (a *DBSourceRepositoryURL) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".v2repositoryid, " + a.SQLTablename + ".host, " + a.SQLTablename + ".path"
}

func (a *DBSourceRepositoryURL) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SourceRepositoryURL, error) {
	var res []*savepb.SourceRepositoryURL
	for rows.Next() {
		// SCANNER:
		foo := &savepb.SourceRepositoryURL{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.V2RepositoryID
		scanTarget_2 := &foo.Host
		scanTarget_3 := &foo.Path
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
func (a *DBSourceRepositoryURL) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),v2repositoryid bigint not null ,host text not null ,path text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),v2repositoryid bigint not null ,host text not null ,path text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS v2repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS path text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS v2repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS host text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS path text not null  default '';`,
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
func (a *DBSourceRepositoryURL) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

