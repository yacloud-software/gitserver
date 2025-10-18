package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBCreateRepoLog
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence createrepolog_seq;

Main Table:

 CREATE TABLE createrepolog (id integer primary key default nextval('createrepolog_seq'),repositoryid bigint not null  ,userid text not null  ,context text not null  ,action integer not null  ,success boolean not null  ,errormessage text not null  ,started integer not null  ,finished integer not null  ,associationtoken text not null  );

Alter statements:
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS context text not null default '';
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS action integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS success boolean not null default false;
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS errormessage text not null default '';
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS started integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS finished integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN IF NOT EXISTS associationtoken text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE createrepolog_archive (id integer unique not null,repositoryid bigint not null,userid text not null,context text not null,action integer not null,success boolean not null,errormessage text not null,started integer not null,finished integer not null,associationtoken text not null);
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
	default_def_DBCreateRepoLog *DBCreateRepoLog
)

type DBCreateRepoLog struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBCreateRepoLog()
	})
}

func DefaultDBCreateRepoLog() *DBCreateRepoLog {
	if default_def_DBCreateRepoLog != nil {
		return default_def_DBCreateRepoLog
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBCreateRepoLog(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBCreateRepoLog = res
	return res
}
func NewDBCreateRepoLog(db *sql.DB) *DBCreateRepoLog {
	foo := DBCreateRepoLog{DB: db}
	foo.SQLTablename = "createrepolog"
	foo.SQLArchivetablename = "createrepolog_archive"
	return &foo
}

func (a *DBCreateRepoLog) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBCreateRepoLog) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBCreateRepoLog) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBCreateRepoLog) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBCreateRepoLog", "insert into "+a.SQLArchivetablename+" (id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10) ", p.ID, p.RepositoryID, p.UserID, p.Context, p.Action, p.Success, p.ErrorMessage, p.Started, p.Finished, p.AssociationToken)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBCreateRepoLog) buildSaveMap(ctx context.Context, p *savepb.CreateRepoLog) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["repositoryid"] = a.get_col_from_proto(p, "repositoryid")
	res["userid"] = a.get_col_from_proto(p, "userid")
	res["context"] = a.get_col_from_proto(p, "context")
	res["action"] = a.get_col_from_proto(p, "action")
	res["success"] = a.get_col_from_proto(p, "success")
	res["errormessage"] = a.get_col_from_proto(p, "errormessage")
	res["started"] = a.get_col_from_proto(p, "started")
	res["finished"] = a.get_col_from_proto(p, "finished")
	res["associationtoken"] = a.get_col_from_proto(p, "associationtoken")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBCreateRepoLog) Save(ctx context.Context, p *savepb.CreateRepoLog) (uint64, error) {
	qn := "save_DBCreateRepoLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBCreateRepoLog) SaveWithID(ctx context.Context, p *savepb.CreateRepoLog) error {
	qn := "insert_DBCreateRepoLog"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBCreateRepoLog) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.CreateRepoLog) (uint64, error) {
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
func (a *DBCreateRepoLog) SaveOrUpdate(ctx context.Context, p *savepb.CreateRepoLog) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBCreateRepoLog) Update(ctx context.Context, p *savepb.CreateRepoLog) error {
	qn := "DBCreateRepoLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, userid=$2, context=$3, action=$4, success=$5, errormessage=$6, started=$7, finished=$8, associationtoken=$9 where id = $10", a.get_RepositoryID(p), a.get_UserID(p), a.get_Context(p), a.get_Action(p), a.get_Success(p), a.get_ErrorMessage(p), a.get_Started(p), a.get_Finished(p), a.get_AssociationToken(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBCreateRepoLog) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBCreateRepoLog_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBCreateRepoLog) ByID(ctx context.Context, p uint64) (*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No CreateRepoLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) CreateRepoLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBCreateRepoLog) TryByID(ctx context.Context, p uint64) (*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) CreateRepoLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBCreateRepoLog) ByIDs(ctx context.Context, p []uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBCreateRepoLog) All(ctx context.Context) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBCreateRepoLog" rows with matching RepositoryID
func (a *DBCreateRepoLog) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching RepositoryID
func (a *DBCreateRepoLog) ByMultiRepositoryID(ctx context.Context, p []uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeRepositoryID"
	l, e := a.fromQuery(ctx, qn, "repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching UserID
func (a *DBCreateRepoLog) ByUserID(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching UserID
func (a *DBCreateRepoLog) ByMultiUserID(ctx context.Context, p []string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByUserID"
	l, e := a.fromQuery(ctx, qn, "userid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeUserID(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeUserID"
	l, e := a.fromQuery(ctx, qn, "userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Context
func (a *DBCreateRepoLog) ByContext(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByContext"
	l, e := a.fromQuery(ctx, qn, "context = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching Context
func (a *DBCreateRepoLog) ByMultiContext(ctx context.Context, p []string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByContext"
	l, e := a.fromQuery(ctx, qn, "context in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeContext(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeContext"
	l, e := a.fromQuery(ctx, qn, "context ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Action
func (a *DBCreateRepoLog) ByAction(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAction"
	l, e := a.fromQuery(ctx, qn, "action = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAction: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching Action
func (a *DBCreateRepoLog) ByMultiAction(ctx context.Context, p []uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAction"
	l, e := a.fromQuery(ctx, qn, "action in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAction: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeAction(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeAction"
	l, e := a.fromQuery(ctx, qn, "action ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAction: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Success
func (a *DBCreateRepoLog) BySuccess(ctx context.Context, p bool) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching Success
func (a *DBCreateRepoLog) ByMultiSuccess(ctx context.Context, p []bool) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_BySuccess"
	l, e := a.fromQuery(ctx, qn, "success in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeSuccess"
	l, e := a.fromQuery(ctx, qn, "success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching ErrorMessage
func (a *DBCreateRepoLog) ByErrorMessage(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByErrorMessage"
	l, e := a.fromQuery(ctx, qn, "errormessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByErrorMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching ErrorMessage
func (a *DBCreateRepoLog) ByMultiErrorMessage(ctx context.Context, p []string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByErrorMessage"
	l, e := a.fromQuery(ctx, qn, "errormessage in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByErrorMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeErrorMessage(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeErrorMessage"
	l, e := a.fromQuery(ctx, qn, "errormessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByErrorMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Started
func (a *DBCreateRepoLog) ByStarted(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByStarted"
	l, e := a.fromQuery(ctx, qn, "started = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching Started
func (a *DBCreateRepoLog) ByMultiStarted(ctx context.Context, p []uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByStarted"
	l, e := a.fromQuery(ctx, qn, "started in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeStarted(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeStarted"
	l, e := a.fromQuery(ctx, qn, "started ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Finished
func (a *DBCreateRepoLog) ByFinished(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByFinished"
	l, e := a.fromQuery(ctx, qn, "finished = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching Finished
func (a *DBCreateRepoLog) ByMultiFinished(ctx context.Context, p []uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByFinished"
	l, e := a.fromQuery(ctx, qn, "finished in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeFinished(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeFinished"
	l, e := a.fromQuery(ctx, qn, "finished ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching AssociationToken
func (a *DBCreateRepoLog) ByAssociationToken(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with multiple matching AssociationToken
func (a *DBCreateRepoLog) ByMultiAssociationToken(ctx context.Context, p []string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeAssociationToken(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeAssociationToken"
	l, e := a.fromQuery(ctx, qn, "associationtoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBCreateRepoLog) get_ID(p *savepb.CreateRepoLog) uint64 {
	return uint64(p.ID)
}

// getter for field "RepositoryID" (RepositoryID) [uint64]
func (a *DBCreateRepoLog) get_RepositoryID(p *savepb.CreateRepoLog) uint64 {
	return uint64(p.RepositoryID)
}

// getter for field "UserID" (UserID) [string]
func (a *DBCreateRepoLog) get_UserID(p *savepb.CreateRepoLog) string {
	return string(p.UserID)
}

// getter for field "Context" (Context) [string]
func (a *DBCreateRepoLog) get_Context(p *savepb.CreateRepoLog) string {
	return string(p.Context)
}

// getter for field "Action" (Action) [uint32]
func (a *DBCreateRepoLog) get_Action(p *savepb.CreateRepoLog) uint32 {
	return uint32(p.Action)
}

// getter for field "Success" (Success) [bool]
func (a *DBCreateRepoLog) get_Success(p *savepb.CreateRepoLog) bool {
	return bool(p.Success)
}

// getter for field "ErrorMessage" (ErrorMessage) [string]
func (a *DBCreateRepoLog) get_ErrorMessage(p *savepb.CreateRepoLog) string {
	return string(p.ErrorMessage)
}

// getter for field "Started" (Started) [uint32]
func (a *DBCreateRepoLog) get_Started(p *savepb.CreateRepoLog) uint32 {
	return uint32(p.Started)
}

// getter for field "Finished" (Finished) [uint32]
func (a *DBCreateRepoLog) get_Finished(p *savepb.CreateRepoLog) uint32 {
	return uint32(p.Finished)
}

// getter for field "AssociationToken" (AssociationToken) [string]
func (a *DBCreateRepoLog) get_AssociationToken(p *savepb.CreateRepoLog) string {
	return string(p.AssociationToken)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBCreateRepoLog) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.CreateRepoLog, error) {
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

func (a *DBCreateRepoLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.CreateRepoLog, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBCreateRepoLog) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.CreateRepoLog, error) {
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
func (a *DBCreateRepoLog) get_col_from_proto(p *savepb.CreateRepoLog, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "repositoryid" {
		return a.get_RepositoryID(p)
	} else if colname == "userid" {
		return a.get_UserID(p)
	} else if colname == "context" {
		return a.get_Context(p)
	} else if colname == "action" {
		return a.get_Action(p)
	} else if colname == "success" {
		return a.get_Success(p)
	} else if colname == "errormessage" {
		return a.get_ErrorMessage(p)
	} else if colname == "started" {
		return a.get_Started(p)
	} else if colname == "finished" {
		return a.get_Finished(p)
	} else if colname == "associationtoken" {
		return a.get_AssociationToken(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBCreateRepoLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBCreateRepoLog) SelectCols() string {
	return "id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken"
}
func (a *DBCreateRepoLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".context, " + a.SQLTablename + ".action, " + a.SQLTablename + ".success, " + a.SQLTablename + ".errormessage, " + a.SQLTablename + ".started, " + a.SQLTablename + ".finished, " + a.SQLTablename + ".associationtoken"
}

func (a *DBCreateRepoLog) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.CreateRepoLog, error) {
	var res []*savepb.CreateRepoLog
	for rows.Next() {
		// SCANNER:
		foo := &savepb.CreateRepoLog{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepositoryID
		scanTarget_2 := &foo.UserID
		scanTarget_3 := &foo.Context
		scanTarget_4 := &foo.Action
		scanTarget_5 := &foo.Success
		scanTarget_6 := &foo.ErrorMessage
		scanTarget_7 := &foo.Started
		scanTarget_8 := &foo.Finished
		scanTarget_9 := &foo.AssociationToken
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9)
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
func (a *DBCreateRepoLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,userid text not null ,context text not null ,action integer not null ,success boolean not null ,errormessage text not null ,started integer not null ,finished integer not null ,associationtoken text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,userid text not null ,context text not null ,action integer not null ,success boolean not null ,errormessage text not null ,started integer not null ,finished integer not null ,associationtoken text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS context text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS action integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS errormessage text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS started integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS finished integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS associationtoken text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repositoryid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS context text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS action integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS success boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS errormessage text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS started integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS finished integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS associationtoken text not null  default '';`,
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
func (a *DBCreateRepoLog) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

