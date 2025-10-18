package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBBuild
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence build_seq;

Main Table:

 CREATE TABLE build (id integer primary key default nextval('build_seq'),repositoryid bigint not null  ,commithash text not null  ,branch text not null  ,logmessage text not null  ,userid text not null  ,r_timestamp integer not null  ,success boolean not null  );

Alter statements:
ALTER TABLE build ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE build ADD COLUMN IF NOT EXISTS commithash text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS branch text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS logmessage text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;
ALTER TABLE build ADD COLUMN IF NOT EXISTS success boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE build_archive (id integer unique not null,repositoryid bigint not null,commithash text not null,branch text not null,logmessage text not null,userid text not null,r_timestamp integer not null,success boolean not null);
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
	default_def_DBBuild *DBBuild
)

type DBBuild struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBBuild()
	})
}

func DefaultDBBuild() *DBBuild {
	if default_def_DBBuild != nil {
		return default_def_DBBuild
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBBuild(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBBuild = res
	return res
}
func NewDBBuild(db *sql.DB) *DBBuild {
	foo := DBBuild{DB: db}
	foo.SQLTablename = "build"
	foo.SQLArchivetablename = "build_archive"
	return &foo
}

func (a *DBBuild) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBBuild) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBBuild) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBBuild) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBBuild", "insert into "+a.SQLArchivetablename+" (id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.RepositoryID, p.CommitHash, p.Branch, p.LogMessage, p.UserID, p.Timestamp, p.Success)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBBuild) buildSaveMap(ctx context.Context, p *savepb.Build) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["commithash"] = a.get_col_from_proto(p, "commithash")
	res["branch"] = a.get_col_from_proto(p, "branch")
	res["logmessage"] = a.get_col_from_proto(p, "logmessage")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["r_timestamp"] = a.get_col_from_proto(p, "r_timestamp")
	res["success"] = a.get_col_from_proto(p, "success")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBBuild) Save(ctx context.Context, p *savepb.Build) (uint64, error) {
	qn := "save_DBBuild"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBBuild) SaveWithID(ctx context.Context, p *savepb.Build) error {
	qn := "insert_DBBuild"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBBuild) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.Build) (uint64, error) {
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
func (a *DBBuild) SaveOrUpdate(ctx context.Context, p *savepb.Build) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBBuild) Update(ctx context.Context, p *savepb.Build) error {
	qn := "DBBuild_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, commithash=$2, branch=$3, logmessage=$4, userid=$5, r_timestamp=$6, success=$7 where id = $8", a.get_RepositoryID(p), a.get_CommitHash(p), a.get_Branch(p), a.get_LogMessage(p), a.get_UserID(p), a.get_Timestamp(p), a.get_Success(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBBuild) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBBuild_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBBuild) ByID(ctx context.Context, p uint64) (*savepb.Build, error) {
	qn := "DBBuild_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No Build with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Build with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBBuild) TryByID(ctx context.Context, p uint64) (*savepb.Build, error) {
	qn := "DBBuild_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Build with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBBuild) ByIDs(ctx context.Context, p []uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBBuild) All(ctx context.Context) ([]*savepb.Build, error) {
	qn := "DBBuild_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBBuild" rows with matching RepositoryID
func (a *DBBuild) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching RepositoryID
func (a *DBBuild) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching CommitHash
func (a *DBBuild) ByCommitHash(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByCommitHash"
	l, e := a.fromQuery(ctx, qn, "commithash = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCommitHash: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching CommitHash
func (a *DBBuild) ByMultiCommitHash(ctx context.Context, p []string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByCommitHash"
	l, e := a.fromQuery(ctx, qn, "commithash in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCommitHash: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeCommitHash(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeCommitHash"
	l, e := a.fromQuery(ctx, qn, "commithash ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCommitHash: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching Branch
func (a *DBBuild) ByBranch(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByBranch"
	l, e := a.fromQuery(ctx, qn, "branch = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching Branch
func (a *DBBuild) ByMultiBranch(ctx context.Context, p []string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByBranch"
	l, e := a.fromQuery(ctx, qn, "branch in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeBranch(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeBranch"
	l, e := a.fromQuery(ctx, qn, "branch ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching LogMessage
func (a *DBBuild) ByLogMessage(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching LogMessage
func (a *DBBuild) ByMultiLogMessage(ctx context.Context, p []string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeLogMessage(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeLogMessage"
	l, e := a.fromQuery(ctx, qn, "logmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching UserID
func (a *DBBuild) ByUserID(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching UserID
func (a *DBBuild) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeUserID(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching Timestamp
func (a *DBBuild) ByTimestamp(ctx context.Context, p uint32) ([]*savepb.Build, error) {
	qn := "DBBuild_ByTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching Timestamp
func (a *DBBuild) ByMultiTimestamp(ctx context.Context, p []uint32) ([]*savepb.Build, error) {
	qn := "DBBuild_ByTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeTimestamp(ctx context.Context, p uint32) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeTimestamp"
	l, e := a.fromQuery(ctx, qn, "r_timestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching Success
func (a *DBBuild) BySuccess(ctx context.Context, p bool) ([]*savepb.Build, error) {
	qn := "DBBuild_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with multiple matching Success
func (a *DBBuild) ByMultiSuccess(ctx context.Context, p []bool) ([]*savepb.Build, error) {
	qn := "DBBuild_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeSuccess"
	l, e := a.fromQuery(ctx, qn, "success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBBuild) get_ID(p *savepb.Build) uint64 {
	return uint64(p.ID)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBBuild) get_RepositoryID(p *savepb.Build) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "CommitHash" (CommitHash) [string]
func (a *DBBuild) get_CommitHash(p *savepb.Build) string {
	return string(p.CommitHash)
}

// getter for field "Branch" (Branch) [string]
func (a *DBBuild) get_Branch(p *savepb.Build) string {
	return string(p.Branch)
}

// getter for field "LogMessage" (LogMessage) [string]
func (a *DBBuild) get_LogMessage(p *savepb.Build) string {
	return string(p.LogMessage)
}

// getter for field "UserID" (UserID) [string]
func (a *DBBuild) get_UserID(p *savepb.Build) string {
	return string(p.UserID)
}

// getter for field "Timestamp" (Timestamp) [uint32]
func (a *DBBuild) get_Timestamp(p *savepb.Build) uint32 {
	return uint32(p.Timestamp)
}

// getter for field "Success" (Success) [bool]
func (a *DBBuild) get_Success(p *savepb.Build) bool {
	return bool(p.Success)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBBuild) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.Build, error) {
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

func (a *DBBuild) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Build, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBBuild) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.Build, error) {
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
func (a *DBBuild) get_col_from_proto(p *savepb.Build, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "commithash" {
		return a.get_CommitHash(p)
	} else if colname == "branch" {
		return a.get_Branch(p)
	} else if colname == "logmessage" {
		return a.get_LogMessage(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "r_timestamp" {
		return a.get_Timestamp(p)
	} else if colname == "success" {
		return a.get_Success(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBBuild) Tablename() string {
	return a.SQLTablename
}

func (a *DBBuild) SelectCols() string {
	return "id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success"
}
func (a *DBBuild) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".commithash, " + a.SQLTablename + ".branch, " + a.SQLTablename + ".logmessage, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".r_timestamp, " + a.SQLTablename + ".success"
}

func (a *DBBuild) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Build, error) {
	var res []*savepb.Build
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Build{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepositoryID
		scanTarget_2 := &foo.CommitHash
		scanTarget_3 := &foo.Branch
		scanTarget_4 := &foo.LogMessage
		scanTarget_5 := &foo.UserID
		scanTarget_6 := &foo.Timestamp
		scanTarget_7 := &foo.Success
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7)
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
func (a *DBBuild) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,commithash text not null ,branch text not null ,logmessage text not null ,userid text not null ,r_timestamp integer not null ,success boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,commithash text not null ,branch text not null ,logmessage text not null ,userid text not null ,r_timestamp integer not null ,success boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS commithash text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS branch text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS logmessage text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS success boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS commithash text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS branch text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS logmessage text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS r_timestamp integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS success boolean not null  default false;`,
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
func (a *DBBuild) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

