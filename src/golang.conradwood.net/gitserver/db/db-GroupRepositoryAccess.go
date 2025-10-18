package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBGroupRepositoryAccess
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence grouprepositoryaccess_seq;

Main Table:

 CREATE TABLE grouprepositoryaccess (id integer primary key default nextval('grouprepositoryaccess_seq'),repoid bigint not null  ,groupid text not null  ,read boolean not null  ,write boolean not null  );

Alter statements:
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS groupid text not null default '';
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS read boolean not null default false;
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS write boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE grouprepositoryaccess_archive (id integer unique not null,repoid bigint not null,groupid text not null,read boolean not null,write boolean not null);
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
	default_def_DBGroupRepositoryAccess *DBGroupRepositoryAccess
)

type DBGroupRepositoryAccess struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBGroupRepositoryAccess()
	})
}

func DefaultDBGroupRepositoryAccess() *DBGroupRepositoryAccess {
	if default_def_DBGroupRepositoryAccess != nil {
		return default_def_DBGroupRepositoryAccess
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGroupRepositoryAccess(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGroupRepositoryAccess = res
	return res
}
func NewDBGroupRepositoryAccess(db *sql.DB) *DBGroupRepositoryAccess {
	foo := DBGroupRepositoryAccess{DB: db}
	foo.SQLTablename = "grouprepositoryaccess"
	foo.SQLArchivetablename = "grouprepositoryaccess_archive"
	return &foo
}

func (a *DBGroupRepositoryAccess) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBGroupRepositoryAccess) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBGroupRepositoryAccess) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBGroupRepositoryAccess) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGroupRepositoryAccess", "insert into "+a.SQLArchivetablename+" (id,repoid, groupid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.GroupID, p.Read, p.Write)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBGroupRepositoryAccess) buildSaveMap(ctx context.Context, p *savepb.GroupRepositoryAccess) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repoid"] = a.get_col_from_proto(p, "repoid")
	res["groupid"] = a.get_col_from_proto(p, "groupid")
	res["read"] = a.get_col_from_proto(p, "read")
	res["write"] = a.get_col_from_proto(p, "write")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBGroupRepositoryAccess) Save(ctx context.Context, p *savepb.GroupRepositoryAccess) (uint64, error) {
	qn := "save_DBGroupRepositoryAccess"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBGroupRepositoryAccess) SaveWithID(ctx context.Context, p *savepb.GroupRepositoryAccess) error {
	qn := "insert_DBGroupRepositoryAccess"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBGroupRepositoryAccess) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.GroupRepositoryAccess) (uint64, error) {
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
func (a *DBGroupRepositoryAccess) SaveOrUpdate(ctx context.Context, p *savepb.GroupRepositoryAccess) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBGroupRepositoryAccess) Update(ctx context.Context, p *savepb.GroupRepositoryAccess) error {
	qn := "DBGroupRepositoryAccess_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repoid=$1, groupid=$2, read=$3, write=$4 where id = $5", a.get_RepoID(p), a.get_GroupID(p), a.get_Read(p), a.get_Write(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGroupRepositoryAccess) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGroupRepositoryAccess_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGroupRepositoryAccess) ByID(ctx context.Context, p uint64) (*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No GroupRepositoryAccess with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GroupRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGroupRepositoryAccess) TryByID(ctx context.Context, p uint64) (*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GroupRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBGroupRepositoryAccess) ByIDs(ctx context.Context, p []uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBGroupRepositoryAccess) All(ctx context.Context) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBGroupRepositoryAccess" rows with matching RepoID
func (a *DBGroupRepositoryAccess) ByRepoID(ctx context.Context, p uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with multiple matching RepoID
func (a *DBGroupRepositoryAccess) ByMultiRepoID(ctx context.Context, p []uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupRepositoryAccess) ByLikeRepoID(ctx context.Context, p uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeRepoID"
	l, e := a.fromQuery(ctx, qn, "repoid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with matching GroupID
func (a *DBGroupRepositoryAccess) ByGroupID(ctx context.Context, p string) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByGroupID"
	l, e := a.fromQuery(ctx, qn, "groupid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with multiple matching GroupID
func (a *DBGroupRepositoryAccess) ByMultiGroupID(ctx context.Context, p []string) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByGroupID"
	l, e := a.fromQuery(ctx, qn, "groupid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupRepositoryAccess) ByLikeGroupID(ctx context.Context, p string) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeGroupID"
	l, e := a.fromQuery(ctx, qn, "groupid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with matching Read
func (a *DBGroupRepositoryAccess) ByRead(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRead"
	l, e := a.fromQuery(ctx, qn, "read = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with multiple matching Read
func (a *DBGroupRepositoryAccess) ByMultiRead(ctx context.Context, p []bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRead"
	l, e := a.fromQuery(ctx, qn, "read in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupRepositoryAccess) ByLikeRead(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeRead"
	l, e := a.fromQuery(ctx, qn, "read ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with matching Write
func (a *DBGroupRepositoryAccess) ByWrite(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with multiple matching Write
func (a *DBGroupRepositoryAccess) ByMultiWrite(ctx context.Context, p []bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupRepositoryAccess) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeWrite"
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
func (a *DBGroupRepositoryAccess) get_ID(p *savepb.GroupRepositoryAccess) uint64 {
	return uint64(p.ID)
}

// getter for field "RepoID" (RepoID) [uint64]
func (a *DBGroupRepositoryAccess) get_RepoID(p *savepb.GroupRepositoryAccess) uint64 {
	return uint64(p.RepoID)
}

// getter for field "GroupID" (GroupID) [string]
func (a *DBGroupRepositoryAccess) get_GroupID(p *savepb.GroupRepositoryAccess) string {
	return string(p.GroupID)
}

// getter for field "Read" (Read) [bool]
func (a *DBGroupRepositoryAccess) get_Read(p *savepb.GroupRepositoryAccess) bool {
	return bool(p.Read)
}

// getter for field "Write" (Write) [bool]
func (a *DBGroupRepositoryAccess) get_Write(p *savepb.GroupRepositoryAccess) bool {
	return bool(p.Write)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGroupRepositoryAccess) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.GroupRepositoryAccess, error) {
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

func (a *DBGroupRepositoryAccess) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GroupRepositoryAccess, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBGroupRepositoryAccess) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.GroupRepositoryAccess, error) {
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
func (a *DBGroupRepositoryAccess) get_col_from_proto(p *savepb.GroupRepositoryAccess, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repoid" {
		return a.get_RepoID(p)
	} else if colname == "groupid" {
		return a.get_GroupID(p)
	} else if colname == "read" {
		return a.get_Read(p)
	} else if colname == "write" {
		return a.get_Write(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBGroupRepositoryAccess) Tablename() string {
	return a.SQLTablename
}

func (a *DBGroupRepositoryAccess) SelectCols() string {
	return "id,repoid, groupid, read, write"
}
func (a *DBGroupRepositoryAccess) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repoid, " + a.SQLTablename + ".groupid, " + a.SQLTablename + ".read, " + a.SQLTablename + ".write"
}

func (a *DBGroupRepositoryAccess) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GroupRepositoryAccess, error) {
	var res []*savepb.GroupRepositoryAccess
	for rows.Next() {
		// SCANNER:
		foo := &savepb.GroupRepositoryAccess{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepoID
		scanTarget_2 := &foo.GroupID
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
func (a *DBGroupRepositoryAccess) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,groupid text not null ,read boolean not null ,write boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,groupid text not null ,read boolean not null ,write boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS groupid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS read boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS write boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repoid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS groupid text not null  default '';`,
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
func (a *DBGroupRepositoryAccess) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

