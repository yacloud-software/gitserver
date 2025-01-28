package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBNRepoTagID
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence nrepotagid_seq;

Main Table:

 CREATE TABLE nrepotagid (id integer primary key default nextval('nrepotagid_seq'),repositoryid bigint not null  ,tag bigint not null  ,buildid bigint not null  );

Alter statements:
ALTER TABLE nrepotagid ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE nrepotagid ADD COLUMN IF NOT EXISTS tag bigint not null default 0;
ALTER TABLE nrepotagid ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE nrepotagid_archive (id integer unique not null,repositoryid bigint not null,tag bigint not null,buildid bigint not null);
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
	default_def_DBNRepoTagID *DBNRepoTagID
)

type DBNRepoTagID struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBNRepoTagID() *DBNRepoTagID {
	if default_def_DBNRepoTagID != nil {
		return default_def_DBNRepoTagID
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBNRepoTagID(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBNRepoTagID = res
	return res
}
func NewDBNRepoTagID(db *sql.DB) *DBNRepoTagID {
	foo := DBNRepoTagID{DB: db}
	foo.SQLTablename = "nrepotagid"
	foo.SQLArchivetablename = "nrepotagid_archive"
	return &foo
}

func (a *DBNRepoTagID) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBNRepoTagID) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBNRepoTagID) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBNRepoTagID) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBNRepoTagID", "insert into "+a.SQLArchivetablename+" (id,repositoryid, tag, buildid) values ($1,$2, $3, $4) ", p.ID, p.RepositoryID, p.Tag, p.BuildID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBNRepoTagID) buildSaveMap(ctx context.Context, p *savepb.NRepoTagID) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["tag"] = a.get_col_from_proto(p, "tag")
	res["buildid"] = a.get_col_from_proto(p, "buildid")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBNRepoTagID) Save(ctx context.Context, p *savepb.NRepoTagID) (uint64, error) {
	qn := "save_DBNRepoTagID"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBNRepoTagID) SaveWithID(ctx context.Context, p *savepb.NRepoTagID) error {
	qn := "insert_DBNRepoTagID"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBNRepoTagID) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.NRepoTagID) (uint64, error) {
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

func (a *DBNRepoTagID) Update(ctx context.Context, p *savepb.NRepoTagID) error {
	qn := "DBNRepoTagID_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, tag=$2, buildid=$3 where id = $4", a.get_RepositoryID(p), a.get_Tag(p), a.get_BuildID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBNRepoTagID) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBNRepoTagID_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBNRepoTagID) ByID(ctx context.Context, p uint64) (*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No NRepoTagID with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) NRepoTagID with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBNRepoTagID) TryByID(ctx context.Context, p uint64) (*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) NRepoTagID with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBNRepoTagID) ByIDs(ctx context.Context, p []uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBNRepoTagID) All(ctx context.Context) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBNRepoTagID" rows with matching RepositoryID
func (a *DBNRepoTagID) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBNRepoTagID" rows with multiple matching RepositoryID
func (a *DBNRepoTagID) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBNRepoTagID) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBNRepoTagID" rows with matching Tag
func (a *DBNRepoTagID) ByTag(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByTag"
	l, e := a.fromQuery(ctx, qn, "tag = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTag: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBNRepoTagID" rows with multiple matching Tag
func (a *DBNRepoTagID) ByMultiTag(ctx context.Context, p []uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByTag"
	l, e := a.fromQuery(ctx, qn, "tag in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTag: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBNRepoTagID) ByLikeTag(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByLikeTag"
	l, e := a.fromQuery(ctx, qn, "tag ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTag: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBNRepoTagID" rows with matching BuildID
func (a *DBNRepoTagID) ByBuildID(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBNRepoTagID" rows with multiple matching BuildID
func (a *DBNRepoTagID) ByMultiBuildID(ctx context.Context, p []uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBNRepoTagID) ByLikeBuildID(ctx context.Context, p uint64) ([]*savepb.NRepoTagID, error) {
	qn := "DBNRepoTagID_ByLikeBuildID"
	l, e := a.fromQuery(ctx, qn, "buildid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBNRepoTagID) get_ID(p *savepb.NRepoTagID) uint64 {
	return uint64(p.ID)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBNRepoTagID) get_RepositoryID(p *savepb.NRepoTagID) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "Tag" (Tag) [uint64]
func (a *DBNRepoTagID) get_Tag(p *savepb.NRepoTagID) uint64 {
	return uint64(p.Tag)
}

// getter for field "BuildID" (BuildID) [uint64]
func (a *DBNRepoTagID) get_BuildID(p *savepb.NRepoTagID) uint64 {
	return uint64(p.BuildID)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBNRepoTagID) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.NRepoTagID, error) {
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

func (a *DBNRepoTagID) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.NRepoTagID, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBNRepoTagID) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.NRepoTagID, error) {
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
func (a *DBNRepoTagID) get_col_from_proto(p *savepb.NRepoTagID, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "tag" {
		return a.get_Tag(p)
	} else if colname == "buildid" {
		return a.get_BuildID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBNRepoTagID) Tablename() string {
	return a.SQLTablename
}

func (a *DBNRepoTagID) SelectCols() string {
	return "id,repositoryid, tag, buildid"
}
func (a *DBNRepoTagID) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".tag, " + a.SQLTablename + ".buildid"
}

func (a *DBNRepoTagID) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.NRepoTagID, error) {
	var res []*savepb.NRepoTagID
	for rows.Next() {
		// SCANNER:
		foo := &savepb.NRepoTagID{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepositoryID
		scanTarget_2 := &foo.Tag
		scanTarget_3 := &foo.BuildID
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
func (a *DBNRepoTagID) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,tag bigint not null ,buildid bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,tag bigint not null ,buildid bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS tag bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS buildid bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS tag bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS buildid bigint not null  default 0;`,
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
func (a *DBNRepoTagID) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

