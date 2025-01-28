package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBPingState
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence pingstate_seq;

Main Table:

 CREATE TABLE pingstate (id integer primary key default nextval('pingstate_seq'),associationtoken text not null  ,created integer not null  ,responsetoken text not null  );

Alter statements:
ALTER TABLE pingstate ADD COLUMN IF NOT EXISTS associationtoken text not null default '';
ALTER TABLE pingstate ADD COLUMN IF NOT EXISTS created integer not null default 0;
ALTER TABLE pingstate ADD COLUMN IF NOT EXISTS responsetoken text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE pingstate_archive (id integer unique not null,associationtoken text not null,created integer not null,responsetoken text not null);
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
	default_def_DBPingState *DBPingState
)

type DBPingState struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBPingState() *DBPingState {
	if default_def_DBPingState != nil {
		return default_def_DBPingState
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBPingState(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBPingState = res
	return res
}
func NewDBPingState(db *sql.DB) *DBPingState {
	foo := DBPingState{DB: db}
	foo.SQLTablename = "pingstate"
	foo.SQLArchivetablename = "pingstate_archive"
	return &foo
}

func (a *DBPingState) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBPingState) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBPingState) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBPingState) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBPingState", "insert into "+a.SQLArchivetablename+" (id,associationtoken, created, responsetoken) values ($1,$2, $3, $4) ", p.ID, p.AssociationToken, p.Created, p.ResponseToken)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBPingState) buildSaveMap(ctx context.Context, p *savepb.PingState) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["associationtoken"] = a.get_col_from_proto(p, "associationtoken")
	res["created"] = a.get_col_from_proto(p, "created")
	res["responsetoken"] = a.get_col_from_proto(p, "responsetoken")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBPingState) Save(ctx context.Context, p *savepb.PingState) (uint64, error) {
	qn := "save_DBPingState"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBPingState) SaveWithID(ctx context.Context, p *savepb.PingState) error {
	qn := "insert_DBPingState"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBPingState) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.PingState) (uint64, error) {
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

func (a *DBPingState) Update(ctx context.Context, p *savepb.PingState) error {
	qn := "DBPingState_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set associationtoken=$1, created=$2, responsetoken=$3 where id = $4", a.get_AssociationToken(p), a.get_Created(p), a.get_ResponseToken(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBPingState) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBPingState_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBPingState) ByID(ctx context.Context, p uint64) (*savepb.PingState, error) {
	qn := "DBPingState_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No PingState with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) PingState with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBPingState) TryByID(ctx context.Context, p uint64) (*savepb.PingState, error) {
	qn := "DBPingState_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) PingState with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBPingState) ByIDs(ctx context.Context, p []uint64) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBPingState) All(ctx context.Context) ([]*savepb.PingState, error) {
	qn := "DBPingState_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBPingState" rows with matching AssociationToken
func (a *DBPingState) ByAssociationToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with multiple matching AssociationToken
func (a *DBPingState) ByMultiAssociationToken(ctx context.Context, p []string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeAssociationToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with matching Created
func (a *DBPingState) ByCreated(ctx context.Context, p uint32) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with multiple matching Created
func (a *DBPingState) ByMultiCreated(ctx context.Context, p []uint32) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByCreated"
	l, e := a.fromQuery(ctx, qn, "created in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeCreated"
	l, e := a.fromQuery(ctx, qn, "created ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with matching ResponseToken
func (a *DBPingState) ByResponseToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByResponseToken"
	l, e := a.fromQuery(ctx, qn, "responsetoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByResponseToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with multiple matching ResponseToken
func (a *DBPingState) ByMultiResponseToken(ctx context.Context, p []string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByResponseToken"
	l, e := a.fromQuery(ctx, qn, "responsetoken in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByResponseToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeResponseToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeResponseToken"
	l, e := a.fromQuery(ctx, qn, "responsetoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByResponseToken: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBPingState) get_ID(p *savepb.PingState) uint64 {
	return uint64(p.ID)
}

// getter for field "AssociationToken" (AssociationToken) [string]
func (a *DBPingState) get_AssociationToken(p *savepb.PingState) string {
	return string(p.AssociationToken)
}

// getter for field "Created" (Created) [uint32]
func (a *DBPingState) get_Created(p *savepb.PingState) uint32 {
	return uint32(p.Created)
}

// getter for field "ResponseToken" (ResponseToken) [string]
func (a *DBPingState) get_ResponseToken(p *savepb.PingState) string {
	return string(p.ResponseToken)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBPingState) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.PingState, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		efname := fmt.Sprintf("EXTRA_FIELD_%d", i)
		query.Add(col_name+" = "+efname, QP{efname: value})
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

func (a *DBPingState) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.PingState, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBPingState) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.PingState, error) {
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
func (a *DBPingState) get_col_from_proto(p *savepb.PingState, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "associationtoken" {
		return a.get_AssociationToken(p)
	} else if colname == "created" {
		return a.get_Created(p)
	} else if colname == "responsetoken" {
		return a.get_ResponseToken(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBPingState) Tablename() string {
	return a.SQLTablename
}

func (a *DBPingState) SelectCols() string {
	return "id,associationtoken, created, responsetoken"
}
func (a *DBPingState) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".associationtoken, " + a.SQLTablename + ".created, " + a.SQLTablename + ".responsetoken"
}

func (a *DBPingState) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.PingState, error) {
	var res []*savepb.PingState
	for rows.Next() {
		// SCANNER:
		foo := &savepb.PingState{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.AssociationToken
		scanTarget_2 := &foo.Created
		scanTarget_3 := &foo.ResponseToken
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
func (a *DBPingState) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),associationtoken text not null ,created integer not null ,responsetoken text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),associationtoken text not null ,created integer not null ,responsetoken text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS associationtoken text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS created integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS responsetoken text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS associationtoken text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS created integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS responsetoken text not null  default '';`,
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
func (a *DBPingState) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

