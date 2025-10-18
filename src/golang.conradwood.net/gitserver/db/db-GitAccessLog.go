package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBGitAccessLog
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence gitaccesslog_seq;

Main Table:

 CREATE TABLE gitaccesslog (id integer primary key default nextval('gitaccesslog_seq'),write boolean not null  ,userid text not null  ,r_timestamp integer not null  ,sourcerepository bigint not null  references sourcerepository (id) on delete cascade  );

Alter statements:
ALTER TABLE gitaccesslog ADD COLUMN IF NOT EXISTS write boolean not null default false;
ALTER TABLE gitaccesslog ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE gitaccesslog ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;
ALTER TABLE gitaccesslog ADD COLUMN IF NOT EXISTS sourcerepository bigint not null references sourcerepository (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE gitaccesslog_archive (id integer unique not null,write boolean not null,userid text not null,r_timestamp integer not null,sourcerepository bigint not null);
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
	default_def_DBGitAccessLog *DBGitAccessLog
)

type DBGitAccessLog struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBGitAccessLog()
	})
}

func DefaultDBGitAccessLog() *DBGitAccessLog {
	if default_def_DBGitAccessLog != nil {
		return default_def_DBGitAccessLog
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGitAccessLog(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGitAccessLog = res
	return res
}
func NewDBGitAccessLog(db *sql.DB) *DBGitAccessLog {
	foo := DBGitAccessLog{DB: db}
	foo.SQLTablename = "gitaccesslog"
	foo.SQLArchivetablename = "gitaccesslog_archive"
	return &foo
}

func (a *DBGitAccessLog) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBGitAccessLog) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBGitAccessLog) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBGitAccessLog) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGitAccessLog", "insert into "+a.SQLArchivetablename+" (id,write, userid, r_timestamp, sourcerepository) values ($1,$2, $3, $4, $5) ", p.ID, p.Write, p.UserID, p.Timestamp, p.SourceRepository.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBGitAccessLog) buildSaveMap(ctx context.Context, p *savepb.GitAccessLog) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["write"] = a.get_col_from_proto(p, "write")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["r_timestamp"] = a.get_col_from_proto(p, "r_timestamp")
	res["sourcerepository"] = a.get_col_from_proto(p, "sourcerepository")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBGitAccessLog) Save(ctx context.Context, p *savepb.GitAccessLog) (uint64, error) {
	qn := "save_DBGitAccessLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBGitAccessLog) SaveWithID(ctx context.Context, p *savepb.GitAccessLog) error {
	qn := "insert_DBGitAccessLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBGitAccessLog) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.GitAccessLog) (uint64, error) {
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
func (a *DBGitAccessLog) SaveOrUpdate(ctx context.Context, p *savepb.GitAccessLog) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBGitAccessLog) Update(ctx context.Context, p *savepb.GitAccessLog) error {
	qn := "DBGitAccessLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set write=$1, userid=$2, r_timestamp=$3, sourcerepository=$4 where id = $5", a.get_Write(p), a.get_UserID(p), a.get_Timestamp(p), a.get_SourceRepository_ID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGitAccessLog) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGitAccessLog_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGitAccessLog) ByID(ctx context.Context, p uint64) (*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No GitAccessLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GitAccessLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGitAccessLog) TryByID(ctx context.Context, p uint64) (*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) GitAccessLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBGitAccessLog) ByIDs(ctx context.Context, p []uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBGitAccessLog) All(ctx context.Context) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBGitAccessLog" rows with matching Write
func (a *DBGitAccessLog) ByWrite(ctx context.Context, p bool) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with multiple matching Write
func (a *DBGitAccessLog) ByMultiWrite(ctx context.Context, p []bool) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByWrite"
	l, e := a.fromQuery(ctx, qn, "write in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeWrite"
	l, e := a.fromQuery(ctx, qn, "write ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with matching UserID
func (a *DBGitAccessLog) ByUserID(ctx context.Context, p string) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with multiple matching UserID
func (a *DBGitAccessLog) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeUserID(ctx context.Context, p string) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with matching Timestamp
func (a *DBGitAccessLog) ByTimestamp(ctx context.Context, p uint32) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with multiple matching Timestamp
func (a *DBGitAccessLog) ByMultiTimestamp(ctx context.Context, p []uint32) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeTimestamp(ctx context.Context, p uint32) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with matching SourceRepository
func (a *DBGitAccessLog) BySourceRepository(ctx context.Context, p uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_BySourceRepository"
	l, e := a.fromQuery(ctx, qn, "sourcerepository = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepository: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with multiple matching SourceRepository
func (a *DBGitAccessLog) ByMultiSourceRepository(ctx context.Context, p []uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_BySourceRepository"
	l, e := a.fromQuery(ctx, qn, "sourcerepository in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepository: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeSourceRepository(ctx context.Context, p uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeSourceRepository"
	l, e := a.fromQuery(ctx, qn, "sourcerepository ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceRepository: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBGitAccessLog) get_ID(p *savepb.GitAccessLog) uint64 {
	return uint64(p.ID)
}

// getter for field "Write" (Write) [bool]
func (a *DBGitAccessLog) get_Write(p *savepb.GitAccessLog) bool {
	return bool(p.Write)
}

// getter for field "UserID" (UserID) [string]
func (a *DBGitAccessLog) get_UserID(p *savepb.GitAccessLog) string {
	return string(p.UserID)
}

// getter for field "Timestamp" (Timestamp) [uint32]
func (a *DBGitAccessLog) get_Timestamp(p *savepb.GitAccessLog) uint32 {
	return uint32(p.Timestamp)
}

// getter for reference "SourceRepository"
func (a *DBGitAccessLog) get_SourceRepository_ID(p *savepb.GitAccessLog) uint64 {
	if p.SourceRepository == nil {
		panic("field SourceRepository must not be nil")
	}
	return p.SourceRepository.ID
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGitAccessLog) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.GitAccessLog, error) {
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

func (a *DBGitAccessLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GitAccessLog, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBGitAccessLog) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.GitAccessLog, error) {
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
func (a *DBGitAccessLog) get_col_from_proto(p *savepb.GitAccessLog, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "write" {
		return a.get_Write(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "r_timestamp" {
		return a.get_Timestamp(p)
	} else if colname == "sourcerepository" {
		return a.get_SourceRepository_ID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBGitAccessLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBGitAccessLog) SelectCols() string {
	return "id,write, userid, r_timestamp, sourcerepository"
}
func (a *DBGitAccessLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".write, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".r_timestamp, " + a.SQLTablename + ".sourcerepository"
}

func (a *DBGitAccessLog) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GitAccessLog, error) {
	var res []*savepb.GitAccessLog
	for rows.Next() {
		// SCANNER:
		foo := &savepb.GitAccessLog{}
		// create the non-nullable pointers
		foo.SourceRepository = &savepb.SourceRepository{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Write
		scanTarget_2 := &foo.UserID
		scanTarget_3 := &foo.Timestamp
		scanTarget_4 := &foo.SourceRepository.ID
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
func (a *DBGitAccessLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),write boolean not null ,userid text not null ,r_timestamp integer not null ,sourcerepository bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),write boolean not null ,userid text not null ,r_timestamp integer not null ,sourcerepository bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS write boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS sourcerepository bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS write boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_timestamp integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS sourcerepository bigint not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_gitaccesslog_sourcerepository_sourcerepositoryid FOREIGN KEY (sourcerepository) references sourcerepository (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBGitAccessLog) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

