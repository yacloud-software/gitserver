package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBInternalGitHost
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence internalgithost_seq;

Main Table:

 CREATE TABLE internalgithost (id integer primary key default nextval('internalgithost_seq'),host text not null  ,expiry integer not null  );

Alter statements:
ALTER TABLE internalgithost ADD COLUMN IF NOT EXISTS host text not null default '';
ALTER TABLE internalgithost ADD COLUMN IF NOT EXISTS expiry integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE internalgithost_archive (id integer unique not null,host text not null,expiry integer not null);
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
	default_def_DBInternalGitHost *DBInternalGitHost
)

type DBInternalGitHost struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBInternalGitHost()
	})
}

func DefaultDBInternalGitHost() *DBInternalGitHost {
	if default_def_DBInternalGitHost != nil {
		return default_def_DBInternalGitHost
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBInternalGitHost(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBInternalGitHost = res
	return res
}
func NewDBInternalGitHost(db *sql.DB) *DBInternalGitHost {
	foo := DBInternalGitHost{DB: db}
	foo.SQLTablename = "internalgithost"
	foo.SQLArchivetablename = "internalgithost_archive"
	return &foo
}

func (a *DBInternalGitHost) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBInternalGitHost) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBInternalGitHost) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBInternalGitHost) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBInternalGitHost", "insert into "+a.SQLArchivetablename+" (id,host, expiry) values ($1,$2, $3) ", p.ID, p.Host, p.Expiry)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBInternalGitHost) buildSaveMap(ctx context.Context, p *savepb.InternalGitHost) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["host"] = a.get_col_from_proto(p, "host")
	res["expiry"] = a.get_col_from_proto(p, "expiry")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBInternalGitHost) Save(ctx context.Context, p *savepb.InternalGitHost) (uint64, error) {
	qn := "save_DBInternalGitHost"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBInternalGitHost) SaveWithID(ctx context.Context, p *savepb.InternalGitHost) error {
	qn := "insert_DBInternalGitHost"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBInternalGitHost) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.InternalGitHost) (uint64, error) {
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
func (a *DBInternalGitHost) SaveOrUpdate(ctx context.Context, p *savepb.InternalGitHost) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBInternalGitHost) Update(ctx context.Context, p *savepb.InternalGitHost) error {
	qn := "DBInternalGitHost_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set host=$1, expiry=$2 where id = $3", a.get_Host(p), a.get_Expiry(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBInternalGitHost) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBInternalGitHost_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBInternalGitHost) ByID(ctx context.Context, p uint64) (*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No InternalGitHost with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) InternalGitHost with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBInternalGitHost) TryByID(ctx context.Context, p uint64) (*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) InternalGitHost with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBInternalGitHost) ByIDs(ctx context.Context, p []uint64) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBInternalGitHost) All(ctx context.Context) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBInternalGitHost" rows with matching Host
func (a *DBInternalGitHost) ByHost(ctx context.Context, p string) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByHost"
	l, e := a.fromQuery(ctx, qn, "host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBInternalGitHost" rows with multiple matching Host
func (a *DBInternalGitHost) ByMultiHost(ctx context.Context, p []string) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByHost"
	l, e := a.fromQuery(ctx, qn, "host in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInternalGitHost) ByLikeHost(ctx context.Context, p string) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByLikeHost"
	l, e := a.fromQuery(ctx, qn, "host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBInternalGitHost" rows with matching Expiry
func (a *DBInternalGitHost) ByExpiry(ctx context.Context, p uint32) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByExpiry"
	l, e := a.fromQuery(ctx, qn, "expiry = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBInternalGitHost" rows with multiple matching Expiry
func (a *DBInternalGitHost) ByMultiExpiry(ctx context.Context, p []uint32) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByExpiry"
	l, e := a.fromQuery(ctx, qn, "expiry in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInternalGitHost) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByLikeExpiry"
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
func (a *DBInternalGitHost) get_ID(p *savepb.InternalGitHost) uint64 {
	return uint64(p.ID)
}

// getter for field "Host" (Host) [string]
func (a *DBInternalGitHost) get_Host(p *savepb.InternalGitHost) string {
	return string(p.Host)
}

// getter for field "Expiry" (Expiry) [uint32]
func (a *DBInternalGitHost) get_Expiry(p *savepb.InternalGitHost) uint32 {
	return uint32(p.Expiry)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBInternalGitHost) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.InternalGitHost, error) {
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

func (a *DBInternalGitHost) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.InternalGitHost, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBInternalGitHost) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.InternalGitHost, error) {
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
func (a *DBInternalGitHost) get_col_from_proto(p *savepb.InternalGitHost, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "host" {
		return a.get_Host(p)
	} else if colname == "expiry" {
		return a.get_Expiry(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBInternalGitHost) Tablename() string {
	return a.SQLTablename
}

func (a *DBInternalGitHost) SelectCols() string {
	return "id,host, expiry"
}
func (a *DBInternalGitHost) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".host, " + a.SQLTablename + ".expiry"
}

func (a *DBInternalGitHost) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.InternalGitHost, error) {
	var res []*savepb.InternalGitHost
	for rows.Next() {
		// SCANNER:
		foo := &savepb.InternalGitHost{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Host
		scanTarget_2 := &foo.Expiry
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2)
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
func (a *DBInternalGitHost) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),host text not null ,expiry integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),host text not null ,expiry integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS expiry integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS host text not null  default '';`,
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
func (a *DBInternalGitHost) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

