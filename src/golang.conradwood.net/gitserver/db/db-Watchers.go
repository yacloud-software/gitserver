package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBWatchers
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBWatchers *DBWatchers
)

type DBWatchers struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBWatchers) Save(ctx context.Context, p *savepb.Watchers) (uint64, error) {
	qn := "DBWatchers_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, repositoryid, notifytype) values ($1, $2, $3) returning id", a.get_UserID(p), a.get_RepositoryID(p), a.get_Notifytype(p))
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
func (a *DBWatchers) SaveWithID(ctx context.Context, p *savepb.Watchers) error {
	qn := "insert_DBWatchers"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, repositoryid, notifytype) values ($1,$2, $3, $4) ", p.ID, p.UserID, p.RepositoryID, p.Notifytype)
	return a.Error(ctx, qn, e)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Watchers with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Watchers with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBWatchers) TryByID(ctx context.Context, p uint64) (*savepb.Watchers, error) {
	qn := "DBWatchers_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Watchers with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBWatchers) All(ctx context.Context) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" order by id")
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

// get all "DBWatchers" rows with matching UserID
func (a *DBWatchers) ByUserID(ctx context.Context, p string) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBWatchers) ByLikeUserID(ctx context.Context, p string) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBWatchers" rows with matching RepositoryID
func (a *DBWatchers) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where repositoryid = $1", p)
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
func (a *DBWatchers) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where repositoryid ilike $1", p)
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

// get all "DBWatchers" rows with matching Notifytype
func (a *DBWatchers) ByNotifytype(ctx context.Context, p uint32) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByNotifytype"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where notifytype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotifytype: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotifytype: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBWatchers) ByLikeNotifytype(ctx context.Context, p uint32) ([]*savepb.Watchers, error) {
	qn := "DBWatchers_ByLikeNotifytype"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, repositoryid, notifytype from "+a.SQLTablename+" where notifytype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotifytype: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByNotifytype: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBWatchers) get_ID(p *savepb.Watchers) uint64 {
	return p.ID
}

func (a *DBWatchers) get_UserID(p *savepb.Watchers) string {
	return p.UserID
}

func (a *DBWatchers) get_RepositoryID(p *savepb.Watchers) uint64 {
	return p.RepositoryID
}

func (a *DBWatchers) get_Notifytype(p *savepb.Watchers) uint32 {
	return p.Notifytype
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBWatchers) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Watchers, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBWatchers) Tablename() string {
	return a.SQLTablename
}

func (a *DBWatchers) SelectCols() string {
	return "id,userid, repositoryid, notifytype"
}
func (a *DBWatchers) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".notifytype"
}

func (a *DBWatchers) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.Watchers, error) {
	var res []*savepb.Watchers
	for rows.Next() {
		foo := savepb.Watchers{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.RepositoryID, &foo.Notifytype)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

