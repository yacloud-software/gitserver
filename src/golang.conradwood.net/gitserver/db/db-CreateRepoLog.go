package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBCreateRepoLog
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
ALTER TABLE createrepolog ADD COLUMN repositoryid bigint not null default 0;
ALTER TABLE createrepolog ADD COLUMN userid text not null default '';
ALTER TABLE createrepolog ADD COLUMN context text not null default '';
ALTER TABLE createrepolog ADD COLUMN action integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN success boolean not null default false;
ALTER TABLE createrepolog ADD COLUMN errormessage text not null default '';
ALTER TABLE createrepolog ADD COLUMN started integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN finished integer not null default 0;
ALTER TABLE createrepolog ADD COLUMN associationtoken text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE createrepolog_archive (id integer unique not null,repositoryid bigint not null,userid text not null,context text not null,action integer not null,success boolean not null,errormessage text not null,started integer not null,finished integer not null,associationtoken text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBCreateRepoLog *DBCreateRepoLog
)

type DBCreateRepoLog struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBCreateRepoLog) Save(ctx context.Context, p *savepb.CreateRepoLog) (uint64, error) {
	qn := "DBCreateRepoLog_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id", p.RepositoryID, p.UserID, p.Context, p.Action, p.Success, p.ErrorMessage, p.Started, p.Finished, p.AssociationToken)
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBCreateRepoLog) SaveWithID(ctx context.Context, p *savepb.CreateRepoLog) error {
	qn := "insert_DBCreateRepoLog"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10) ", p.ID, p.RepositoryID, p.UserID, p.Context, p.Action, p.Success, p.ErrorMessage, p.Started, p.Finished, p.AssociationToken)
	return a.Error(ctx, qn, e)
}

func (a *DBCreateRepoLog) Update(ctx context.Context, p *savepb.CreateRepoLog) error {
	qn := "DBCreateRepoLog_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, userid=$2, context=$3, action=$4, success=$5, errormessage=$6, started=$7, finished=$8, associationtoken=$9 where id = $10", p.RepositoryID, p.UserID, p.Context, p.Action, p.Success, p.ErrorMessage, p.Started, p.Finished, p.AssociationToken, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No CreateRepoLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) CreateRepoLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBCreateRepoLog) All(ctx context.Context) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBCreateRepoLog" rows with matching RepositoryID
func (a *DBCreateRepoLog) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching UserID
func (a *DBCreateRepoLog) ByUserID(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeUserID(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Context
func (a *DBCreateRepoLog) ByContext(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByContext"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where context = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeContext(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeContext"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where context ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByContext: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Action
func (a *DBCreateRepoLog) ByAction(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAction"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where action = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAction: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAction: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeAction(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeAction"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where action ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAction: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAction: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Success
func (a *DBCreateRepoLog) BySuccess(ctx context.Context, p bool) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_BySuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeSuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching ErrorMessage
func (a *DBCreateRepoLog) ByErrorMessage(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByErrorMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where errormessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByErrorMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByErrorMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeErrorMessage(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeErrorMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where errormessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByErrorMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByErrorMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Started
func (a *DBCreateRepoLog) ByStarted(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByStarted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where started = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeStarted(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeStarted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where started ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByStarted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching Finished
func (a *DBCreateRepoLog) ByFinished(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByFinished"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where finished = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeFinished(ctx context.Context, p uint32) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeFinished"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where finished ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFinished: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBCreateRepoLog" rows with matching AssociationToken
func (a *DBCreateRepoLog) ByAssociationToken(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByAssociationToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where associationtoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBCreateRepoLog) ByLikeAssociationToken(ctx context.Context, p string) ([]*savepb.CreateRepoLog, error) {
	qn := "DBCreateRepoLog_ByLikeAssociationToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, userid, context, action, success, errormessage, started, finished, associationtoken from "+a.SQLTablename+" where associationtoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBCreateRepoLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.CreateRepoLog, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
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
		foo := savepb.CreateRepoLog{}
		err := rows.Scan(&foo.ID, &foo.RepositoryID, &foo.UserID, &foo.Context, &foo.Action, &foo.Success, &foo.ErrorMessage, &foo.Started, &foo.Finished, &foo.AssociationToken)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBCreateRepoLog) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null  ,userid text not null  ,context text not null  ,action integer not null  ,success boolean not null  ,errormessage text not null  ,started integer not null  ,finished integer not null  ,associationtoken text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null  ,userid text not null  ,context text not null  ,action integer not null  ,success boolean not null  ,errormessage text not null  ,started integer not null  ,finished integer not null  ,associationtoken text not null  );`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
