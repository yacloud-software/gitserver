package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBGitAccessLog
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBGitAccessLog *DBGitAccessLog
)

type DBGitAccessLog struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBGitAccessLog) Save(ctx context.Context, p *savepb.GitAccessLog) (uint64, error) {
	qn := "DBGitAccessLog_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (write, userid, r_timestamp, sourcerepository) values ($1, $2, $3, $4) returning id", a.get_Write(p), a.get_UserID(p), a.get_Timestamp(p), a.get_SourceRepository_ID(p))
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
func (a *DBGitAccessLog) SaveWithID(ctx context.Context, p *savepb.GitAccessLog) error {
	qn := "insert_DBGitAccessLog"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,write, userid, r_timestamp, sourcerepository) values ($1,$2, $3, $4, $5) ", p.ID, p.Write, p.UserID, p.Timestamp, p.SourceRepository.ID)
	return a.Error(ctx, qn, e)
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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No GitAccessLog with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GitAccessLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGitAccessLog) TryByID(ctx context.Context, p uint64) (*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GitAccessLog with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBGitAccessLog) All(ctx context.Context) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" order by id")
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

// get all "DBGitAccessLog" rows with matching Write
func (a *DBGitAccessLog) ByWrite(ctx context.Context, p bool) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where write = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where write ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with matching UserID
func (a *DBGitAccessLog) ByUserID(ctx context.Context, p string) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBGitAccessLog) ByLikeUserID(ctx context.Context, p string) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBGitAccessLog" rows with matching Timestamp
func (a *DBGitAccessLog) ByTimestamp(ctx context.Context, p uint32) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where r_timestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeTimestamp(ctx context.Context, p uint32) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where r_timestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitAccessLog" rows with matching SourceRepository
func (a *DBGitAccessLog) BySourceRepository(ctx context.Context, p uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_BySourceRepository"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where sourcerepository = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepository: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepository: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitAccessLog) ByLikeSourceRepository(ctx context.Context, p uint64) ([]*savepb.GitAccessLog, error) {
	qn := "DBGitAccessLog_ByLikeSourceRepository"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,write, userid, r_timestamp, sourcerepository from "+a.SQLTablename+" where sourcerepository ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepository: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceRepository: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBGitAccessLog) get_ID(p *savepb.GitAccessLog) uint64 {
	return p.ID
}

func (a *DBGitAccessLog) get_Write(p *savepb.GitAccessLog) bool {
	return p.Write
}

func (a *DBGitAccessLog) get_UserID(p *savepb.GitAccessLog) string {
	return p.UserID
}

func (a *DBGitAccessLog) get_Timestamp(p *savepb.GitAccessLog) uint32 {
	return p.Timestamp
}

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
func (a *DBGitAccessLog) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GitAccessLog, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBGitAccessLog) Tablename() string {
	return a.SQLTablename
}

func (a *DBGitAccessLog) SelectCols() string {
	return "id,write, userid, r_timestamp, sourcerepository"
}
func (a *DBGitAccessLog) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".write, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".r_timestamp, " + a.SQLTablename + ".sourcerepository"
}

func (a *DBGitAccessLog) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.GitAccessLog, error) {
	var res []*savepb.GitAccessLog
	for rows.Next() {
		foo := savepb.GitAccessLog{SourceRepository: &savepb.SourceRepository{}}
		err := rows.Scan(&foo.ID, &foo.Write, &foo.UserID, &foo.Timestamp, &foo.SourceRepository.ID)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

